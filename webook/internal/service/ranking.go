package service

import (
	"awesomeProject/webook/internal/domain"
	"awesomeProject/webook/internal/repository"
	"context"
	"github.com/ecodeclub/ekit/queue"
	"github.com/ecodeclub/ekit/slice"
	"math"
	"time"
)

// 热榜的设计与实现
type RankingService interface {
	TopN(ctx context.Context) error
}

type BatchRankingService struct {
	artSvc    ArticleService
	intrSvc   InteractiveService
	repo      repository.RankingRepository
	batchSize int
	n         int
	//计算得分函数
	scoreFunc func(t time.Time, likeCnt int64) float64
}

func NewBatchRankingService(artSvc ArticleService, intrSvc InteractiveService) RankingService {
	return &BatchRankingService{
		artSvc:    artSvc,
		intrSvc:   intrSvc,
		batchSize: 100,
		n:         100,
		scoreFunc: func(t time.Time, likeCnt int64) float64 {
			sec := time.Since(t).Seconds()
			return float64(likeCnt-1) / math.Pow(float64(sec+2), 1.5)
		},
	}
}

// 准备分批
func (svc *BatchRankingService) TopN(ctx context.Context) error {
	//调用下面的热榜模型实现更新并且放入redis中
	arts, err := svc.topN(ctx)

	if err == nil {
		return err
	}
	//在这里放到缓存中存起来
	return svc.repo.ReplaceTopN(ctx, arts)

}

// 先测试这个这个是纯粹的算法
func (svc *BatchRankingService) topN(ctx context.Context) ([]domain.Article, error) {
	//先拿一批数据出来
	offset := 0
	//我只取七天内的数据
	now := time.Now()
	//每篇文章对应的得分
	type Score struct {
		art   domain.Article
		score float64
	}
	//把优先级队列给实现使小根堆每次计算出新的一篇文章的得分后会尝试吧文章插到队列中  队列的容量设置是n
	topN := queue.NewConcurrentPriorityQueue[Score](svc.n,
		func(src Score, dst Score) int {
			//表示优先级高于
			if src.score > dst.score {
				return 1
			} else if src.score == dst.score {
				return 0
			} else {
				return -1
			}
		})

	//配合分批取数据避免加载一次性的所有数据

	//退出循环的条件就是数据量不足批次大小，超出七天的范畴--------后面结合异步定时器进行使用可以保证热榜的实时性
	for {
		//这里拿了一批的数据出来 ,,,每一批都是数据库中的文章我并不知道他的点赞数
		arts, err := svc.artSvc.ListPub(ctx, now, offset, svc.batchSize)
		if err != nil {
			return nil, err
		}
		//这里的【】表示要把 article类型转化成int64类型
		//也就是说把查询出来的批量文章的id给提取出来
		ids := slice.Map[domain.Article, int64](arts,
			func(idx int, src domain.Article) int64 {
				return src.Id
			})

		//这里要把拿出来的id数据去找对应批量文章的点赞数据，这个才是点赞记录高低的查询
		//取出来的数据是一个集合的形式
		intrs, err := svc.intrSvc.GetByIds(ctx, "article", ids)
		if err != nil {
			return nil, err
		}

		//对数据库进行查询的操作然后进行查询
		//合并计算得分
		//排序
		for _, art := range arts {
			intr := intrs[art.Id]
			//计算该篇文章的得分
			score := svc.scoreFunc(art.Utime, intr.LikeCnt)
			//我要考虑我这个score是否是在前一百名
			//小顶堆让他按照顺序出列就是有序的

			//尝试将每篇文章的Score都插入到优先级队列-----点赞榜单只是一个简单的计算得分的方式
			//你要返回单的还是一个文章所有的信息，可能不包括点赞
			err = topN.Enqueue(Score{
				art:   art,
				score: score,
			})

			//处理队列中已经满的情况
			if err == queue.ErrOutOfCapacity {
				//出列
				val, _ := topN.Dequeue()
				//比较新文章和文章最低得分
				if val.score < score {
					//放入新文章、
					err = topN.Enqueue(Score{
						art:   art,
						score: score,
					})
				} else {
					//不然就是把文章重新返回去
					_ = topN.Enqueue(val)
				}
			}

		}
		if len(arts) < svc.batchSize || now.Sub(arts[len(arts)-1].Utime).Hours() > 7*24 {
			//我这一批都没有取够我当然可以没有下一批
			//取到了七天之前的数据就可以中断了
			break
		}
		//跳过之前已经处理好了的文章记录
		offset = offset + len(arts)

	}
	//最终将优先级队列中的文章按照分数从高到低去除100名文章
	res := make([]domain.Article, svc.n)
	//得出结果  采用倒叙存储
	for i := svc.n - 1; i >= 0; i-- {
		val, err := topN.Dequeue()
		if err != nil {
			break
		}
		res[i] = val.art
	}
	return res, nil
}

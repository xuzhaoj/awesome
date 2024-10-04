package service

import (
	"awesomeProject/webook/internal/domain"
	"context"
	"github.com/ecodeclub/ekit/queue"
	"github.com/ecodeclub/ekit/slice"
	"log"
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
	batchSize int
	n         int
	scoreFunc func(t time.Time, likeCnt int64) float64
}

func NewBatchRankingService(artSvc ArticleService, intrSvc InteractiveService) *BatchRankingService {
	return &BatchRankingService{
		artSvc:    artSvc,
		intrSvc:   intrSvc,
		batchSize: 100,
		n:         100,
		scoreFunc: func(t time.Time, likeCnt int64) float64 {
			return float64(likeCnt-1) / math.Pow(float64(likeCnt+2), 1.5)
		},
	}
}

// 准备分批
func (svc *BatchRankingService) TopN(ctx context.Context) error {

	arts, err := svc.topN(ctx)
	if err == nil {
		//执行逻辑看看要存储在redis中还是存在哪里
		log.Println(arts)
	}
	return nil
}

// 先测试这个这个是纯粹的算法
func (svc *BatchRankingService) topN(ctx context.Context) ([]domain.Article, error) {
	//先拿一批数据出来
	offset := 0
	//我只取七天内的数据
	now := time.Now()

	type Score struct {
		art   domain.Article
		score float64
	}
	//把优先级队列给实现使用大跟堆
	topN := queue.NewConcurrentPriorityQueue[Score](svc.n,
		func(src Score, dst Score) int {
			if src.score > dst.score {
				return 1
			} else if src.score == dst.score {
				return 0
			} else {
				return -1
			}
		})

	for {
		//这里拿了一批的数据出来 ,,,每一批开始的地方都会变只是说批次大小不一样而已
		arts, err := svc.artSvc.ListPub(ctx, now, offset, svc.batchSize)
		if err != nil {
			return nil, err
		}
		ids := slice.Map[domain.Article, int64](arts,
			func(idx int, src domain.Article) int64 {
				return src.Id
			})
		//这里要把拿出来的id数据去找对应的点赞数据
		intrs, err := svc.intrSvc.GetByIds(ctx, "article", ids)
		if err != nil {
			return nil, err
		}
		//合并计算得分
		//排序
		for _, art := range arts {
			intr := intrs[art.Id]

			score := svc.scoreFunc(art.Utime, intr.LikeCnt)
			//我要考虑我这个score是否是在前一百名
			//小顶堆让他按照顺序出列就是有序的

			err = topN.Enqueue(Score{
				art:   art,
				score: score,
			})

			//这种写法要求topN已经满了
			if err == queue.ErrOutOfCapacity {
				val, _ := topN.Dequeue()
				if val.score < score {
					err = topN.Enqueue(Score{
						art:   art,
						score: score,
					})
				} else {
					_ = topN.Enqueue(val)
				}
			}

		}
		if len(arts) < svc.batchSize {
			//我这一批都没有取够我当然可以没有下一批
			break
		}
		offset = offset + len(arts)

	}
	res := make([]domain.Article, svc.n)
	//得出结果
	for i := svc.n - 1; i >= 0; i-- {
		val, err := topN.Dequeue()
		if err != nil {
			break
		}
		res[i] = val.art
	}
	return res, nil
}

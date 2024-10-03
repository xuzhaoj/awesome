package service

import (
	"awesomeProject/webook/internal/domain"
	events "awesomeProject/webook/internal/events/article"
	"awesomeProject/webook/internal/repository/article"
	"awesomeProject/webook/pkg/logger"
	"context"
	"time"
)

type ArticleService interface {
	//就是一个创建的服务
	Save(ctx context.Context, art domain.Article) (int64, error)
	//通过状态码实现对文章状态的更新，同步线上库和线下库，也就是实现了保存到草稿箱，然后发布
	Publish(ctx context.Context, art domain.Article) (int64, error)
	PublishV1(ctx context.Context, art domain.Article) (int64, error)
	Withdraw(ctx context.Context, art domain.Article) error
	//查询文章数组接口
	List(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error)
	//根据文章帖子id查询文章信息
	GetById(ctx context.Context, id int64) (domain.Article, error)
	GetPublishedById(ctx context.Context, id int64, uid int64) (domain.Article, error)
}

type articleService struct {
	//因为要区分制作库和线上库，所以把一统的进行拆分
	repo article.ArticleRepository

	//V1
	author article.ArticleAuthorRepository
	reader article.ArticleReaderRepository
	l      logger.LoggerV1
	//来一个生产者
	producer events.Producer
}

// 你查询你的文章信息我一个异步去增加阅读次数
// 查询文章，组合文章和作者的信息
func (a *articleService) GetPublishedById(ctx context.Context, id int64, uid int64) (domain.Article, error) {
	art, err := a.repo.GetPublishedById(ctx, id)
	//成功的查询出来文章的信息
	if err == nil {
		go func() {
			er := a.producer.ProduceReadEvent(
				ctx,
				events.ReadEvent{
					Uid: uid,
					Aid: id,
				})
			if er != nil {
				a.l.Error("发送读者阅读事件失败")
			}
		}()
	}
	return art, err
}

func (a *articleService) GetById(ctx context.Context, id int64) (domain.Article, error) {

	return a.repo.GetById(ctx, id)
}

func (a *articleService) List(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error) {
	//TODO implement me
	return a.repo.List(ctx, uid, offset, limit)
}

// 传递给repo的时候就不需要把结构体往下面传递你只需要传递一些值就可以了
func (a *articleService) Withdraw(ctx context.Context, art domain.Article) error {
	return a.repo.SyncStatus(ctx, art.Id, art.Author.Id, domain.ArticleStatusPrivate.ToUint8())
}

// 控制一个aythrepo库实现两者之间的发布，
func (a *articleService) Publish(ctx context.Context, art domain.Article) (int64, error) {
	//web的req是变量，我这里直接赋值常量就没事了，进来这个逻辑后手动的赋值状态码
	art.Status = domain.ArticleStatusPublished
	//return a.repo.SyncV1(ctx, art)
	return a.repo.Sync(ctx, art)
}

// 线上库的操作功能,通过art和read rep同时实现线上库和制作库的同步
func (a *articleService) PublishV1(ctx context.Context, art domain.Article) (int64, error) {
	//就算新建art是0，，，更新art也是0进行同时的id插入
	var (
		id  = art.Id
		err error
	)
	//没有id就是新文章需要创建
	if art.Id > 0 {
		//更新帖子发布
		err = a.author.UpdateById(ctx, art)
	} else {
		//创建帖子记录
		id, err = a.author.Create(ctx, art)
	}
	//更新或者创建过程返回的错误
	if err != nil {
		return 0, err
	}
	//确保文章id的一致性同步到线上库中，反正不管怎样你的id一定是可以拿到的
	art.Id = id
	for i := 0; i < 3; i++ {
		//睡眠一小段时间来修复持续一段时间所存在的问题
		time.Sleep(time.Second * time.Duration(i))
		id, err = a.reader.Save(ctx, art)
		if err == nil {
			break
		}
		a.l.Error("部分失败，保存到线上库失败", logger.Int64("art_id", art.Id),
			logger.Error(err))
	}
	if err != nil {
		a.l.Error("部分失败，重试彻底失败", logger.Int64("art_id", art.Id),
			logger.Error(err))
		//接入你的告警系统，手工处理
		//走异步，保存到本地文件
		//走cannl
		//打MQ
	}
	//实现的逻辑就是往线上库进行插入或者更新
	return id, err
}

func NewArticleService(repo article.ArticleRepository, l logger.LoggerV1, producer events.Producer) ArticleService {
	return &articleService{
		repo:     repo,
		l:        l,
		producer: producer,
	}
}
func NewArticleServiceV1(author article.ArticleAuthorRepository, reader article.ArticleReaderRepository, l logger.LoggerV1) ArticleService {
	return &articleService{
		author: author,
		reader: reader,
		l:      l,
	}
}
func (a *articleService) Save(ctx context.Context, art domain.Article) (int64, error) {
	art.Status = domain.ArticleStatusUnpublished
	//判断是新增还是修改
	if art.Id > 0 {
		//是修改
		err := a.repo.UpdateById(ctx, art)
		return art.Id, err
	}
	return a.repo.Create(ctx, art)
}

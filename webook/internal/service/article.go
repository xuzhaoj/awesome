package service

import (
	"awesomeProject/webook/internal/domain"
	"awesomeProject/webook/internal/repository/article"
	"awesomeProject/webook/pkg/logger"
	"context"
	"time"
)

type ArticleService interface {
	//就是一个创建的服务
	Save(ctx context.Context, art domain.Article) (int64, error)
	Publish(ctx context.Context, art domain.Article) (int64, error)
	PublishV1(ctx context.Context, art domain.Article) (int64, error)
}

type articleService struct {
	//因为要区分制作库和线上库，所以把一统的进行拆分
	repo article.ArticleRepository

	//V1
	author article.ArticleAuthorRepository
	reader article.ArticleReaderRepository
	l      logger.LoggerV1
}

// 制作库的操作功能
func (a *articleService) Publish(ctx context.Context, art domain.Article) (int64, error) {
	//TODO implement me
	panic("implement me")
}

// 线上库的操作功能
func (a *articleService) PublishV1(ctx context.Context, art domain.Article) (int64, error) {

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

func NewArticleService(repo article.ArticleRepository) ArticleService {
	return &articleService{
		repo: repo,
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
	//判断是新增还是修改
	if art.Id > 0 {
		//是修改
		err := a.repo.UpdateById(ctx, art)
		return art.Id, err
	}
	return a.repo.Create(ctx, art)
}

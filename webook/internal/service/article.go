package service

import (
	"awesomeProject/webook/internal/domain"
	"awesomeProject/webook/internal/repository"
	"context"
)

type ArticleService interface {
	//就是一个创建的服务
	Save(ctx context.Context, art domain.Article) (int64, error)
}

type articleService struct {
	repo repository.ArticleRepository
}

func NewArticleService(repo repository.ArticleRepository) ArticleService {
	return &articleService{
		repo: repo,
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

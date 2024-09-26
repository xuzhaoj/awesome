package repository

import (
	"awesomeProject/webook/internal/domain"
	"awesomeProject/webook/internal/repository/dao"
	"context"
)

type ArticleRepository interface {
	Create(ctx context.Context, art domain.Article) (int64, error)
	UpdateById(ctx context.Context, art domain.Article) error
}

type CachedArticleRepository struct {
	dao dao.ArticleDao
}

// 插入数据库返回的就是1，nil没有错误
func (c CachedArticleRepository) Create(ctx context.Context, art domain.Article) (int64, error) {
	return c.dao.Insert(ctx, dao.Article{
		Title:    art.Title,
		Content:  art.Content,
		AuthorId: art.Author.Id,
	})
}
func NewArticleRepository(dao dao.ArticleDao) ArticleRepository {
	return &CachedArticleRepository{
		dao: dao,
	}
}

func (c CachedArticleRepository) UpdateById(ctx context.Context, art domain.Article) error {
	return c.dao.UpdateById(ctx, dao.Article{
		//必须把文章的id传递下去
		Id:       art.Id,
		Title:    art.Title,
		Content:  art.Content,
		AuthorId: art.Author.Id,
	})
}

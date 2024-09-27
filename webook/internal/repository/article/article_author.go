package article

import (
	"awesomeProject/webook/internal/domain"
	"context"
)

type ArticleAuthorRepository interface {
	//通过service实现两个repo分别控制线上库和线下库
	Create(ctx context.Context, art domain.Article) (int64, error)
	UpdateById(ctx context.Context, art domain.Article) error
}

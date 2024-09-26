package article

import (
	"awesomeProject/webook/internal/domain"
	"context"
)

type ArticleAuthorRepository interface {
	Create(ctx context.Context, art domain.Article) (int64, error)
	UpdateById(ctx context.Context, art domain.Article) error
	//UpdateById(ctx context.Context, art domain.Article) error
}

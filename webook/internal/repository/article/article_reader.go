package article

import (
	"awesomeProject/webook/internal/domain"
	"context"
)

type ArticleReaderRepository interface {
	//有就更新，没有就新建，即update，insert语义
	Save(ctx context.Context, art domain.Article) (int64, error)
}

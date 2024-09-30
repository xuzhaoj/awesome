package article

import (
	"context"
	"gorm.io/gorm"
)

type ReaderDAO interface {
	Upsert(ctx context.Context, art Article) error
	UpsertV2(ctx context.Context, art PublishedArticle) error
}

// 这个代表的是线上库，同库不同表的操作
type PublishedArticle struct {
	Article
}

func NewReaderDAO(db *gorm.DB) ReaderDAO {
	panic("implement me")
}

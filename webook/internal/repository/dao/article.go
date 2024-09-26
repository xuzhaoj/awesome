package dao

import (
	"context"
	"fmt"
	"gorm.io/gorm"
	"time"
)

// 这个人是要存储到数据库的
type ArticleDao interface {
	Insert(ctx context.Context, art Article) (int64, error)
	UpdateById(ctx context.Context, article Article) error
}

func NewGORMArticleDao(db *gorm.DB) ArticleDao {
	return &GORMArticleDao{
		db: db,
	}
}

type GORMArticleDao struct {
	db *gorm.DB
}

// 制作库
type Article struct {
	Id      int64  `gorm:"primaryKey,autoIncrement"`
	Title   string `gorm:"type=varchar(1024)"`
	Content string `gorm:"type=BLOB"`
	//设计索引,为什么要创建联合索引
	//AuthorId int64 `gorm:"index=aid_ctime"`
	//Ctime    int64 `gorm:"index=aid_ctime"`
	AuthorId int64 `gorm:"index"`
	Ctime    int64
	Utime    int64
}

func (dao *GORMArticleDao) Insert(ctx context.Context, art Article) (int64, error) {
	now := time.Now().UnixMilli()
	art.Ctime = now
	art.Utime = now
	err := dao.db.WithContext(ctx).Create(&art).Error
	//返回插入生成的id
	return art.Id, err
}

//func (dao *GORMArticleDao) UpdateById(ctx context.Context, art Article) error {
//	now := time.Now().UnixMilli()
//	art.Utime = now
//	//gorm会忽略0值的特性，会用主键进行更新
//	//可读性很差
//	//err := dao.db.WithContext(ctx).Create(&art).Error
//	err := dao.db.WithContext(ctx).Model(&art).
//		Where("id=?", art.Id).
//		Updates(map[string]any{
//			"title":   art.Title,
//			"content": art.Content,
//			"utime":   art.Utime,
//		}).Error
//
//	return err
//}

func (dao *GORMArticleDao) UpdateById(ctx context.Context, art Article) error {
	now := time.Now().UnixMilli()
	//art.Utime = now
	//gorm会忽略0值的特性，会用主键进行更新
	//可读性很差
	//res := dao.db.WithContext(ctx).Model(&art).
	//	Where("id=? AND author_id=?", art.Id, art.AuthorId).
	//	Updates(map[string]any{
	//		"title":   art.Title,
	//		"content": art.Content,
	//		"utime":   art.Utime,
	//	})
	res := dao.db.WithContext(ctx).Model(&art).
		Where("id = ? AND author_id = ?", art.Id, art.AuthorId).
		Updates(map[string]any{
			"title":   art.Title,
			"content": art.Content,
			"utime":   now,
		})

	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		//这样写是最好的用来排查错误信息
		return fmt.Errorf("更新失败，可能是创作者非法id %d,author_id %d", art.Id, art.AuthorId)
	}

	return res.Error
}

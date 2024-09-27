package article

import (
	"context"
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

// 这个人是要存储到数据库的
type ArticleDao interface {
	Insert(ctx context.Context, art Article) (int64, error)
	UpdateById(ctx context.Context, article Article) error
	Sync(ctx context.Context, article Article) (int64, error)
	Upsert(ctx context.Context, art PublishArticle) error
}

func NewGORMArticleDao(db *gorm.DB) ArticleDao {
	return &GORMArticleDao{
		db: db,
	}
}

type GORMArticleDao struct {
	db *gorm.DB
}

func (dao *GORMArticleDao) Sync(ctx context.Context, art Article) (int64, error) {
	var (
		id = art.Id
	)
	//事务的内部开启了闭包的形态   tx是Transaction
	err := dao.db.Transaction(func(tx *gorm.DB) error {
		var err error
		txDAO := NewGORMArticleDao(tx)
		//判断执行的是新建还是更新
		if id > 0 {
			err = txDAO.UpdateById(ctx, art)

		} else {
			id, err = txDAO.Insert(ctx, art)
		}
		if err != nil {
			//两个数据库中有一个出错了
			return err
		}
		art.Id = id
		//往reader数据库中去写入数
		return txDAO.Upsert(ctx, PublishArticle{Article: art})

	})
	return id, err
}

// 插入新记录更新现有记录：如果插入时发生了主键或唯一索引冲突执行更新操作，只更新指定的字段，而不会插入新的记录。
func (dao *GORMArticleDao) Upsert(ctx context.Context, art PublishArticle) error {
	now := time.Now().UnixMilli()
	art.Ctime = now
	art.Utime = now
	//这个是插入，不支持带where的写法这个就是upsert的局限性
	err := dao.db.Clauses(clause.OnConflict{
		//只会更新内容
		DoUpdates: clause.Assignments(map[string]interface{}{
			"title":   art.Title,
			"content": art.Content,
			"utime":   now,
		}),
	}).Create(&art).Error
	//如果有数据冲突就执行对应xxx
	//MYSQL最终的语句INSERT xxx on DUPICATE KEY UPDATE xxx
	return err
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
	//GORM 会将数据库生成的字段值自动填充到 art 结构体的相关字段中
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

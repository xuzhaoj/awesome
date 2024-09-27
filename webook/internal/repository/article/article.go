package article

import (
	"awesomeProject/webook/internal/domain"
	dao "awesomeProject/webook/internal/repository/dao/article"
	"context"
	"gorm.io/gorm"
)

// repository应该用来操作缓存和DAO
// 事务概念应该在DAO这层进行
type ArticleRepository interface {
	Create(ctx context.Context, art domain.Article) (int64, error)
	UpdateById(ctx context.Context, art domain.Article) error
	//通过存储并更新数据库
	SyncV1(ctx context.Context, art domain.Article) (int64, error)
}

type CachedArticleRepository struct {
	dao dao.ArticleDao

	//V1 操作两个DAO
	readerDAO dao.ReaderDAO
	authorDAO dao.AuthorDAO
	db        *gorm.DB
}

func (c *CachedArticleRepository) Sync(ctx context.Context, art domain.Article) (int64, error) {
	return c.dao.Sync(ctx, c.toEntity(art))

}

// 尝试在repository层面上解决事务（保存到线上库和制作库同时成功）问题
func (c *CachedArticleRepository) SyncV2(ctx context.Context, art domain.Article) (int64, error) {
	//开启事务
	tx := c.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return 0, tx.Error
	}
	//如果出现了panic的情况，defer 关键字会延迟 tx.Rollback() 的执行，直到当前函数返回时才会执行。事务结束都释放了
	defer tx.Rollback()
	author := dao.NewAuthorDAO(tx)
	reader := dao.NewReaderDAO(tx)
	var (
		id  = art.Id
		err error
	)
	if id > 0 {
		//说明是修改更新
		err = author.UpdateById(ctx, c.toEntity(art))
	} else {
		id, err = author.Insert(ctx, c.toEntity(art))
	}
	if err != nil {
		//执行出问题要回滚
		//tx.Rollback()
		return id, err
	}
	//制作库更新成功后操作线上库，进行同步,考虑到线上有或者没用的情况

	//数据库要使用两张表的情况应该调用下面的方法
	//err = reader.UpsertV2(ctx, dao.PublishArticle{Article: c.toEntity(art)})
	err = reader.Upsert(ctx, c.toEntity(art))
	//执行成功直接提交
	tx.Commit()
	return id, nil
}
func (c *CachedArticleRepository) SyncV1(ctx context.Context, art domain.Article) (int64, error) {
	//先保存发哦制作库在保存到线上库---------同时操作两个dao
	var (
		id  = art.Id
		err error
	)
	if id > 0 {
		//说明是修改更新
		err = c.authorDAO.UpdateById(ctx, c.toEntity(art))
	} else {
		id, err = c.authorDAO.Insert(ctx, c.toEntity(art))
	}
	if err != nil {
		return id, err
	}
	art.Id = id
	//制作库更新成功后操作线上库，进行同步,考虑到线上有或者没用的情况
	c.readerDAO.Upsert(ctx, c.toEntity(art))
	return id, nil
}

// 插入数据库返回的就是1，nil没有错误
func (c *CachedArticleRepository) Create(ctx context.Context, art domain.Article) (int64, error) {
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
func (c *CachedArticleRepository) toEntity(art domain.Article) dao.Article {
	return dao.Article{
		Id:       art.Id,
		Title:    art.Title,
		Content:  art.Content,
		AuthorId: art.Author.Id,
	}
}
func (c *CachedArticleRepository) UpdateById(ctx context.Context, art domain.Article) error {
	return c.dao.UpdateById(ctx, dao.Article{
		//必须把文章的id传递下去
		Id:       art.Id,
		Title:    art.Title,
		Content:  art.Content,
		AuthorId: art.Author.Id,
	})
}

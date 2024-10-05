package article

import (
	"awesomeProject/webook/internal/domain"
	"awesomeProject/webook/internal/repository"
	"awesomeProject/webook/internal/repository/cache"
	dao "awesomeProject/webook/internal/repository/dao/article"
	"awesomeProject/webook/pkg/logger"
	"context"
	"github.com/ecodeclub/ekit/slice"
	"gorm.io/gorm"
	"time"
)

// repository应该用来操作缓存和DAO
// 事务概念应该在DAO这层进行
type ArticleRepository interface {
	Create(ctx context.Context, art domain.Article) (int64, error)
	UpdateById(ctx context.Context, art domain.Article) error
	//通过存储并更新数据库
	SyncV1(ctx context.Context, art domain.Article) (int64, error)
	//发帖功能增强的实现，草稿的保存到同步线上库，制作库
	Sync(ctx context.Context, art domain.Article) (int64, error)
	//撤回功能
	SyncStatus(ctx context.Context, id int64, author int64, status uint8) error
	//查询文章的数据
	List(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error)
	GetById(ctx context.Context, id int64) (domain.Article, error)
	GetPublishedById(ctx context.Context, id int64) (domain.Article, error)
	ListPub(ctx context.Context, start time.Time, offset int, limit int) ([]domain.Article, error)
}

type CachedArticleRepository struct {
	dao      dao.ArticleDao
	userRepo repository.UserRepository
	//V1 操作两个DAO
	readerDAO dao.ReaderDAO
	authorDAO dao.AuthorDAO
	db        *gorm.DB
	//另外一个包进行使用的时候就得包名。去调用
	cache cache.ArticleCache
	l     logger.LoggerV1
}

func (c *CachedArticleRepository) ListPub(ctx context.Context, start time.Time, offset int, limit int) ([]domain.Article, error) {
	//返回的数据是文章的切片
	res, err := c.dao.ListPub(ctx, start, offset, limit)
	if err != nil {
		return nil, err
	}
	//每个 dao.Article 转换为 domain.Article，最终返回一个 domain.Article 切片。
	return slice.Map(res, func(idx int, src dao.Article) domain.Article {
		return c.toDomain(src)
	}), nil
}

func (c *CachedArticleRepository) GetPublishedById(
	ctx context.Context, id int64) (domain.Article, error) {

	//读者查询对应文章的id信息
	art, err := c.dao.GetPubById(ctx, id)
	if err != nil {
		return domain.Article{}, err

	}
	//文章里面存放着作者的id   直接在userpository中操作查询信息
	user, err := c.userRepo.FindById(ctx, art.AuthorId)
	//把数据库中查到的信息进行转化
	res := domain.Article{
		Id:      art.Id,
		Title:   art.Title,
		Content: art.Content,
		//这个字段里面还嵌套别的字段
		Status: domain.ArticleStatus(art.Status),
		Author: domain.Author{
			Id:   art.AuthorId,
			Name: user.NickName,
		},
		Ctime: time.UnixMilli(art.Ctime),
		Utime: time.UnixMilli(art.Utime),
	}
	return res, nil
}

// 根据id对文章的详情信息就行复现
func (c *CachedArticleRepository) GetById(ctx context.Context, id int64) (domain.Article, error) {
	data, err := c.dao.GetById(ctx, id)
	if err != nil {
		return domain.Article{}, err
	}
	//转化为domain对象
	return c.toDomain(data), nil

}

// 查询作者所写的所有文章接口
func (c *CachedArticleRepository) List(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error) {
	//在这个地方集成复杂的缓存方案
	//缓存第一页100条数据
	if offset == 0 && limit <= 100 {
		//先往缓存中查
		data, err := c.cache.GetFirstPage(ctx, uid)
		if err == nil {
			go func() {
				c.preCache(ctx, data)
			}()
			//return data[:limit],err//应该是返回这个的
			return data, err
		}

	}
	//数据库查
	//查询出来了一条条的文章数组，同时我们要将其映射成domain类型
	res, err := c.dao.GetByAuthor(ctx, uid, offset, limit)
	if err != nil {
		return nil, err
	}

	//要把数据库查询出来的数据转化成domain的形式往回传，你也可以在dao中进行转化先往下看

	data := slice.Map[dao.Article, domain.Article](res, func(idx int, src dao.Article) domain.Article {
		return c.toDomain(src)
	})
	//回写缓存set/del
	//启动了一个新的 goroutine（协程），使得其中的代码异步执行，而不阻塞主线程或当前函数的执行。
	//goroutine，程序可以在缓存操作执行时继续执行其他操作，不会因为等待缓存写入而停滞。
	go func() {
		err := c.cache.SetFirstPage(ctx, uid, data)
		//println(err.Error())
		c.l.Error("回写缓存失败", logger.Error(err))
		c.preCache(ctx, data)
	}()
	return data, nil

}

// 让调用者觉得是否异步
func (c *CachedArticleRepository) preCache(ctx context.Context, data []domain.Article) {
	//太长表示不缓存
	if len(data) > 0 && len(data[0].Content) < 1024*1024 {
		//查询数据库后缓存第一条数据便于文章进行调用
		err := c.cache.Set(ctx, data[0])
		if err != nil {
			c.l.Error("缓存预加载失败", logger.Error(err))
		}
	}
}

// 这就是一个前端传递数据的赋值
func (req *CachedArticleRepository) toDomain(art dao.Article) domain.Article {
	return domain.Article{

		Id:    art.Id,
		Title: art.Title,
		//类型转化
		Status:  domain.ArticleStatus(art.Status),
		Content: art.Content,
		Author: domain.Author{
			Id: art.AuthorId,
		},
		Ctime: time.UnixMilli(art.Ctime),
		Utime: time.UnixMilli(art.Utime),
	}

}

func (c *CachedArticleRepository) SyncStatus(ctx context.Context, id int64, author int64, status uint8) error {
	return c.dao.SyncStatus(ctx, id, author, uint8(status))
}

// 保存到线上库和线下库,,,,,,,,,,,数据库保存好了之后再才在redis上面进行domain数据的存储
func (c *CachedArticleRepository) Sync(ctx context.Context, art domain.Article) (int64, error) {
	id, err := c.dao.Sync(ctx, c.toEntity(art))
	if err == nil {
		//缓存一般都设置==进行工作
		c.cache.DelFirstPage(ctx, art.Author.Id)
		c.cache.SetPub(ctx, art)
	}
	return id, err
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
	defer func() {
		//清空缓存
		c.cache.DelFirstPage(ctx, art.Author.Id)
	}()
	return c.dao.Insert(ctx, dao.Article{
		Title:    art.Title,
		Content:  art.Content,
		AuthorId: art.Author.Id,
		Status:   art.Status.ToUint8(),
	})
}
func NewArticleRepository(dao dao.ArticleDao, l logger.LoggerV1) ArticleRepository {
	return &CachedArticleRepository{
		dao: dao,
		l:   l,
	}
}
func (c *CachedArticleRepository) toEntity(art domain.Article) dao.Article {
	return dao.Article{
		Id:       art.Id,
		Title:    art.Title,
		Content:  art.Content,
		AuthorId: art.Author.Id,
		Status:   art.Status.ToUint8(),
	}
}
func (c *CachedArticleRepository) UpdateById(ctx context.Context, art domain.Article) error {
	defer func() {
		//清空缓存
		c.cache.DelFirstPage(ctx, art.Author.Id)
	}()
	return c.dao.UpdateById(ctx, dao.Article{
		//必须把文章的id传递下去
		Id:       art.Id,
		Title:    art.Title,
		Content:  art.Content,
		AuthorId: art.Author.Id,
		Status:   art.Status.ToUint8(),
	})
}

package repository

import (
	"awesomeProject/webook/internal/domain"
	"awesomeProject/webook/internal/repository/cache"
	"awesomeProject/webook/internal/repository/dao"
	"awesomeProject/webook/pkg/logger"
	"context"
)

type InteractiveRepository interface {
	IncrReadCnt(ctx context.Context, biz string, bizId int64) error
	// BatchIncrReadCnt biz 和 bizId 长度必须一致
	BatchIncrReadCnt(ctx context.Context, biz []string, bizId []int64) error
	IncrLike(ctx context.Context, biz string, bizId int64, uid int64) error
	DecrLike(ctx context.Context, biz string, bizId int64, uid int64) error
	AddCollectionItem(ctx context.Context, biz string, bizId int64, cid int64, uid int64) error
	Get(ctx context.Context, biz string, bizId int64) (domain.Interactive, error)
	Liked(ctx context.Context, biz string, id int64, uid int64) (bool, error)
	Collected(ctx context.Context, biz string, id int64, uid int64) (bool, error)
}

// 都是接口管那么干嘛，不然就是结构体
type CacheReadCntRepository struct {
	cache cache.InteractiveCache
	dao   dao.InteractiveDAO
	l     logger.LoggerV1
}

func NewCachedInteractiveRepository(cache cache.InteractiveCache,
	dao dao.InteractiveDAO, l logger.LoggerV1) InteractiveRepository {
	return &CacheReadCntRepository{cache: cache, dao: dao, l: l}
}

// 现在数据库中增加阅读数，然后把阅读数存储在redis
func (c *CacheReadCntRepository) IncrReadCnt(ctx context.Context,
	biz string, bizId int64) error {
	//考虑缓存方案
	//更新数据库
	err := c.dao.IncrReadCnt(ctx, biz, bizId)
	if err != nil {
		return err
	}
	//更新缓存
	return c.cache.IncrReadCntIfPresent(ctx, biz, bizId)
}

func (c *CacheReadCntRepository) BatchIncrReadCnt(ctx context.Context, biz []string, bizId []int64) error {
	//TODO implement me
	panic("implement me")
}

// 点赞
func (c *CacheReadCntRepository) IncrLike(ctx context.Context, biz string, bizId int64, uid int64) error {
	//Article,文章id,用户id"  存储的层面上就是增加点赞数取消点赞数
	//逻辑梳理-  DAO中操作两张表   1先插入点赞然后   2更新点赞计数，             3更新缓存
	err := c.dao.InsertLikeInfo(ctx, biz, bizId, uid)
	if err != nil {
		return err
	}
	//将数据写入缓存
	return c.cache.IncrLikeCntIfPresent(ctx, biz, bizId)

}

// 取消点赞功能的实现
func (c *CacheReadCntRepository) DecrLike(ctx context.Context, biz string, bizId int64, uid int64) error {
	err := c.dao.DeleteLikeInfo(ctx, biz, bizId, uid)
	if err != nil {
		return err
	}

	return c.cache.DecrLikeCntIfPresent(ctx, biz, bizId)

}

// 收藏夹功能的实现，还有实现计数
func (c *CacheReadCntRepository) AddCollectionItem(ctx context.Context,
	biz string, bizId int64, cid int64, uid int64) error {
	err := c.dao.InsertCollectionBiz(ctx, dao.UserCollectionBiz{
		Cid:   cid,
		Biz:   biz,
		BizId: bizId,
		Uid:   uid,
	})
	if err != nil {
		return err
	}
	//收藏个数      biz+bizId
	return c.cache.IncrCollectCntIfPresent(ctx, biz, bizId)

}

func (c *CacheReadCntRepository) Get(ctx context.Context,
	biz string, bizId int64) (domain.Interactive, error) {
	//***************************要从缓存中拿出来阅读数，点赞数和收藏数，如果拿不到那么就查询数据库
	intr, err := c.cache.Get(ctx, biz, bizId)
	if err != nil {
		return intr, nil
	}

	//******************************************************查询数据库,
	daoIntr, err := c.dao.Get(ctx, biz, bizId)
	//数据库查询出错的时候才会err！=nil
	if err != nil {
		return domain.Interactive{}, err
	}
	//将查询出来的dao数据的部分字段封装到domain中
	intr = domain.Interactive{
		//指针的使用最简单的原则，接收器永远用指针，输入输出都使用结构体去接受
		ReadCnt:    daoIntr.ReadCnt,
		LikeCnt:    daoIntr.LikeCnt,
		CollectCnt: daoIntr.CollectCnt,
	}

	//******************************可以开一个异步把数据库终于查询出来的数据放到redis缓存中
	go func() {
		//缓存把这些东西给写进去,异步不需要把缓存写进去
		er := c.cache.Set(ctx, biz, bizId, intr)
		if er != nil {
			c.l.Error("回写缓存失败",
				logger.String("biz", biz),
				logger.Int64("bizId", bizId),
			)
		}
	}()
	return intr, nil

}

func (c *CacheReadCntRepository) Liked(ctx context.Context, biz string, id int64, uid int64) (bool, error) {
	_, err := c.dao.GetLikeInfo(ctx, biz, id, uid)
	switch err {
	case nil:
		return true, nil
	case dao.ErrRecordNotFound:
		return false, nil
	default:
		return false, err
	}

}

func (c *CacheReadCntRepository) Collected(ctx context.Context, biz string, id int64, uid int64) (bool, error) {
	_, err := c.dao.GetCollectInfo(ctx, biz, id, uid)
	switch err {
	case nil:
		return true, nil
	case dao.ErrRecordNotFound:
		return false, nil
	default:
		return false, err
	}
}

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

func (c *CacheReadCntRepository) AddCollectionItem(ctx context.Context, biz string, bizId int64, cid int64, uid int64) error {
	//TODO implement me
	panic("implement me")
}

func (c *CacheReadCntRepository) Get(ctx context.Context, biz string, bizId int64) (domain.Interactive, error) {
	//TODO implement me
	panic("implement me")
}

func (c *CacheReadCntRepository) Liked(ctx context.Context, biz string, id int64, uid int64) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (c *CacheReadCntRepository) Collected(ctx context.Context, biz string, id int64, uid int64) (bool, error) {
	//TODO implement me
	panic("implement me")
}

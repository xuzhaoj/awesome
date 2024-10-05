package repository

import (
	"awesomeProject/webook/internal/domain"
	"awesomeProject/webook/internal/repository/cache"
	"context"
)

type RankingRepository interface {

	//热榜数据没有放到数据库中直接放在缓存中---组装定时任务我不知道怎么用我只知道要放在缓存中-分为本地缓存和redis缓存

	//替换热榜数据
	ReplaceTopN(ctx context.Context, arts []domain.Article) error
	//获取热榜数据
	GetTopN(ctx context.Context) ([]domain.Article, error)
}

type CachedRankingRepository struct {
	//使用具体的实现代码的可读性会更好      本地缓存的访问速度快-实时性搞的场景才需要
	redis *cache.RankingRedisCache

	//本地缓存，每个应用的实例维护自己的缓存无法共享
	local *cache.RankingLocalCache
}

func NewCachedRankingRepository(redis *cache.RankingRedisCache,
	local *cache.RankingLocalCache) *CachedRankingRepository {

	return &CachedRankingRepository{redis: redis, local: local}

}

// 获取热榜数据
func (c *CachedRankingRepository) GetTopN(ctx context.Context) ([]domain.Article, error) {

	data, err := c.local.Get(ctx) // 优先从本地缓存获取数据
	if err != nil {
		return data, nil // 如果本地缓存有数据且未过期，直接返回
	}
	data, err = c.redis.Get(ctx) // 如果本地缓存没有数据或已过期，则尝试从 Redis 获取数据
	if err == nil {
		c.local.Set(ctx, data) // 如果从 Redis 获取成功，将数据写入本地缓存
	} else {
		return c.local.ForceGet(ctx) // 如果 Redis 获取失败，强制从本地缓存获取
	}
	return data, err // 返回从 Redis 或本地缓存获取到的数据
}

// 替换热榜数据
func (c *CachedRankingRepository) ReplaceTopN(ctx context.Context, arts []domain.Article) error {
	//先操作本地缓存
	_ = c.local.Set(ctx, arts)    // 先更新本地缓存
	return c.redis.Set(ctx, arts) // 再更新 Redis 缓存

}

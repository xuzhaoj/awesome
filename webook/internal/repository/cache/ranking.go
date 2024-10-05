package cache

import (
	"awesomeProject/webook/internal/domain"
	"context"
	"encoding/json"
	"github.com/redis/go-redis/v9"
	"time"
)

type RankingCache interface {
	Set(ctx context.Context, arts []domain.Article) error
	Get(ctx context.Context) ([]domain.Article, error)
}

type RankingRedisCache struct {
	client *redis.Client
	key    string
}

func NewRankingRedisCache(client *redis.Client) RankingCache {
	return &RankingRedisCache{
		client: client,
		key:    "ranking",
	}
}

// 了确保在计算热榜时能够快速访问数据
func (r *RankingRedisCache) Set(ctx context.Context, arts []domain.Article) error {
	for i := 0; i < len(arts); i++ {
		//文章数据存入Redis之前，将每篇文章的正文内容（Content字段）清空。的元数据（如标题、ID等），而不需要正文内容。
		arts[i].Content = ""
	}
	//json.Marshal 方法序列化成JSON格式的字节数组，以便存储到Redis中。-domain转字节
	val, err := json.Marshal(&arts)
	if err != nil {
		return err
	}
	//这个过期时间要设置长一点，超过计算热榜的时间
	return r.client.Set(ctx, r.key, val, time.Minute*10).Err()
}

func (r *RankingRedisCache) Get(ctx context.Context) ([]domain.Article, error) {
	data, err := r.client.Get(ctx, r.key).Bytes()
	if err != nil {
		return nil, err
	}
	var res []domain.Article
	err = json.Unmarshal(data, &res)
	return res, err
}

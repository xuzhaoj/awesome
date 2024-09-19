package cache

import (
	"awesomeProject/webook/internal/domain"
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

var ErrKeyNotExist = redis.Nil

type UserCache struct {
	client     redis.Cmdable
	expiration time.Duration
}

func NewUserCache(client redis.Cmdable) *UserCache {
	return &UserCache{
		client:     client,
		expiration: time.Minute * 15,
	}
}

func (cache *UserCache) Set(ctx context.Context, u domain.User) error {
	val, err := json.Marshal(u)
	if err != nil {
		return err
	}
	//获取键名字
	key := cache.key(u.Id)
	//直接通过err（）返回就好成功就是返回nil，不成功就是返回错误信息
	return cache.client.Set(ctx, key, val, cache.expiration).Err()
}

// 这个是用来生成键，值需要自己去传递
func (cache *UserCache) key(id int64) string {
	return fmt.Sprintf("user:info:%d", id)
}
func (cache *UserCache) Get(ctx context.Context, id int64) (domain.User, error) {
	key := cache.key(id)
	//返回的是一个redis.StringCmd，需要调用result（）
	val, err := cache.client.Get(ctx, key).Result()
	if err != nil {
		return domain.User{}, err
	}
	u := domain.User{}
	err = json.Unmarshal([]byte(val), &u)
	return u, err
}

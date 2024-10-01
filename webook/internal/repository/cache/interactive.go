package cache

import (
	"awesomeProject/webook/internal/domain"
	"context"
	_ "embed"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

type InteractiveCache interface {
	IncrReadCntIfPresent(ctx context.Context, biz string, bizId int64) error
	IncrLikeCntIfPresent(ctx context.Context, biz string, id int64) error
	DecrLikeCntIfPresent(ctx context.Context, biz string, id int64) error
	IncrCollectCntIfPresent(ctx context.Context, biz string, id int64) error
	Get(ctx context.Context, biz string, id int64) (domain.Interactive, error)
	Set(ctx context.Context, biz string, bizId int64, res domain.Interactive) error
}

type InteractiveRedisCache struct {
	client     redis.Cmdable
	expiration time.Duration
}

//go:embed lua/interative_incr_cnt.lua
var luaIncrCnt string

const fieldReadCnt = "read_cnt"
const fieldLikeCnt = "like_cnt"
const fieldCollectCnt = "collect_cnt"

// 生成redis的key
func (i *InteractiveRedisCache) key(biz string, bizId int64) string {
	return fmt.Sprintf("interactive:%s:%d", biz, bizId)
}

// 方案一将点赞数收藏数阅读数直接存储成一个map结构只需要一个查询就可以生成
// 方案二，生成三个主键，这个就意味着需要查询三次
// 阅读次数通过lua脚本解决，使用了hash存储结构，通过lua存储，类似一个map
func (i *InteractiveRedisCache) IncrReadCntIfPresent(ctx context.Context, biz string, bizId int64) error {
	//生成key
	key := i.key(biz, bizId)

	return i.client.Eval(ctx, luaIncrCnt,
		[]string{key}, fieldReadCnt, 1).Err()
}

// 点赞数的缓存,缓存什么啊这个
func (i *InteractiveRedisCache) IncrLikeCntIfPresent(ctx context.Context, biz string, id int64) error {
	key := i.key(biz, id)
	return i.client.Eval(ctx, luaIncrCnt,
		[]string{key}, fieldLikeCnt, 1).Err()
}

// 取消点赞数的缓存
func (i *InteractiveRedisCache) DecrLikeCntIfPresent(ctx context.Context, biz string, id int64) error {
	key := i.key(biz, id)
	return i.client.Eval(ctx, luaIncrCnt,
		[]string{key}, fieldLikeCnt, -1).Err()
}

func (i *InteractiveRedisCache) IncrCollectCntIfPresent(ctx context.Context, biz string, id int64) error {
	//TODO implement me
	panic("implement me")
}

func (i *InteractiveRedisCache) Get(ctx context.Context, biz string, id int64) (domain.Interactive, error) {
	//TODO implement me
	panic("implement me")
}

func (i *InteractiveRedisCache) Set(ctx context.Context, biz string, bizId int64, res domain.Interactive) error {
	//TODO implement me
	panic("implement me")
}

func NewInteractiveRedisCache(client redis.Cmdable) InteractiveCache {
	return &InteractiveRedisCache{
		client:     client,
		expiration: time.Hour,
	}
}

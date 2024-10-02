package cache

import (
	"awesomeProject/webook/internal/domain"
	"context"
	_ "embed"
	"fmt"
	"github.com/redis/go-redis/v9"
	"strconv"
	"time"
)

type InteractiveCache interface {
	IncrReadCntIfPresent(ctx context.Context, biz string, bizId int64) error
	IncrLikeCntIfPresent(ctx context.Context, biz string, id int64) error
	DecrLikeCntIfPresent(ctx context.Context, biz string, id int64) error
	IncrCollectCntIfPresent(ctx context.Context, biz string, id int64) error
	Get(ctx context.Context, biz string, bizId int64) (domain.Interactive, error)
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
	key := i.key(biz, id)
	return i.client.Eval(ctx, luaIncrCnt,
		[]string{key}, fieldCollectCnt, 1).Err()
}

func (i *InteractiveRedisCache) Get(ctx context.Context, biz string, bizId int64) (domain.Interactive, error) {
	key := i.key(biz, bizId)
	//拿到key里面所有的VAL
	//map[string]string类型map[string]string{
	//    "collect_cnt": "10",
	//    "read_cnt": "200",
	//    "like_cnt": "50",
	//}
	//原生的数据不需要反序列化
	data, err := i.client.HGetAll(ctx, key).Result()
	//缓存如果找不到key会返回空,这里的err主要是超时错误，客户端配置不行
	if err != nil {
		return domain.Interactive{}, err
	}

	if len(data) == 0 {
		//缓存不存在----------key找不到返回的结果返回空对象
		return domain.Interactive{}, ErrKeyNotExist
	}
	//字段的值从字符串转换为 int64 类型  因为你相应的点赞收藏还有阅读数都是数字类型
	collectCnt, _ := strconv.ParseInt(data[fieldCollectCnt], 10, 64)
	likeCnt, _ := strconv.ParseInt(data[fieldLikeCnt], 10, 64)
	readCnt, _ := strconv.ParseInt(data[fieldReadCnt], 10, 64)
	return domain.Interactive{
		CollectCnt: collectCnt,
		LikeCnt:    likeCnt,
		ReadCnt:    readCnt,
	}, err

}

// 将数据库中查询到的点赞数收藏数还有阅读数存储到缓存中
func (i *InteractiveRedisCache) Set(ctx context.Context,
	biz string, bizId int64, res domain.Interactive) error {

	//存储了数据库返回过来的点赞数收藏数还有阅读数         Hash形式
	//KEY  interactive:article:123  Val： collect_cnt 10 KEY-VAL  read_cnt 200 like_cnt 50
	key := i.key(biz, bizId)
	err := i.client.HSet(ctx, key, fieldCollectCnt, res.CollectCnt,
		fieldReadCnt, res.ReadCnt,
		fieldLikeCnt, res.LikeCnt).Err()
	if err != nil {
		return err
	}
	return i.client.Expire(ctx, key, time.Minute*15).Err()
}

func NewInteractiveRedisCache(client redis.Cmdable) InteractiveCache {
	return &InteractiveRedisCache{
		client:     client,
		expiration: time.Hour,
	}
}

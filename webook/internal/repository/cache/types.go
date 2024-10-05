package cache

import (
	"context"
	"github.com/ecodeclub/ekit"
	"time"
)

type Cache interface {
	Set(ctx context.Context, key string, val any, exp time.Duration) error
	Get(ctx context.Context, key string) ekit.AnyValue
}

type LocalCache struct {
}
type RedisCache struct {
}
type DoubleCache struct {
	local Cache
	redis Cache
}

func (d *DoubleCache) Set(ctx context.Context, key string, val any, exp time.Duration) error {

	//TODO implement me
	panic("implement me")
}

func (d *DoubleCache) Get(ctx context.Context, key string) ekit.AnyValue {
	//TODO implement me
	panic("implement me")
}

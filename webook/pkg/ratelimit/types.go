package ratelimit

import "context"

type Limiter interface {
	//有无触发限流 key 就是限流的对象,true代表限流
	Limit(ctx context.Context, key string) (bool, error)
}

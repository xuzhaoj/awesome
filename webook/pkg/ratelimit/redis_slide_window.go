package ratelimit

import (
	"context"
	_ "embed"
	"github.com/redis/go-redis/v9"
	"time"
)

// Redis上面的滑动窗口算法限流器的实现
type RedisSlidingWindowLimiter struct {
	cmd redis.Cmdable
	//窗口大小
	interval time.Duration
	// 阈值
	rate int
}

// 返回接口类型接口变量保存了指向redisSlidingWindowLimiter的指针
func NewRedisSlidingWindowLimiter(cmd redis.Cmdable,
	interval time.Duration, rate int) Limiter {
	return &RedisSlidingWindowLimiter{
		cmd:      cmd,
		interval: interval,
		rate:     rate,
	}

}

//go:embed slide_window.lua
var luaSlideWindow string

// 结构体内的属性
// 结构体类型实现了接口中的方法，方法名，参数，返回值相同就代表实现了接口，同一包下均可不同包下只需要导入包名即可
func (r *RedisSlidingWindowLimiter) Limit(ctx context.Context, key string) (bool, error) {

	return r.cmd.Eval(ctx, luaSlideWindow, []string{key},
		r.interval.Milliseconds(), r.rate, time.Now().UnixMilli()).Bool()

}

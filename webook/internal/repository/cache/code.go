package cache

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
)

// 编译器会在编译的时候将变量放进来这个变量里面
//
//go:embed lua/set_code.lua
var luaSetCode string

//go:embed lua/verify_code.lua
var luaVerifyCode string

var (
	ErrCodeSendTooMany        = errors.New("验证码发送太频繁")
	ErrCodeVerifyTooManyTimes = errors.New("验证次数太多")
	ErrUnknownForceCode       = errors.New("我也不知道是什么错误")
)

type CodeCache struct {
	client redis.Cmdable
}

func NewCodeCache(client redis.Cmdable) *CodeCache {
	return &CodeCache{
		client: client,
	}

}
func (c *CodeCache) Set(ctx context.Context, biz, phone, code string) error {
	//执行lua脚本的命令与redis进行交互
	res, err := c.client.Eval(ctx, luaSetCode, []string{c.key(biz, phone)}, code).Int()
	if err != nil {
		return err
	}
	switch res {
	case 0:
		return nil
	case -1:
		return ErrCodeSendTooMany
	default:
		return errors.New("系统错误")

	}

}
func (c *CodeCache) Verify(ctx context.Context, biz, phone, inputCode string) (bool, error) {
	res, err := c.client.Eval(ctx, luaVerifyCode, []string{c.key(biz, phone)}, inputCode).Int()
	if err != nil {
		return false, err
	}
	switch res {
	case 0:
		return true, nil
	case -1:
		return false, ErrCodeVerifyTooManyTimes
	case -2:
		return false, nil
	default:
		return false, ErrUnknownForceCode
	}

}

func (c *CodeCache) key(biz, phone string) string {
	//模板字符串
	return fmt.Sprintf("phone_code:%s:%s", biz, phone)

}

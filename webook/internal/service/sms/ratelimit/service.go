package ratelimit

import (
	"awesomeProject/webook/internal/service/sms"
	"awesomeProject/webook/pkg/ratelimit"
	"context"
	"fmt"
)

var errLimited = fmt.Errorf("触发了限流")

type RateLimitSMSService struct {
	//被装饰的东西
	svc     sms.Service
	limiter ratelimit.Limiter
}

func NewRateLimitSMSService(svc sms.Service, limiter ratelimit.Limiter) sms.Service {
	return &RateLimitSMSService{
		svc:     svc,
		limiter: limiter,
	}

}

func (s *RateLimitSMSService) Send(ctx context.Context, tpl string, args []string, numbers ...string) error {
	//装饰器加一些代码
	limited, err := s.limiter.Limit(ctx, "sms:tencent")
	if err != nil {
		//系统错误
		//可以限流：保守策略：微信系统差

		//可以不限：容错策略下游很强，业务可用性要求很高
		return fmt.Errorf("短信服务判断是否限流出现问题 %w", err)
	}
	if limited {
		return errLimited
	}
	//装饰器加一些代码
	err = s.svc.Send(ctx, tpl, args, numbers...)
	//装饰器加一些代码
	return err
}

package ratelimit

import (
	"awesomeProject/webook/internal/service/sms"
	"awesomeProject/webook/pkg/ratelimit"
)

//
//var errLimited = fmt.Errorf("触发了限流")

type RateLimitSMSServiceV1 struct {
	//被装饰的东西
	sms.Service
	limiter ratelimit.Limiter
}

func NewRateLimitSMSServiceV1(svc sms.Service, limiter ratelimit.Limiter) sms.Service {
	return &RateLimitSMSService{
		svc:     svc,
		limiter: limiter,
	}

}

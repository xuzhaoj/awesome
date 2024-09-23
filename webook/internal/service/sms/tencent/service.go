package tencent

import (
	"awesomeProject/webook/pkg/ratelimit"
	"context"
	"fmt"
	"github.com/ecodeclub/ekit"
	"github.com/ecodeclub/ekit/slice"
	sms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"
)

type Service struct {
	appId    *string
	signName *string
	client   *sms.Client
	limiter  ratelimit.Limiter
}

// 构造函数
func NewService(client *sms.Client, appId string, signName string, limiter ratelimit.Limiter) *Service {
	return &Service{
		client: client,
		//将传入的值转化为指针类型
		appId:    ekit.ToPtr[string](appId),
		signName: ekit.ToPtr[string](signName),
		limiter:  limiter,
	}

}
func (s *Service) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {

	req := sms.NewSendSmsRequest()
	req.SmsSdkAppId = s.appId
	req.SignName = s.signName
	req.TemplateId = ekit.ToPtr[string](tplId)
	req.PhoneNumberSet = s.toStringPtrSlice(numbers)
	req.TemplateParamSet = s.toStringPtrSlice(args)
	resp, err := s.client.SendSms(req)
	if err != nil {
		return err
	}
	for _, status := range resp.Response.SendStatusSet {
		if status.Code == nil || *(status.Code) != "Ok" {
			return fmt.Errorf("发送失败, code: %s, 原因: %s", *status.Code, *status.Message)
		}
	}
	return nil

}

func (s *Service) toStringPtrSlice(src []string) []*string {
	return slice.Map[string, *string](src, func(idx int, src string) *string {
		return &src
	})
}

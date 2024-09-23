package failover

import (
	"awesomeProject/webook/internal/service/sms"
	"context"
	"errors"
	"log"
)

type FailoverSMSService struct {
	svcs []sms.Service
}

// 调用者返回的还是struct结构体，只是说里面特定的实现了send接口方法
func NewFailoverSMSService(svcs []sms.Service) sms.Service {
	return &FailoverSMSService{
		svcs: svcs,
	}
}

// 接口方法的实现
func (f *FailoverSMSService) Send(ctx context.Context, tpl string, args []string, numbers ...string) error {
	//使用循环对发送的短信请求挨个去轮询短信服务商，
	for _, svc := range f.svcs {
		err := svc.Send(ctx, tpl, args, numbers...)
		//发送成功，就直接退出出去不在循环检查了
		if err == nil {
			return nil
		}
		//发送失败，切换下一个服务商
		log.Println(err)
	}
	return errors.New("全部服务商都失败了")

}

package sms

import "context"

type Service interface {
	//传递零个或者多个参数
	Send(ctx context.Context, tpl string, args []string, numbers ...string) error
}

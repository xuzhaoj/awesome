package memory

import (
	"context"
	"fmt"
)

type Service struct {
}

func NewService() *Service {
	return &Service{}
}

// 模拟发送短信的过程
func (s *Service) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	fmt.Println(args)
	return nil
}

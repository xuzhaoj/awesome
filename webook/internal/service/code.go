package service

import (
	"awesomeProject/webook/internal/repository"
	"awesomeProject/webook/internal/service/sms"
	"context"
	"fmt"
	"math/rand"
)

const codeTpId = "1877556"

var (
	ErrCodeSendTooMany        = repository.ErrCodeSendTooMany
	ErrCodeVerifyTooManyTimes = repository.ErrCodeVerifyTooManyTimes
)

type CodeService interface {
	Send(ctx context.Context,
		biz string,
		phone string) error
	Verify(ctx context.Context, biz string,
		phone string, inputCode string) (bool, error)
}

type codeService struct {
	//包中的结构体
	repo   repository.CodeRepository
	smsSvc sms.Service
}

func NewCodeService(repo repository.CodeRepository, smsSvc sms.Service) CodeService {
	return &codeService{repo: repo, smsSvc: smsSvc}
}

//1.发送验证码

func (svc *codeService) Send(ctx context.Context,
	biz string,
	phone string) error {

	//生成验证码，塞进redis发送出去
	code := svc.generateCode()
	//存储验证码
	err := svc.repo.Store(ctx, biz, phone, code)
	if err != nil {
		return err
	}
	//发送出去
	err = svc.smsSvc.Send(ctx, codeTpId, []string{code}, phone)
	return err
}

// 生成任意六位数字的验证码
func (svc *codeService) generateCode() string {
	//左闭右开
	num := rand.Intn(1000000)
	return fmt.Sprintf("%06d", num)
}

// 2.校验验证码
func (svc *codeService) Verify(ctx context.Context, biz string,
	phone string, inputCode string) (bool, error) {
	//可以只返回error，有错误代表系统或者验证码验证出错，没用则代表校验通过
	return svc.repo.Verify(ctx, biz, phone, inputCode)
}

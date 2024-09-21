package ioc

import (
	"awesomeProject/webook/internal/service/sms"
	"awesomeProject/webook/internal/service/sms/memory"
)

func InitSMSService() sms.Service {
	return memory.NewService()
}

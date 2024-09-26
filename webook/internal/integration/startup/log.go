package startup

import (
	"awesomeProject/webook/pkg/logger"
)

func InitLogger() logger.LoggerV1 {
	return &logger.NopLogger{}
}

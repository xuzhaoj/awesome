package ioc

import (
	"awesomeProject/webook/internal/repository/dao"
	"awesomeProject/webook/pkg/logger"
	"fmt"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	glogger "gorm.io/gorm/logger"
	"time"
)

func InitDB(l logger.LoggerV1) *gorm.DB {
	dsn := viper.GetString("db.mysql.dsn")
	fmt.Println(dsn)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: glogger.New(gormLoggerFunc(l.Debug), glogger.Config{
			SlowThreshold:             time.Millisecond * 10,
			IgnoreRecordNotFoundError: true,
			LogLevel:                  glogger.Info,
		}),
	})
	if err != nil {
		//只在初始化过程,相当于goroutine结束,出错就不启动了
		panic(err)
	}
	//见表也不成功
	err = dao.InitTable(db)
	if err != nil {
		panic(err)
	}
	return db

}

type gormLoggerFunc func(msg string, fields ...logger.Field)

func (g gormLoggerFunc) Printf(msg string, args ...interface{}) {
	g(msg, logger.Field{Key: "args", Val: args})

}

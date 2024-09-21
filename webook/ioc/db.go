package ioc

import (
	"awesomeProject/webook/config"
	"awesomeProject/webook/internal/repository/dao"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func InitDB() *gorm.DB {

	db, err := gorm.Open(mysql.Open(config.Config.DB.DSN))
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

//go:build wireinject

package main

import (
	"awesomeProject/webook/internal/repository"
	"awesomeProject/webook/internal/repository/cache"
	"awesomeProject/webook/internal/repository/dao"
	"awesomeProject/webook/internal/service"
	"awesomeProject/webook/internal/web"
	"awesomeProject/webook/ioc"
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

func InitWebServer() *gin.Engine {
	//自动生成依赖注入
	wire.Build(ioc.InitDB, ioc.InitRedis,
		dao.NewUserDao,
		cache.NewUserCache,
		cache.NewCodeCache,
		repository.NewUserRepository,
		repository.NewCodeRepository,
		service.NewUserService,
		service.NewCodeService,
		//给予内存实现对应不上
		ioc.InitSMSService,
		web.NewUserHandler,
		//中间件，注册路由呢
		ioc.InitWebServer,
		ioc.InitMiddlewares,
	)

	//分配内存，返回一个gin.engine类型的指针
	return gin.Default()

}

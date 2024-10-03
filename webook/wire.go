//go:build wireinject

package main

import (
	"awesomeProject/webook/internal/events/article"
	"awesomeProject/webook/internal/repository"
	article2 "awesomeProject/webook/internal/repository/article"
	"awesomeProject/webook/internal/repository/cache"
	"awesomeProject/webook/internal/repository/dao"
	article3 "awesomeProject/webook/internal/repository/dao/article"
	"awesomeProject/webook/internal/service"
	"awesomeProject/webook/internal/web"
	"awesomeProject/webook/ioc"
	"github.com/google/wire"
)

func InitWebServer() *App {
	//自动生成依赖注入
	wire.Build(ioc.InitDB, ioc.InitRedis,
		dao.NewUserDao,
		dao.NewGORMInteractiveDAO,
		repository.NewCachedInteractiveRepository,
		service.NewArticleService,
		article3.NewGORMArticleDao,
		cache.NewUserCache,
		cache.NewInteractiveRedisCache,
		cache.NewCodeCache,
		repository.NewUserRepository,
		repository.NewCodeRepository,

		service.NewUserService,
		service.NewCodeService,
		service.NewInteractiveService,

		article2.NewArticleRepository,
		article.NewKafkaProducer,

		//给予内存实现对应不上
		ioc.InitSMSService,
		//消费者单个消费者
		//article.NewInteractiveReadEventConsumer,
		article.NewInteractiveReadEventBatchConsumer,
		//article.NewKafkaProducer,
		web.NewUserHandler,
		web.NewArticleHandler,
		//中间件，注册路由呢
		ioc.InitWebServer,
		ioc.InitMiddlewares,
		ioc.InitLogger,
		ioc.InitKafka,
		ioc.NewSyncProducer,
		ioc.NewConsumers,
		wire.Struct(new(App), "*"),
	)

	//分配内存，返回一个gin.engine类型的指针
	return new(App)

}

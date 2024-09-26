//go:build wireinject

package startup

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

var thirdProvider = wire.NewSet( // 第三方依赖
	InitRedis, InitDB,
	InitLogger)

var userSvcProvider = wire.NewSet(
	dao.NewUserDao,
	cache.NewUserCache,
	repository.NewUserRepository,
	service.NewUserService)

//var articlSvcProvider = wire.NewSet(
//	repository.NewCachedArticleRepository,
//	cache.NewArticleRedisCache,
//	dao.NewArticleGORMDAO,
//	service.NewArticleService)

func InitWebServer() *gin.Engine {
	wire.Build(
		thirdProvider,
		userSvcProvider,
		//articlSvcProvider,
		// cache 部分
		cache.NewCodeCache,
		dao.NewGORMArticleDao,
		// repository 部分
		repository.NewCodeRepository,
		repository.NewArticleRepository,

		// Service 部分
		ioc.InitSMSService,
		service.NewCodeService,
		service.NewArticleService,
		//InitWechatService,

		// handler 部分
		web.NewUserHandler,
		web.NewArticleHandler,
		//web.NewOAuth2WechatHandler,
		//ijwt.NewRedisJWTHandler,
		ioc.InitMiddlewares,
		ioc.InitWebServer,
	)
	return gin.Default()
}

// 不需要注入那么多的东西所以直接
func InitArticleHandler() *web.ArticleHandler {
	wire.Build(
		thirdProvider,
		//userSvcProvider,
		//repository.NewCachedArticleRepository,
		//cache.NewArticleRedisCache,
		dao.NewGORMArticleDao,
		service.NewArticleService,
		web.NewArticleHandler,
		repository.NewArticleRepository,
	)
	return &web.ArticleHandler{}
}

func InitUserSvc() service.UserService {
	wire.Build(thirdProvider, userSvcProvider)
	return service.NewUserService(nil, nil)
}

//func InitJwtHdl() ijwt.Handler {
//	wire.Build(thirdProvider, ijwt.NewRedisJWTHandler)
//	return ijwt.NewRedisJWTHandler(nil)
//}

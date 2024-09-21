////go:generate go run -mod=mod github.com/google/wire/cmd/wire
////go:build !wireinject
//// +build !wireinject
//
package main

//
//import (
//	"awesomeProject/webook/internal/repository"
//	"awesomeProject/webook/internal/repository/cache"
//	"awesomeProject/webook/internal/repository/dao"
//	"awesomeProject/webook/internal/service"
//	"awesomeProject/webook/internal/web"
//	"awesomeProject/webook/ioc"
//	"github.com/gin-gonic/gin"
//)
//
//// Injectors from wire.go:
//
//func InitWebServer() *gin.Engine {
//	cmdable := ioc.InitRedis()
//	v := ioc.InitMiddlewares(cmdable)
//	db := ioc.InitDB()
//	userDAO := dao.NewUserDao(db)
//	userCache := cache.NewUserCache(cmdable)
//	userRepository := repository.NewUserRepository(userDAO, userCache)
//	userService := service.NewUserService(userRepository)
//	codeCache := cache.NewCodeCache(cmdable)
//	codeRepository := repository.NewCodeRepository(codeCache)
//	smsService := ioc.InitSMSService()
//	codeService := service.NewCodeService(codeRepository, smsService)
//	userHandler := web.NewUserHandler(userService, codeService)
//	engine := ioc.InitWebServer(v, userHandler)
//	return engine
//}

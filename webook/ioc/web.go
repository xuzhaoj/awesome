package ioc

import (
	"awesomeProject/webook/internal/web"
	"awesomeProject/webook/internal/web/middleware"
	"awesomeProject/webook/pkg/ginx/middlewares/ratelimit"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"strings"
	"time"
)

func InitWebServer(mdls []gin.HandlerFunc, hdl *web.UserHandler) *gin.Engine {
	server := gin.Default()
	//中间件 make定义
	server.Use(mdls...)
	hdl.RegisterRoutes(server)
	return server
}

func InitMiddlewares(redisClient redis.Cmdable) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		corsHdl(),
		middleware.NewLoginJWTMiddlewareBuilder().
			IgnorePaths("/users/signup").
			IgnorePaths("/users/login_sms/code/send").
			IgnorePaths("/users/login_sms").
			IgnorePaths("/users/login").Build(),
		ratelimit.NewBuilder(redisClient, time.Second, 100).Build(),
	}
}

func corsHdl() gin.HandlerFunc {
	return cors.New(cors.Config{
		//AllowOrigins:     []string{"http://localhost:3000"}, ------不写表示默认,来源复杂可以在下面orgin中定义
		//AllowMethods:     []string{"POST"},				   ------
		//业务请求可以带上的头,前端解决跨域
		AllowHeaders: []string{"Authorization,Content-type"},
		//ExposeHeaders:    []string{"authorization,content-type"},
		//是否允许带上cookie之类的东西
		AllowCredentials: true,
		ExposeHeaders:    []string{"x-jwt-token"},
		//定义哪些来源的允许的
		AllowOriginFunc: func(origin string) bool {
			if strings.HasPrefix(origin, "http://localhost") {
				return true
			}
			//来自公司的域名也会被允许
			return strings.Contains(origin, "your company.com")
		},
		MaxAge: 12 * time.Hour,
	})
}

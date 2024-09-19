package web

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"strings"
	"time"
)

func RegisterRoutes() *gin.Engine {
	server := gin.Default()
	u := &UserHandler{}
	//按照顺序执行
	server.Use(func(context *gin.Context) {
		println("这是第一个middleware")
	})
	server.Use(func(context *gin.Context) {
		println("这是第二个middleware")
	})
	//解决跨域问题
	server.Use(cors.New(cors.Config{
		//AllowOrigins:     []string{"http://localhost:3000"},
		//AllowMethods:     []string{"POST"},
		//业务请求可以带上的头
		AllowHeaders: []string{"authorization,content-type"},
		//ExposeHeaders:    []string{"authorization,content-type"},
		//是否允许带上cookie之类的东西
		AllowCredentials: true,
		//定义哪些来源的允许的
		AllowOriginFunc: func(origin string) bool {
			if strings.HasPrefix(origin, "http://localhost") {
				return true
			}
			//来自公司的域名也会被允许
			return strings.Contains(origin, "your company.com")
		},
		MaxAge: 12 * time.Hour,
	}))

	u.RegisterRoutes(server)
	return server
	//server.Run(":8080")

}

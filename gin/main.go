package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func main() {
	server := gin.Default()
	server.GET("/hello", func(c *gin.Context) {
		c.String(http.StatusOK, "hello,go")
	})
	server.POST("/post", func(context *gin.Context) {
		context.String(http.StatusOK, "hello post")
	})
	//参数路由必须要带:
	server.GET("/user/:name", func(context *gin.Context) {
		//:直接提取出来,参数
		name := context.Param("name")
		//context.String(http.StatusOK, "参数路由")
		context.String(http.StatusOK, "参数路由"+name)
	})
	server.GET("/views/*.html", func(context *gin.Context) {
		page := context.Param(".html")
		context.String(http.StatusOK, "hello,这是通配符路由"+page)
	})
	server.GET("/order", func(context *gin.Context) {
		//地址栏后面要跟上?id=xxx
		oid := context.Query("id")
		context.String(http.StatusOK, "hhello,这是查询参数"+oid)
	})
	//路径后面加不加/其实无所谓只要你不跟着参数就可以了
	server.GET("/users/", func(context *gin.Context) {
		context.String(http.StatusOK, "你好")
	})

	server.Run(":8080") // 监听并在 0.0.0.0:8080 上启动服务
}

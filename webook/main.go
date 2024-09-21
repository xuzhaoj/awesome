package main

import (
	"awesomeProject/webook/internal/repository"
	"awesomeProject/webook/internal/repository/cache"
	"awesomeProject/webook/internal/repository/dao"
	"awesomeProject/webook/internal/service"
	"awesomeProject/webook/internal/web"
	"awesomeProject/webook/ioc"
	"net/http"

	//Gin框架API
	"github.com/gin-gonic/gin"
)

func main() {
	//db := initDB()
	//
	server := InitWebServer()
	//
	//rdb := initRedis()
	//
	//u := initUser(db, rdb)
	//
	//u.RegisterRoutes(server)

	//server := InitWebServer()
	//server := gin.Default()
	server.GET("/hello", func(context *gin.Context) {
		context.String(http.StatusOK, "hello world")
	})
	server.Run(":8080")
}

// // wire_go生成的代码直接复制过来就行了因为有bug
func InitWebServer() *gin.Engine {
	cmdable := ioc.InitRedis()
	v := ioc.InitMiddlewares(cmdable)
	db := ioc.InitDB()
	userDAO := dao.NewUserDao(db)
	userCache := cache.NewUserCache(cmdable)
	userRepository := repository.NewUserRepository(userDAO, userCache)
	userService := service.NewUserService(userRepository)
	codeCache := cache.NewCodeCache(cmdable)
	codeRepository := repository.NewCodeRepository(codeCache)
	smsService := ioc.InitSMSService()
	codeService := service.NewCodeService(codeRepository, smsService)
	userHandler := web.NewUserHandler(userService, codeService)
	engine := ioc.InitWebServer(v, userHandler)
	return engine
}

//func initRedis() redis.Cmdable {
//	redisClient := redis.NewClient(&redis.Options{
//		Addr: config.Config.Redis.Addr,
//	})
//	return redisClient
//}
//
//func initWebServer() *gin.Engine {
//	server := gin.Default()
//	//按照顺序执行
//	server.Use(func(context *gin.Context) {
//		println("这是第一个middleware")
//	})
//	server.Use(func(context *gin.Context) {
//		println("这是第二个middleware")
//	})
//	//redisClient := redis.NewClient(&redis.Options{
//	//	Addr: config.Config.Redis.Addr,
//	//})
//	//server.Use(ratelimit.NewBuilder(redisClient, time.Second, 100).Build())
//	//解决跨域问题
//	server.Use(cors.New(cors.Config{
//		//AllowOrigins:     []string{"http://localhost:3000"}, ------不写表示默认,来源复杂可以在下面orgin中定义
//		//AllowMethods:     []string{"POST"},				   ------
//		//业务请求可以带上的头,前端解决跨域
//		AllowHeaders: []string{"Authorization,Content-type"},
//		//ExposeHeaders:    []string{"authorization,content-type"},
//		//是否允许带上cookie之类的东西
//		AllowCredentials: true,
//		ExposeHeaders:    []string{"x-jwt-token"},
//		//定义哪些来源的允许的
//		AllowOriginFunc: func(origin string) bool {
//			if strings.HasPrefix(origin, "http://localhost") {
//				return true
//			}
//			//来自公司的域名也会被允许
//			return strings.Contains(origin, "your company.com")
//		},
//		MaxAge: 12 * time.Hour,
//	}))

//步骤1
//使用cookie作为会话存储，将会话中间件添加到服务器中，mysession是会话名称
//store := cookie.NewStore([]byte("secret"))
//使用基于内存的绘画存储
//store := memstore.NewStore([]byte("8s6Mh0NLjf2P9cDZzLcc9bcQMBKoPwK6"), []byte("iSQXg9EZhWMbSxYkExJaj4zbflnHCppl"))
//store, err := redis.NewStore(16, "tcp", "localhost:6379", "",
//	[]byte("8s6Mh0NLjf2P9cDZzLcc9bcQMBKoPwK6"), []byte("iSQXg9EZhWMbSxYkExJaj4zbflnHCppl"))
//if err != nil {
//	panic(err)
//}

//store := memstore.NewMemStore([]byte("8s6Mh0NLjf2P9cDZzLcc9bcQMBKoPwK6"), []byte("iSQXg9EZhWMbSxYkExJaj4zbflnHCppl"))

//server.Use(sessions.Sessions("mysession", store))
//步骤三  这个校验的功能放在了login.go， 中了
//	server.Use(middleware.NewLoginJWTMiddlewareBuilder().
//		IgnorePaths("/users/signup").
//		IgnorePaths("/users/login_sms/code/send").
//		IgnorePaths("/users/login_sms").
//		IgnorePaths("/users/login").Build())
//
//	return server
//
//}

//func initUser(db *gorm.DB, rdb redis.Cmdable) *web.UserHandler {
//	ud := dao.NewUserDao(db)
//	uc := cache.NewUserCache(rdb)
//	repo := repository.NewUserRepository(ud, uc)
//	svc := service.NewUserService(repo)
//	//u := &web.UserHandler{}
//	codeCache := cache.NewCodeCache(rdb)
//	codeRepo := repository.NewCodeRepository(codeCache)
//	smsSvc := memory.NewService()
//	codeSvc := service.NewCodeService(codeRepo, smsSvc)
//	u := web.NewUserHandler(svc, codeSvc)
//	return u
//}

//func initDB() *gorm.DB {
//
//	db, err := gorm.Open(mysql.Open(config.Config.DB.DSN))
//	if err != nil {
//		//只在初始化过程,相当于goroutine结束,出错就不启动了
//		panic(err)
//	}
//	//见表也不成功
//	err = dao.InitTable(db)
//	if err != nil {
//		panic(err)
//	}
//	return db
//
//}

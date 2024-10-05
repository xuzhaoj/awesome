package main

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"net/http"
	"time"
)

func main() {
	//db := initDB()
	//
	initViper()
	initLogger()
	initPrometheus()
	app := InitWebServer()
	app.cron.Start()
	server := app.web
	//遍历数组和切片会返回每个元素的索引和元素值
	//遍历map会返回每个键和值
	//遍历channel会接受channel中的值直到关闭为止
	//不需要索引值的时候可以用下划线忽略
	for _, c := range app.consumers {
		err := c.Start()
		if err != nil {
			panic(err)
		}
	}
	//
	//rdb := initRedis()
	//
	//u := initUser(db, rdb)
	//
	//u.RegisterRoutes(server)

	//server := InitWebServer()
	//server := gin.Default()
	//server := app.web
	server.GET("/hello", func(context *gin.Context) {
		context.String(http.StatusOK, "hello world")
	})
	server.Run(":8080")
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	ctx = app.cron.Stop()
	tm := time.NewTimer(time.Minute * 10)
	select {
	case <-tm.C:
	case <-ctx.Done():
	}
}

//	func InitWebServer() *gin.Engine {
//		cmdable := ioc.InitRedis()
//		loggerV1 := ioc.InitLogger()
//		v := ioc.InitMiddlewares(cmdable, loggerV1)
//		db := ioc.InitDB(loggerV1)
//		userDAO := dao.NewUserDao(db)
//		userCache := cache.NewUserCache(cmdable)
//		userRepository := repository.NewUserRepository(userDAO, userCache)
//		userService := service.NewUserService(userRepository, loggerV1)
//		codeCache := cache.NewCodeCache(cmdable)
//		codeRepository := repository.NewCodeRepository(codeCache)
//		smsService := ioc.InitSMSService()
//		codeService := service.NewCodeService(codeRepository, smsService)
//		userHandler := web.NewUserHandler(userService, codeService)
//		engine := ioc.InitWebServer(v, userHandler)
//		return engine
//	}
func initPrometheus() {
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":8081", nil)
	}()
}

func initLogger() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	zap.L().Info("这是replace之前")
	zap.ReplaceGlobals(logger)
	zap.L().Info("hello，你初始化log成功")
	type Demo struct {
		Name string `json:"name"`
	}
	zap.L().Info("这是实验参数", zap.Error(errors.New("这是一个error")),
		zap.Int64("id", 123),
		zap.Any("一个结构体", Demo{Name: "hello"}))

}
func initViper() {
	viper.SetDefault("db.mysql.dsn", "root:root@tcp(webook-live-mysql:11309)/webook")
	viper.SetConfigName("dev")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./webook/config")

	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
	//可以有多个viper实例对象
	//otherViper := viper.New()

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

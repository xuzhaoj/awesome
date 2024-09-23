package integration

import (
	"awesomeProject/webook/internal/repository"
	"awesomeProject/webook/internal/repository/cache"
	"awesomeProject/webook/internal/repository/dao"
	"awesomeProject/webook/internal/service"
	"awesomeProject/webook/internal/web"
	"awesomeProject/webook/ioc"
	"bytes"
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestUserHandler_e2e_SendLoginSMSCode(t *testing.T) {
	server := InitWebServer()
	rdb := ioc.InitRedis()
	testCases := []struct {
		name string
		//准备数据
		before func(t *testing.T)

		//验证数据
		after    func(t *testing.T)
		reqBody  string
		wantCode int
		wantBody web.Result
	}{
		{
			name: "发送成功",
			before: func(t *testing.T) {

			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				val, err := rdb.GetDel(ctx, "phone_code:login:13328703332").Result()
				cancel()
				assert.NoError(t, err)
				assert.True(t, len(val) == 6)
			},
			reqBody: `{
				"phone": "13328703332"
	
}`,
			wantCode: 200,
			wantBody: web.Result{
				Msg: "发送验证码成功",
			},
		},
		{
			name: "发送太频繁",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				_, err := rdb.Set(ctx, "phone_code:login:13328703332", "123456",
					time.Minute*9+time.Second*30).Result()
				cancel()
				assert.NoError(t, err)

			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				val, err := rdb.GetDel(ctx, "phone_code:login:13328703332").Result()
				cancel()
				assert.NoError(t, err)
				assert.Equal(t, "123456", val)
			},
			reqBody: `{
				"phone": "13328703332"
	
}`,
			wantCode: 200,
			wantBody: web.Result{
				Msg: "发送频繁，请稍后再试",
			},
		},
		{
			name: "系统错误",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				//有验证码没有过期时间
				_, err := rdb.Set(ctx, "phone_code:login:13328703332", "123456", 0).Result()
				cancel()
				assert.NoError(t, err)

			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				val, err := rdb.GetDel(ctx, "phone_code:login:13328703332").Result()
				cancel()
				assert.NoError(t, err)
				assert.Equal(t, "123456", val)
			},
			reqBody: `{
				"phone": "13328703332"
	
}`,
			wantCode: 200,
			wantBody: web.Result{
				Code: 5,
				Msg:  "系统错误",
			},
		},
		{
			name: "手机代码为空",
			before: func(t *testing.T) {

			},
			after: func(t *testing.T) {

			},
			reqBody: `{
				"phone": ""
	
}`,
			wantCode: 200,
			wantBody: web.Result{
				Code: 4,
				Msg:  "输入的手机号有误请重新输入",
			},
		},
		{
			name: "数据格式有误",
			before: func(t *testing.T) {

			},
			after: func(t *testing.T) {

			},
			reqBody: `{
				"phone": ,
	
}`,
			wantCode: 400,
			//wantBody: web.Result{
			//	Code: 4,
			//	Msg:  "输入的手机号有误请重新输入",
			//},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			req, err := http.NewRequest(http.MethodPost,
				"/users/login_sms/code/send", bytes.NewBuffer([]byte(tc.reqBody)))
			//踩坑---请求的数据是json
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			//构造的请求或者路径不对会返回err

			//返回数据存储的地方
			resp := httptest.NewRecorder()

			//HTTP进入gin框架的入口
			server.ServeHTTP(resp, req)

			assert.Equal(t, tc.wantCode, resp.Code)
			if resp.Code != 200 {
				return
			}
			var webRes web.Result
			err = json.NewDecoder(resp.Body).Decode(&webRes)
			require.NoError(t, err)
			assert.Equal(t, tc.wantBody, webRes)
			tc.after(t)

		})
	}

}
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

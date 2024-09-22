package web

import (
	"awesomeProject/webook/internal/domain"
	"awesomeProject/webook/internal/service"
	svcmocks "awesomeProject/webook/internal/service/mocks"
	"bytes"
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUserHandler_SingUp(t *testing.T) {

	//结构体匿名对象数组
	testCases := []struct {
		name     string
		mock     func(ctrl *gomock.Controller) service.UserService
		reqBody  string
		wantCode int
		wantBody string
	}{

		{
			name: "注册成功",
			mock: func(ctrl *gomock.Controller) service.UserService {

				userSvc := svcmocks.NewMockUserService(ctrl)
				userSvc.EXPECT().SignUp(gomock.Any(), domain.User{
					Email:    "123@qq.com",
					Password: "123456789",
				}).Return(nil)
				return userSvc

			},

			reqBody: `
{
	"email": "123@qq.com",
	"password": "123456789",
	"confirmPassword": "123456789"

}
`,
			wantCode: http.StatusOK,
			wantBody: "注册成功",
		},

		{
			name: "参数不对，bind失败",
			mock: func(ctrl *gomock.Controller) service.UserService {

				userSvc := svcmocks.NewMockUserService(ctrl)

				return userSvc

			},

			reqBody: `
{
	"email": "123@qq.com",
	"password": "123456789",
 
}
`,
			wantCode: http.StatusBadRequest,
		},

		{
			name: "邮箱格式不对",
			mock: func(ctrl *gomock.Controller) service.UserService {

				userSvc := svcmocks.NewMockUserService(ctrl)

				return userSvc

			},

			reqBody: `
{
	"email": "123@2",
	"password": "123456789",
	"confirmPassword": "123456789"

}
`,
			wantCode: http.StatusOK,
			wantBody: "你的邮箱格式不对",
		},

		{
			name: "两次密码不匹配",
			mock: func(ctrl *gomock.Controller) service.UserService {

				userSvc := svcmocks.NewMockUserService(ctrl)
				//userSvc.EXPECT().SignUp(gomock.Any(), domain.User{
				//	Email:    "123@qq.com",
				//	Password: "123456789",
				//}).Return(nil)
				return userSvc

			},

			reqBody: `
{
	"email": "123@qq.com",
	"password": "1234567ss89",
	"confirmPassword": "123456789"

}
`,
			wantCode: http.StatusOK,
			wantBody: "两次输入的密码不一致,请重新输入",
		},

		{
			name: "密码的输入格式不对",
			mock: func(ctrl *gomock.Controller) service.UserService {

				userSvc := svcmocks.NewMockUserService(ctrl)

				return userSvc

			},

			reqBody: `
{
	"email": "123@qq.com",
	"password": "12345#6789",
	"confirmPassword": "12345#6789"

}
`,
			wantCode: http.StatusOK,
			wantBody: "你的密码不对,必须大于八位包含特殊字符",
		},
		{
			name: "邮箱冲突",
			mock: func(ctrl *gomock.Controller) service.UserService {

				userSvc := svcmocks.NewMockUserService(ctrl)
				userSvc.EXPECT().SignUp(gomock.Any(), domain.User{
					Email:    "123@qq.com",
					Password: "123456789",
				}).Return(service.ErrUserDuplicate)
				return userSvc

			},

			reqBody: `
{
	"email": "123@qq.com",
	"password": "123456789",
	"confirmPassword": "123456789"

}
`,
			wantCode: http.StatusOK,
			wantBody: "邮箱冲突，请使用另外一个",
		},
		{
			name: "系统异常",
			mock: func(ctrl *gomock.Controller) service.UserService {

				userSvc := svcmocks.NewMockUserService(ctrl)
				userSvc.EXPECT().SignUp(gomock.Any(), domain.User{
					Email:    "123@qq.com",
					Password: "123456789",
				}).Return(errors.New("随便一个ERRor"))
				return userSvc

			},

			reqBody: `
{
	"email": "123@qq.com",
	"password": "123456789",
	"confirmPassword": "123456789"

}
`,
			wantCode: http.StatusOK,
			wantBody: "系统错误",
		},
	}

	//每个测试用例都会通过 t.Run 创建一个子测试，t.Run 的第一个参数是子测试的名称，
	//这里使用的是 tc.name，表示每个测试用例的名称。第二个参数是一个匿名函数，里面可以包含该测试用例的具体测试逻辑。
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			server := gin.Default()

			h := NewUserHandler(tc.mock(ctrl), nil)
			h.RegisterRoutes(server)

			req, err := http.NewRequest(http.MethodPost,
				"/users/signup", bytes.NewBuffer([]byte(tc.reqBody)))
			//踩坑---请求的数据是json

			req.Header.Set("Content-Type", "application/json")

			//构造的请求或者路径不对会返回err
			require.NoError(t, err)
			//返回数据存储的地方
			resp := httptest.NewRecorder()

			//HTTP进入gin框架的入口
			server.ServeHTTP(resp, req)

			assert.Equal(t, tc.wantCode, resp.Code)
			assert.Equal(t, tc.wantBody, resp.Body.String())

		})
	}
}

func TestMock(t *testing.T) {
	// 先创建一个控制 mock 的控制器
	ctrl := gomock.NewController(t)
	// 每个测试结束都要调用 Finish,
	// 然后 mock 就会验证你的测试流程是否符合预期
	defer ctrl.Finish()

	usersvc := svcmocks.NewMockUserService(ctrl)

	// 开始设计一个模拟调用
	// 预期第一个是 Signup 的调用
	// 模拟的条件是 gomock.Any, gomock.Any.
	// 然后返回,随便传递参数

	usersvc.EXPECT().SignUp(gomock.Any(), gomock.Any()).
		Return(errors.New("模拟的错误"))

	err := usersvc.SignUp(context.Background(), domain.User{
		Email: "123@qq.com",
	})
	t.Log(err)

}

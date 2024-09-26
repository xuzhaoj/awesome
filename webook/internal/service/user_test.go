package service

import (
	"awesomeProject/webook/internal/domain"
	"awesomeProject/webook/internal/repository"
	repomocks "awesomeProject/webook/internal/repository/mocks"
	"awesomeProject/webook/pkg/logger"
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
	"testing"
	"time"
)

func Test_userService_Login(t *testing.T) {
	now := time.Now()
	testCases := []struct {
		name string
		mock func(ctrl *gomock.Controller) repository.UserRepository

		//ctx context.Context
		u domain.User

		wantUser domain.User
		wantErr  error
	}{
		{
			name: "登录成功",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				//传递的是一个domain对象
				repo.EXPECT().FindByEmail(gomock.Any(), domain.User{
					Email:    "123@qq.com",
					Password: "helloworld123",
				}).
					Return(domain.User{
						Email:    "123@qq.com",
						Password: "$2a$10$I7h/5HE0AkGF0SASAMHT7u9td/yBEbutk5OIQE1nhOcZfIXROGT8O",
						Phone:    "13328703332",
						Ctime:    now,
					}, nil)
				return repo
			},
			u: domain.User{
				Email:    "123@qq.com",
				Password: "helloworld123",
			},
			wantUser: domain.User{
				Email:    "123@qq.com",
				Password: "$2a$10$I7h/5HE0AkGF0SASAMHT7u9td/yBEbutk5OIQE1nhOcZfIXROGT8O",
				Phone:    "13328703332",
				Ctime:    now,
			},
			wantErr: nil,
		},
		{
			name: "用户不存在",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				//传递的是一个domain对象
				repo.EXPECT().FindByEmail(gomock.Any(), domain.User{
					Email:    "123@qq.com",
					Password: "helloworld123",
				}).
					Return(domain.User{}, repository.ErrUserNotFound)
				return repo
			},
			u: domain.User{
				Email:    "123@qq.com",
				Password: "helloworld123",
			},
			wantUser: domain.User{},
			//上面的return是返回的错误信息，下面的这个是比较错误后在返回的错误
			wantErr: ErrInvalidUserOrPassword,
		},
		{
			name: "DB错误",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				//传递的是一个domain对象
				repo.EXPECT().FindByEmail(gomock.Any(), domain.User{
					Email:    "123@qq.com",
					Password: "helloworld123",
				}).
					Return(domain.User{}, errors.New("乱七八糟的错误"))
				return repo
			},
			u: domain.User{
				Email:    "123@qq.com",
				Password: "helloworld123",
			},
			wantUser: domain.User{},
			wantErr:  errors.New("乱七八糟的错误"),
		},
		{
			name: "密码错误",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				//传递的是一个domain对象
				repo.EXPECT().FindByEmail(gomock.Any(), domain.User{
					Email:    "123@qq.com",
					Password: "hellowosssrld123",
				}).
					Return(domain.User{
						Email:    "123@qq.com",
						Password: "$2a$10$I7h/5HE0AkGF0SASAMHT7u9td/yBEbutk5OIQE1nhOcZfIXROGT8O",
						Phone:    "13328703332",
						Ctime:    now,
					}, nil)
				return repo
			},
			u: domain.User{
				Email:    "123@qq.com",
				Password: "hellowosssrld123",
			},
			wantUser: domain.User{},
			wantErr:  ErrInvalidUserOrPassword,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			svc := NewUserService(tc.mock(ctrl), &logger.NopLogger{})
			user, err := svc.Login(context.Background(), tc.u)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantUser, user)
		})
	}

}

func TestEncrypted(t *testing.T) {
	res, err := bcrypt.GenerateFromPassword([]byte("helloworld123"), bcrypt.DefaultCost)
	if err == nil {
		t.Log(string(res))
	}
}

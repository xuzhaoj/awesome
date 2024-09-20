package service

import (
	"awesomeProject/webook/internal/domain"
	"awesomeProject/webook/internal/repository"
	"context"
	"errors"
	"golang.org/x/crypto/bcrypt"
)

var ErrUserDuplicate = repository.ErrUserDuplicate
var ErrInvalidUserOrPassword = errors.New("帐号/或者邮箱密码不对请重新登录")

type UserService struct {

	//对象用指针方便
	repo *repository.UserRepository
	//redis *redis.Client,,,,,,,涉及到数据层面的操作的时候可以丢费repository去做
}

// 返回这个service的对象
func NewUserService(repo *repository.UserRepository) *UserService {
	//现场定义结构体的内容
	return &UserService{repo: repo}

}

func (svc *UserService) SignUp(ctx context.Context, u domain.User) error {
	//用指针还需要去排空,所以不去用User的指针,不可以调用上层定义好的结构体,只允许用下层定义好的
	//思考加密存放在哪里
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hash)
	//然后就是存起来
	return svc.repo.Create(ctx, u)
	//if err != nil {
	//	return err
	//}
	//
	//return err
}

func (svc *UserService) Login(ctx context.Context, u domain.User) (domain.User, error) {
	//u是传递下来的,user是查询后的结果,两者进行密码的比较
	//登录的逻辑,根据邮箱找到数据库用户信息

	user, err := svc.repo.FindByEmail(ctx, u)
	if err == repository.ErrUserNotFound {
		//返回的是空的user对象
		return domain.User{}, ErrInvalidUserOrPassword

	}
	if err != nil {
		return domain.User{}, err
	}
	//比较加密的密码了
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(u.Password))
	if err != nil {
		//debug
		//返回的是空的user对象
		return domain.User{}, ErrInvalidUserOrPassword
	}
	//nil表示没有错误的返回
	return user, nil
}

func (svc *UserService) Profile(ctx context.Context, id int64) (domain.User, error) {
	user, err := svc.repo.FindById(ctx, id)
	return user, err
	//redis找不到从数据库中进行查找

}

func (svc *UserService) FindOrCreate(ctx context.Context,
	phone string) (domain.User, error) {
	//手机号查询用户存在与否
	u, err := svc.repo.FindByPhone(ctx, phone)

	//存在
	if err != repository.ErrUserNotFound {
		return u, err
	}

	//不存在，插入新用户
	u = domain.User{
		Phone: phone,
	}
	err = svc.repo.Create(ctx, u)
	//存在错误,错误的返回是变量类型的就只返回变量类型
	if err != nil && err != repository.ErrUserDuplicate {
		return u, err
	}
	//逻辑就是创建完成后在查询一次
	return svc.repo.FindByPhone(ctx, phone)

}

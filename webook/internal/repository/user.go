package repository

import (
	"awesomeProject/webook/internal/domain"
	"awesomeProject/webook/internal/repository/cache"
	"awesomeProject/webook/internal/repository/dao"
	"context"
)

var (
	ErrUserDuplicateEmail = dao.ErrUserDuplicateEmail
	ErrUserNotFound       = dao.ErrUserNotFound
)

//var ErrUserDuplicateEmailV1 = fmt.Errorf("%w 邮箱冲突", dao.ErrUserDuplicateEmail)

// 指针指向的是这个结构体，然后去调用作用在结构体上面的方法
type UserRepository struct {
	dao   *dao.UserDAO
	cache *cache.UserCache
}

func NewUserRepository(dao *dao.UserDAO, cache *cache.UserCache) *UserRepository {
	return &UserRepository{
		dao:   dao,
		cache: cache,
	}

}

func (r *UserRepository) Create(ctx context.Context, u domain.User) error {
	return r.dao.Insert(ctx, dao.User{
		Email:    u.Email,
		Password: u.Password,
	})

}

func (r *UserRepository) FindById(ctx context.Context, id int64) (domain.User, error) {

	//先从cache里面找
	u, err := r.cache.Get(ctx, id)
	if err != nil {
		return u, err
	}
	//查询数据库
	u1, err := r.dao.FindById(ctx, id)
	if err != nil {
		return domain.User{}, err
	}
	u = domain.User{
		Id:       u1.Id,
		Email:    u1.Email,
		Password: u1.Password,
	}
	//查完数据库用户信息还是应该放进redis中
	err = r.cache.Set(ctx, u)
	if err != nil {
		//打印日志做监控
	}

	return u, err

	//1.缓存里面有数据

	//2.缓存里面没有数据
	//3.redis崩了

	//再从dao里面找
	//找到了回写cache
}

// 根据邮箱返回用户和错误信息
func (r *UserRepository) FindByEmail(ctx context.Context, u domain.User) (domain.User, error) {
	email := u.Email
	user, err := r.dao.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}
	return domain.User{
		Id:       user.Id,
		Email:    user.Email,
		Password: user.Password,
	}, nil
}

package repository

import (
	"awesomeProject/webook/internal/domain"
	"awesomeProject/webook/internal/repository/cache"
	"awesomeProject/webook/internal/repository/dao"
	"context"
	"database/sql"
	"time"
)

var (
	ErrUserDuplicate = dao.ErrUserDuplicate
	ErrUserNotFound  = dao.ErrUserNotFound
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

// 查询出来的都是数据库的结果
func (r *UserRepository) Create(ctx context.Context, u domain.User) error {
	return r.dao.Insert(ctx, r.domainToEntity(u))

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
	//领域用户对象，由于notnull使用用新的方法生成进行复用
	//u = domain.User{
	//	Id:       u1.Id,
	//	Email:    u1.Email,
	//	Password: u1.Password,
	//}
	u = r.entityToDomain(u1)
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
	//查询出来的是数据库的用户需要转化成domain
	user, err := r.dao.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}
	return r.entityToDomain(user), nil
}

func (r *UserRepository) FindByPhone(ctx context.Context, phone string) (domain.User, error) {
	user, err := r.dao.FindByPhone(ctx, phone)
	if err != nil {
		return domain.User{}, err
	}
	return r.entityToDomain(user), nil
}

func (r *UserRepository) domainToEntity(u domain.User) dao.User {
	//领域用户类与数据库用户类进行转化
	return dao.User{
		Id: u.Id,
		Email: sql.NullString{
			String: u.Email,
			Valid:  u.Email != "",
		},
		Password: u.Password,
		Phone: sql.NullString{
			String: u.Phone,
			Valid:  u.Phone != "",
		},
		Ctime: u.Ctime.UnixMilli(),
	}

}

func (r *UserRepository) entityToDomain(u dao.User) domain.User {
	return domain.User{
		Id:       u.Id,
		Email:    u.Email.String,
		Password: u.Password,
		Phone:    u.Phone.String,
		Ctime:    time.UnixMilli(u.Ctime),
	}
}

package dao

import (
	"context"
	"database/sql"
	"errors"
	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
	"time"
)

var (
	ErrUserDuplicate = errors.New("邮箱冲突")
	ErrUserNotFound  = gorm.ErrRecordNotFound
)

type UserDAO struct {
	//使用了Gorm中的对他进行处理
	db *gorm.DB
}

func NewUserDao(db *gorm.DB) *UserDAO {
	return &UserDAO{
		db: db,
	}
}

func (dao *UserDAO) Insert(ctx context.Context, u User) error {
	//毫秒数
	now := time.Now().UnixMilli()
	u.Ctime = now
	u.Utime = now
	err := dao.db.WithContext(ctx).Create(&u).Error
	if mysqlErr, ok := err.(*mysql.MySQLError); ok {
		const uniqueConflictsErrNo uint16 = 1062
		if mysqlErr.Number == 1062 {
			//邮箱冲突or手机号码冲突
			return ErrUserDuplicate

		}
	}
	return err

}

func (dao *UserDAO) FindByEmail(ctx context.Context, email string) (User, error) {
	//用于存储查询结果
	var u User
	//First 方法来获取第一条匹配的记录，并将结果存储到 u 变量中,GORM ，没有错误信息，赋值过程在返回错误 nil表明查询准确无物
	err := dao.db.WithContext(ctx).Where("email = ?", email).First(&u).Error
	//err := dao.db.WithContext(ctx).First(&u, "email = ?", email).Error
	return u, err
}

func (dao *UserDAO) FindByPhone(ctx context.Context, phone string) (User, error) {
	//用于存储查询结果
	var u User
	//First 方法来获取第一条匹配的记录，并将结果存储到 u 变量中,GORM ，没有错误信息，赋值过程在返回错误 nil表明查询准确无物
	err := dao.db.WithContext(ctx).Where("phone = ?", phone).First(&u).Error
	//err := dao.db.WithContext(ctx).First(&u, "email = ?", email).Error
	return u, err
}
func (dao *UserDAO) FindById(ctx context.Context, id int64) (User, error) {
	//用于存储查询结果
	var u User
	//First 方法来获取第一条匹配的记录，并将结果存储到 u 变量中,GORM ，没有错误信息，赋值过程在返回错误 nil表明查询准确无物
	err := dao.db.WithContext(ctx).Where("id = ?", id).First(&u).Error

	return u, err
}

// domain.user是业务概念,不一定和数据库中的表一一对应,但是dao.User就是一一对应
// 直接对应数据库表结构一一对应  也称之为entity,po,model
type User struct {
	Id int64 `gorm:"primaryKey,autoIncrement"`
	//如果不提供则返回null而不是""
	Email    sql.NullString `gorm:"unique"`
	Password string
	//如果不提供则返回null而不是"",定义唯一索引的时候才需要这样
	Phone sql.NullString `gorm:"unique"`
	Ctime int64

	Utime int64
}

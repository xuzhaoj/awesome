package dao

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

// 这个是一张点赞新表用来记录id 文章 文章id 用户id 点赞状态 更新时间 插入时间-----------和用户绑定的硬软删除
type UserLikeBiz struct {
	Id  int64 `gorm:"primaryKey,autoIncrement"`
	Uid int64 `gorm:"uniqueIndex:uid_biz_type_id"`
	//在前端展示的时候where uid=？and biz_id=? and biz=?高频查询索引需要创建联合索引
	BizId int64  `gorm:"uniqueIndex:uid_biz_type_id"`
	Biz   string `gorm:"type:varchar(128);uniqueIndex:uid_biz_type_id"`
	//如果删掉点赞数后，会造成数据库空行对性能就行影响开始软删除    0代表删除，1代表有效
	Status int
	Utime  int64
	Ctime  int64
}

// 收藏夹，记录了用户收藏了文章记录，收藏数并不在这个表中逻辑就是这样
type UserCollectionBiz struct {
	Id int64 `gorm:"primaryKey,autoIncrement"`
	// 这边还是保留了了唯一索引
	Uid   int64  `gorm:"uniqueIndex:uid_biz_type_id"`
	BizId int64  `gorm:"uniqueIndex:uid_biz_type_id"`
	Biz   string `gorm:"type:varchar(128);uniqueIndex:uid_biz_type_id"`
	// 收藏夹的ID
	//收藏夹的id 其实就是默认收藏夹、后端开发收藏夹，前端开发收藏夹  前端会返回对应的id
	Cid   int64 `gorm:"index"`
	Utime int64
	Ctime int64
}
type InteractiveV1 struct {
	Id int64 `gorm:"primaryKey,autoIncrement"`
	//业务标识符号//创建联合索引，第一个条件查询最左前追的原则，第二个条件区分度，排序。，
	BizId int64  `gorm:"uniqueIndex:biz_type_id"`
	Biz   string `gorm:"type:varchar(128);uniqueIndex:biz_type_id"`

	Cnt int64
	//点赞数/阅读数/收藏数
	CntType string
	Utime   int64
	Ctime   int64
}

// 用来记录文章总的点赞数和阅读数和收藏数，，，，，，，记录文章总的点赞收藏阅读计数
type Interactive struct {
	Id int64 `gorm:"primaryKey,autoIncrement"`
	//业务标识符号//创建联合索引，第一个条件查询最左前追的原则，第二个条件区分度，排序。，
	BizId int64  `gorm:"uniqueIndex:biz_type_id"`
	Biz   string `gorm:"type:varchar(128);uniqueIndex:biz_type_id"`
	//阅读计数   高频的查询需要创建索引优化速度
	ReadCnt int64
	//点赞数
	LikeCnt int64
	//收藏数
	CollectCnt int64
	Utime      int64
	Ctime      int64
}

var ErrRecordNotFound = gorm.ErrRecordNotFound

type InteractiveDAO interface {
	IncrReadCnt(ctx context.Context, biz string, bizId int64) error
	BatchIncrReadCnt(ctx context.Context, bizs []string, bizIds []int64) error
	InsertLikeInfo(ctx context.Context, biz string, bizId int64, uid int64) error
	DeleteLikeInfo(ctx context.Context, biz string, bizId int64, uid int64) error
	InsertCollectionBiz(ctx context.Context, cb UserCollectionBiz) error
	GetLikeInfo(ctx context.Context,
		biz string, id int64, uid int64) (UserLikeBiz, error)
	GetCollectInfo(ctx context.Context,
		biz string, id int64, uid int64) (UserCollectionBiz, error)
	Get(ctx context.Context, biz string, id int64) (Interactive, error)
}

type GORMInteractiveDAO struct {
	db *gorm.DB
}

// 阅读次数的增加
func (dao *GORMInteractiveDAO) IncrReadCnt(ctx context.Context, biz string, bizId int64) error {
	//时，你想要更新该文章的阅读次数。如果你先查询 read_cnt 再执行更新操作，
	//在高并发情况下会导致多个进程读取到同一个值，从而造成更新时的计数冲突
	now := time.Now().UnixMilli()
	//要执行upset语句，你不可以先执行查询后在更新数值,这样会有并发问题不安全
	//但是你在查询的时候有考虑一个问题数据库有没数据，此时应该要有upsert语义
	return dao.db.WithContext(ctx).Clauses(clause.OnConflict{
		//尝试插入一条记录的时候主键或者唯一键进行冲突---primarykey和unique两者冲突
		DoUpdates: clause.Assignments(map[string]any{
			//遇到冲突的时候更新下面的字段，自增1
			"read_cnt": gorm.Expr("read_cnt + 1"),
			"utime":    time.Now().UnixMilli(),
		}),
	}).Create(&Interactive{
		//如果没有冲突直接插入一条新的记录，常识插入你这条数据，biz和bizId均存在的情况进行upsert
		Biz:     biz,
		BizId:   bizId,
		ReadCnt: 1,
		Ctime:   now,
		Utime:   now,
	}).Error
}

func (dao *GORMInteractiveDAO) BatchIncrReadCnt(ctx context.Context, bizs []string, bizIds []int64) error {
	//TODO implement me
	panic("implement me")
}

// 记录点赞的行为并更新点赞计数
func (dao *GORMInteractiveDAO) InsertLikeInfo(ctx context.Context, biz string, bizId int64, uid int64) error {
	//同时记录点赞数和更新点赞计数
	//需要一张点赞表记录谁给资源点赞
	now := time.Now().UnixMilli()
	return dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Clauses(clause.OnConflict{
			DoUpdates: clause.Assignments(map[string]any{
				//如果存在biz bizid uid就会进行update更新
				"utime":  now,
				"status": 1,
			}),
		}).Create(&UserLikeBiz{
			//如果用户没有对资源点赞过则插入一条新纪录，
			Biz:    biz,
			BizId:  bizId,
			Uid:    uid,
			Status: 1,
			Ctime:  now,
			Utime:  now,
		}).Error
		if err != nil {
			return err
		}
		//这个是总的表包含点赞收藏阅读的表
		return tx.Clauses(clause.OnConflict{
			DoUpdates: clause.Assignments(map[string]any{
				"like_cnt": gorm.Expr("like_cnt+1"),
				"utime":    time.Now().UnixMilli(),
			}),
		}).Create(&Interactive{
			Biz:     biz,
			BizId:   bizId,
			LikeCnt: 1,
			Ctime:   now,
			Utime:   now,
		}).Error

	})
}

// 数据库层面删除数据更新两张表，进行软删除减少点赞数量
func (dao *GORMInteractiveDAO) DeleteLikeInfo(ctx context.Context, biz string, bizId int64, uid int64) error {
	now := time.Now().UnixMilli()
	return dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Model(&UserLikeBiz{}).Where("biz=? AND biz_id=? AND uid=?",
			biz, bizId, uid).Updates(map[string]any{
			"utime":  now,
			"status": 0,
		}).Error
		if err != nil {
			return err
		}
		//取消点赞数据库中一定是会有数据的，所以你没必要upset直接update
		return tx.Model(&Interactive{}).Where("biz=? AND biz_id=?",
			biz, bizId).Updates(map[string]any{
			"utime":    now,
			"like_cnt": gorm.Expr("like_cnt-1"),
		}).Error

	})
}

// 插入收藏记录并且更新计数
func (dao *GORMInteractiveDAO) InsertCollectionBiz(ctx context.Context, cb UserCollectionBiz) error {
	now := time.Now().UnixMilli()

	return dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Create(&cb).Error
		if err != nil {
			return err
		}

		//upset语义
		return tx.WithContext(ctx).Clauses(clause.OnConflict{
			DoUpdates: clause.Assignments(map[string]interface{}{
				"collect_cnt": gorm.Expr("`collect_cnt` + 1"),
				"utime":       now,
			}),
		}).Create(&Interactive{
			Biz:        cb.Biz,
			BizId:      cb.BizId,
			CollectCnt: 1,
			Ctime:      now,
			Utime:      now,
		}).Error
	})
}

// 得到数据库中存储点赞表的数据库字段的第一条记录
func (dao *GORMInteractiveDAO) GetLikeInfo(ctx context.Context, biz string, id int64, uid int64) (UserLikeBiz, error) {
	var res UserLikeBiz
	err := dao.db.WithContext(ctx).
		Where("biz=? AND biz_id=? AND uid=? AND status=? ",
			biz, id, uid, 1).First(&res).Error
	return res, err
}

// 得到数据库中存储收藏表的数据库字段的第一条记录
func (dao *GORMInteractiveDAO) GetCollectInfo(ctx context.Context, biz string, id int64, uid int64) (UserCollectionBiz, error) {
	var res UserCollectionBiz
	err := dao.db.WithContext(ctx).
		Where("biz=? AND biz_id=? AND uid=?", biz, id, uid).First(&res).Error
	return res, err
}

// 缓存失败直接      从数据库中查询出来交互表中对应文章的点赞收藏评论数量
func (dao *GORMInteractiveDAO) Get(ctx context.Context, biz string, id int64) (Interactive, error) {
	var res Interactive
	err := dao.db.WithContext(ctx).
		Where("biz=? AND biz_id=?", biz, id).First(&res).Error
	return res, err
}

func NewGORMInteractiveDAO(db *gorm.DB) InteractiveDAO {
	return &GORMInteractiveDAO{db: db}
}

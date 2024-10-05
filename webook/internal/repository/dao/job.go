package dao

import (
	"context"
	"gorm.io/gorm"
	"time"
)

type JobDAO interface {
	Preempt(ctx context.Context) (Job, error)
	Release(ctx context.Context, id int64) error
	UpdateUtime(ctx context.Context, id int64) error
	UpdateNextTime(ctx context.Context, id int64, next time.Time) error
	Stop(ctx context.Context, id int64) error
}

type GORMJobDAO struct {
	db *gorm.DB
}

func (g *GORMJobDAO) Stop(ctx context.Context, id int64) error {
	return g.db.WithContext(ctx).Where("id=?", id).Updates(map[string]any{
		"status": jobStatusWaiting,
		"utime":  time.Now().UnixMilli(),
	}).Error
}

func (g *GORMJobDAO) UpdateNextTime(ctx context.Context, id int64, next time.Time) error {
	return g.db.WithContext(ctx).Model(&Job{}).Where("id=?", id).Updates(map[string]any{
		"next_time": next.UnixMilli(),
	}).Error
}
func (g *GORMJobDAO) UpdateUtime(ctx context.Context, id int64) error {
	return g.db.WithContext(ctx).Model(&Job{}).Where("id=?", id).
		Updates(map[string]any{
			"utime": time.Now().UnixMilli(),
		}).Error
}
func (g *GORMJobDAO) Release(ctx context.Context, id int64) error {
	return g.db.WithContext(ctx).Model(&Job{}).Where("id=?", id).Updates(map[string]any{
		"status": jobStatusWaiting,
		"utime":  time.Now().UnixMilli(),
	}).Error

}

func (g *GORMJobDAO) Preempt(ctx context.Context) (Job, error) {
	db := g.db.WithContext(ctx)
	//高并发的情况下大部分都是陪太子读书
	for {
		//第一个问题哪些任务可以抢，哪些任务呗站着哪些任务永远不会被运行
		//所有的goroutine执行的次数加载一起是1+2+3+4+。。。。99
		now := time.Now()
		var j Job
		//假设有多个进程或线程在并发执行，当它们同时进入第一个查询语句，
		//查找到相同的任务后，会尝试同时更新这个任务的状态

		//分布式任务调度系统
		//1.一次啦一批，我一次性的取出100条，然后我随机的从某一条开始向后进行抢占
		//2.我搞个随机偏移量，0-1000生成一个随机偏移量2，第一轮没查到回到0
		//3.搞一个id取余分配，
		err := db.Where("status=? AND next_time<=?", jobStatusWaiting, now).
			First(&j).Error
		//你找到了可以被抢占的之后你就要抢占
		if err != nil {
			//没有任务要从这里返回
			return Job{}, err
		}
		//多个进程拿到了相同的任务之后,都会去尝试更新任务同时更改字段会让数据库的一致性进行混乱
		//两个进程ab都认为自己抢占任务成功，并且同时开始执行
		//乐观锁，CAS操作，compare and swap 使用乐观锁取代       for update
		res := db.Where("id=? AND version=?", j.Id, j.Version).Model(&Job{}).Updates(map[string]any{
			"status":  jobStatusRunning,
			"utime":   now,
			"version": j.Version + 1,
		})
		if res.Error != nil {
			return Job{}, err
		}
		if res.RowsAffected == 0 {
			//抢占失败我只能继续下一轮
			continue
			//return Job{}, errors.New("没抢到")
		}
		return j, nil

	}

}

type Job struct {
	Id       int64 `gorm:"primaryKey,autoIncrement"`
	Cfg      string
	Executer string
	Status   int
	Name     string `gorm:"unique"`
	//定时任务我怎么知道已经到时间了
	//NextTime 下一次调度的时间
	//next<=nowandstatus=0
	NextTime int64 `gorm:"index"`
	Version  int
	Cron     string
	Utime    int64
	Ctime    int64
}

const (
	jobStatusWaiting = iota
	jobStatusRunning
)

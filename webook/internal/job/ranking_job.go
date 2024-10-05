package job

import (
	"awesomeProject/webook/internal/service"
	"awesomeProject/webook/pkg/logger"
	"context"
	rlock "github.com/gotomicro/redis-lock"
	"sync"
	"time"
)

type RankingJob struct {
	svc     service.RankingService
	timeout time.Duration
	//分布式锁
	client    *rlock.Client
	key       string
	l         logger.LoggerV1
	lock      *rlock.Lock
	localLock *sync.Mutex
}

func NewRankingJob(svc service.RankingService, client *rlock.Client,
	l logger.LoggerV1, timeout time.Duration) *RankingJob {
	return &RankingJob{
		svc:       svc,
		timeout:   timeout,
		client:    client,
		key:       "rlock:cron_job:ranking",
		l:         l,
		localLock: &sync.Mutex{},
	}
}

func (r *RankingJob) Name() string {
	return "ranking"
}

// 按照时间调度的三分钟一次
func (r *RankingJob) Run() error {
	r.localLock.Lock()
	if r.lock == nil {
		defer r.localLock.Unlock()
		//说明你没用拿到锁试着拿

		//这里运行的时候就需要加锁了
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		lock, err := r.client.Lock(ctx, r.key, r.timeout, &rlock.FixIntervalRetry{
			Interval: time.Millisecond * 100,
			Max:      0,
		}, time.Second)
		if err != nil {
			//这边没有拿到锁极大概率就是别人持有了锁
			return nil
		}
		r.lock = lock
		//如何保证这里一直拿着锁
		go func() {
			r.localLock.Lock()
			defer r.localLock.Unlock()
			//自动续约机制
			err1 := lock.AutoRefresh(r.timeout/2, time.Second)
			//续约失败了怎么办
			if err1 != nil {
				r.l.Error("续约失败", logger.Error(err))
			}
			r.lock = nil
		}()

	}
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()
	return r.svc.TopN(ctx)
}

func (r *RankingJob) Close() error {
	r.localLock.Lock()
	lock := r.lock
	r.lock = nil
	r.localLock.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return lock.Unlock(ctx)
}

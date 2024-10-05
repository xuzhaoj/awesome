package service

import (
	"awesomeProject/webook/internal/domain"
	"awesomeProject/webook/internal/repository"
	"awesomeProject/webook/pkg/logger"
	"context"
	"time"
)

type JobService interface {
	//Preempt抢占
	Preempt(ctx context.Context) (domain.Job, error)
	ResetNextTime(ctx context.Context, j domain.Job) error
}

type cronJobService struct {
	repo            repository.JobRepository
	refreshInterval time.Duration
	l               logger.LoggerV1
}

func (p *cronJobService) ResetNextTime(ctx context.Context, j domain.Job) error {
	next := j.NextTime()
	if next.IsZero() {
		return p.repo.Stop(ctx, j.Id)

	}
	return p.repo.UpdateNextTime(ctx, j.Id, next)
}
func (p *cronJobService) Preempt(ctx context.Context) (domain.Job, error) {
	j, err := p.repo.Preempt(ctx)
	//你的续约呢
	ticker := time.NewTicker(p.refreshInterval)
	go func() {
		for range ticker.C {
			p.refresh(j.Id)
		}
	}()
	j.CancelFunc = func() error {
		//自己在这里是反掉
		ticker.Stop()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		return p.repo.Release(ctx, j.Id)
	}
	return j, err

}
func (p *cronJobService) refresh(id int64) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	err := p.repo.UpdateUtime(ctx, id)
	if err != nil {
		p.l.Error("续约失败", logger.Error(err),
			logger.Int64("jid", id))

	}
}

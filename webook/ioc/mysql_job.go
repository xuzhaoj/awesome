package ioc

import (
	"awesomeProject/webook/internal/domain"
	"awesomeProject/webook/internal/job"
	"awesomeProject/webook/internal/service"
	"awesomeProject/webook/pkg/logger"
	"context"
	"time"
)

func InitScheduler(l logger.LoggerV1,
	local *job.LocalFuncExecutor,
	svc service.JobService) *job.Scheduler {
	res := job.NewScheduler(svc, l)
	res.RegisterExecutor(local)
	return res

}
func InitLocalFuncExecutor(svc service.RankingService) *job.LocalFuncExecutor {
	res := job.NewLocalFuncExecutor()
	//要在数据库中插入一条记录
	res.RegisterFuncs("ranking", func(ctx context.Context, j domain.Job) error {
		ctx, cancel := context.WithTimeout(ctx, time.Second*30)
		defer cancel()
		return svc.TopN(ctx)
	})
	return res
}

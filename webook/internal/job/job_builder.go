package job

import (
	"awesomeProject/webook/pkg/logger"
	"github.com/robfig/cron/v3"
	"time"
)

type CronJobBuilder struct {
	l logger.LoggerV1
}

func NewCronJobBuilder(l logger.LoggerV1) *CronJobBuilder {

	return &CronJobBuilder{
		l: l,
	}
}
func (b *CronJobBuilder) Build(job Job) cron.Job {
	name := job.Name()
	return cronJobFuncAdapter(func() error {
		start := time.Now()
		b.l.Info("任务开始",
			logger.String("job", name))

		defer func() {
			duration := time.Since(start).Milliseconds()
			b.l.Info("任务结束",
				logger.String("job", name),
				logger.Int64("duration_ms", duration))
		}()

		// 执行任务
		err := job.Run()
		if err != nil {
			b.l.Error("运行任务失败", logger.Error(err),
				logger.String("job", name))
		}

		return nil
	})
}

type cronJobFuncAdapter func() error

func (c cronJobFuncAdapter) Run() {
	_ = c()
}

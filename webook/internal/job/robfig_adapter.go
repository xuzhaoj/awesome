package job

import (
	"awesomeProject/webook/pkg/logger"
	"log"
	"time"
)

type RankingJobAdapter struct {
	j Job
	l logger.LoggerV1
}

func NewRankingJobAdapter(j Job, l logger.LoggerV1) *RankingJobAdapter {
	return &RankingJobAdapter{j: j, l: l}
}
func (r *RankingJobAdapter) Run() {
	start := time.Now()
	defer func() {
		duration := time.Since(start).Milliseconds()
		log.Println("execution time", duration)
	}()
	err := r.j.Run()
	if err != nil {
		r.l.Error("运行任务失败", logger.Error(err),
			logger.String("job", r.j.Name()))
	}
}

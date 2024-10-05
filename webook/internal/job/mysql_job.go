package job

import (
	"awesomeProject/webook/internal/domain"
	"awesomeProject/webook/internal/service"
	"awesomeProject/webook/pkg/logger"
	"context"
	"fmt"
	"golang.org/x/sync/semaphore"
	"time"
)

type Executer interface {
	Name() string
	Exec(ctx context.Context, j domain.Job) error
}

type LocalFuncExecutor struct {
	funcs map[string]func(ctx context.Context, j domain.Job) error
}

func NewLocalFuncExecutor() *LocalFuncExecutor {
	return &LocalFuncExecutor{funcs: make(map[string]func(ctx context.Context, j domain.Job) error)}
}

func (l *LocalFuncExecutor) RegisterFuncs(name string, fn func(ctx context.Context, j domain.Job) error) {
	l.funcs[name] = fn
}
func (l *LocalFuncExecutor) Name() string {
	return "local"
}
func (l *LocalFuncExecutor) Exec(ctx context.Context, j domain.Job) error {
	fn, ok := l.funcs[j.Name]
	if !ok {
		return fmt.Errorf("未知任务，你是否已经祖册好了？%s", j.Name)
	}
	return fn(ctx, j)
}

// Scheduler调度器
type Scheduler struct {
	execs   map[string]Executer
	svc     service.JobService
	l       logger.LoggerV1
	limiter *semaphore.Weighted
}

func NewScheduler(svc service.JobService, l logger.LoggerV1) *Scheduler {
	return &Scheduler{svc: svc, l: l, limiter: semaphore.NewWeighted(200), execs: make(map[string]Executer)}
}

func (s *Scheduler) RegisterExecutor(exec Executer) {

}

func (s *Scheduler) Schedule(ctx context.Context) error {

	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		err := s.limiter.Acquire(ctx, 1)
		if err != nil {
			return err
		}
		dbCtx, cancel := context.WithTimeout(ctx, time.Second)
		j, err := s.svc.Preempt(dbCtx)
		cancel()
		if err != nil {
			//不能return要继续进行
			s.l.Error("抢占任务失败", logger.Error(err))
		}

		exec, ok := s.execs[j.Executor]
		if !ok {
			s.l.Error("未找到对应的执行器",
				logger.String("executor", j.Executor))
			continue
		}

		//抢完开始执行
		go func() {
			defer func() {
				s.limiter.Release(1)
				err1 := j.CancelFunc()
				if err1 != nil {
					s.l.Error("释放任务失败",
						logger.Error(err1),
						logger.Int64("jid", j.Id))

				}

			}()
			err1 := exec.Exec(ctx, j)
			if err1 != nil {
				s.l.Error("任务执行失败", logger.Error(err1))

			}
			//你要不要考虑下此次调度
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			err1 = s.svc.ResetNextTime(ctx, j)
			if err1 != nil {
				s.l.Error("设置下一次执行时间失败", logger.Error(err1))
			}

		}()

	}
}

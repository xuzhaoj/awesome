package cronjob

import (
	"github.com/robfig/cron/v3"
	"log"
	"testing"
	"time"
)

func TestCronExpression(t *testing.T) {
	//创建一个新的 cron 调度器实例。调度器负责管理所有定时任务以秒为一个级别
	expr := cron.New(cron.WithSeconds())
	//添加一个新的定时任务
	//expr.AddJob("@every 1s", myJob{})
	//每隔 3 秒，cron 会启动一个新的 Goroutine 来执行这个任务。
	expr.AddFunc("@every 3s", func() {
		t.Log("开始长任务")
		time.Sleep(time.Second * 12)
		t.Log("结束长任务")
	})
	//启动调度器按照前面定义的规则进行执行
	expr.Start()
	//主线程休眠10秒
	time.Sleep(time.Second * 10)
	stop := expr.Stop()
	t.Log("已经发出停止信号")
	//这一句会阻塞等到所有已经调度结束才会返回
	<-stop.Done()
	t.Log("彻底结束")
}

type myJob struct{}

func (job myJob) Run() {
	log.Println("运行了！")

}

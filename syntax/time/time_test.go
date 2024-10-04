package time

import (
	"context"
	"testing"
	"time"
)

// 每秒打印一次
func TestTicker(t *testing.T) {
	//用ticker   固定时间间隔触发
	tm := time.NewTicker(time.Second)

	//保证当 TestTicker 函数返回时会停止 Ticker，释放资源
	defer tm.Stop()
	//在 10 秒后会自动超时
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	//无限循环
	for {
		select {
		//超过十秒的取消
		case <-ctx.Done():
			t.Log("超时或者被取消了")
			return
			//接受时间搓
		case now := <-tm.C:
			t.Log(now.Unix())
		}

	}

}

// 设置一个定时器，并通过 Goroutine 处理定时器触发的事件
func TestTimer(t *testing.T) {
	tm := time.NewTimer(time.Second)
	defer tm.Stop()

	go func() {
		for now := range tm.C {
			t.Log(now.Unix())
		}
	}()

	time.Sleep(time.Second * 10)

}

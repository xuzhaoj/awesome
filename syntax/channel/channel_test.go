package channel

import (
	"sync"
	"testing"
	"time"
)

func TestChannel(t *testing.T) {
	//创建容量为2的channel
	ch2 := make(chan int, 2)
	//往channel中写数据
	ch2 <- 123
	//从channel中读数据
	close(ch2)
	val, ok := <-ch2
	if !ok {

	}
	println(val)
}
func TestCloseChannel(t *testing.T) {
	ch := make(chan int, 2)
	ch <- 123
	ch <- 234
	val, ok := <-ch
	t.Log(val, ok)
	close(ch)
	//关掉之后就不能往里面去写
	//ch <- 124
	//特性可以把已经放入之后的数据给你读完即使你关掉了
	val1, ok1 := <-ch
	t.Log(val1, ok1)

	val2, ok2 := <-ch
	t.Log(val2, ok2)

}

// 谁创建的实力谁来关，最好就是定义在结构体上不要暴露
type MyStruct struct {
	ch        chan struct{}
	closeOnce sync.Once
}

//用户会多次调同，或者多个goroutine调同

func (m *MyStruct) Close() error {
	m.closeOnce.Do(func() {
		close(m.ch)
	})
	return nil
}

func TestLoopChannel(t *testing.T) {
	//无缓冲的channel
	ch := make(chan int)
	go func() {
		//生产者把数据都写道channel中      如果没有接收方准备好接收数据，发送方会被阻塞
		for i := 0; i < 10; i++ {
			ch <- i
			//发送完数据后线程会休眠100毫秒
			time.Sleep(time.Millisecond * 100)
		}
		close(ch)

	}()
	//这个是主goroutine，只有当channel被关闭后才会退出循环否则就是阻塞等待别的数据往channel中写数据并读出来
	for val := range ch {
		t.Log(val)
	}
	t.Log("channel被我关掉了")
}

//使用channel的时候接受者读不到消息就会阻塞1.没有缓存或者有缓存但是缓存中没有数据对面没有发送者
//如果发送者写不了数据就会阻塞，，，没有缓存或者缓存的空间已经满了。   对边没有接受者在等待接收数据

func TestSelect(t *testing.T) {
	ch1 := make(chan int, 1)
	ch2 := make(chan int, 1)
	go func() {
		time.Sleep(time.Millisecond * 100)
		ch1 <- 123
	}()
	go func() {
		time.Sleep(time.Millisecond * 100)
		ch2 <- 234
	}()

	//不阻塞的先执行,,,,,同时不阻塞的情况只会随机执行其中的一个，没有default的时候select就会全部阻塞
	select {
	case val := <-ch1:
		t.Log("ch1", val)
	case val := <-ch2:
		t.Log("ch2", val)
	}
}

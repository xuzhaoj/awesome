package sarama

import (
	"context"
	"github.com/IBM/sarama"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
	"log"
	"testing"
	"time"
)

func TestConsumer(t *testing.T) {
	cfg := sarama.NewConfig()
	//一个消费者都是归属于一个消费者组的也就是业务
	consumer, err := sarama.NewConsumerGroup(addrs, "test_group", cfg)
	require.NoError(t, err)
	//超时的context
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	//测试超时的context
	err = consumer.Consume(ctx, []string{"test_topic"}, testConsumerGroupHandler{})

	//err = consumer.Consume(context.Background(),
	//	[]string{"test_topic"}, testConsumerGroupHandler{})
	//你消费结束就会到这里
	t.Log(err, time.Since(start).String())

}

type testConsumerGroupHandler struct {
}

func (t testConsumerGroupHandler) Setup(session sarama.ConsumerGroupSession) error {
	//topic代表便宜量
	partitions := session.Claims()["test_topic"]
	for _, part := range partitions {
		session.ResetOffset("test_topic", part,
			sarama.OffsetOldest, "")
	}
	return nil
}

func (t testConsumerGroupHandler) Cleanup(session sarama.ConsumerGroupSession) error {
	log.Println("Cleanup")
	return nil
}
func (t testConsumerGroupHandler) ConsumeClaim(
	session sarama.ConsumerGroupSession,
	claim sarama.ConsumerGroupClaim) error {
	msgs := claim.Messages()
	for msg := range msgs {
		////msg.Value是字节形式需要序列化和反序列化
		//var bizMsg MyBizMsg
		//
		////基于json中的字段进行结构体反序列化
		//err := json.Unmarshal(msg.Value, &bizMsg)
		//if err != nil {
		//	//消费消息出错
		//	//大多数都是重试
		//	continue
		//}
		log.Println(string(msg.Value))
		//标记消费成功
		session.MarkMessage(msg, "")
	}
	//msg被关了要退出消费逻辑的时候
	return nil
}

// 异步消费，批量提交
func (t testConsumerGroupHandler) ConsumeClaimV1(
	session sarama.ConsumerGroupSession,
	claim sarama.ConsumerGroupClaim) error {
	msgs := claim.Messages()
	//msg 在每次迭代时会被复用，而不是为每次循环创建新的变量
	//for msg := range msgs {
	//	//需要重新的赋值不然在开启线程后会导致for循环先更新完最后拿到同一个msg
	//	m1 := msg
	//	go func() {
	//		//消费msg
	//		log.Println(string(m1.Value))
	//		session.MarkMessage(m1, "")
	//	}()
	//}
	//msg被关了要退出消费逻辑的时候

	const batchSize = 10
	for {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		var eg errgroup.Group
		var last *sarama.ConsumerMessage
		for i := 0; i < batchSize; i++ {
			done := false
			select {
			case <-ctx.Done():
				done = true
			case msg, ok := <-msgs:
				if !ok {
					cancel()
					//消费者被关闭了
					return nil
				}
				last = msg
				eg.Go(func() error {
					time.Sleep(time.Second)
					log.Println(string(msg.Value))
					return nil
				})

			}
			if done {
				break
			}

		}
		cancel()
		err := eg.Wait()
		if err != nil {
			continue
		}

		if last != nil {
			session.MarkMessage(last, "")
		}

	}
}

type MyBizMsg struct {
	Name string
}

// 返回只读的chan
func ChannelV1() <-chan struct{} {
	panic("sdfsdf")
}

// 返回可读可写
func ChannelV3() chan struct{} {
	panic("implement")
}

// 返回只写
func ChannelV2() chan<- struct{} {
	panic("implement")
}

package saramax

import (
	"awesomeProject/webook/pkg/logger"
	"context"
	"encoding/json"
	"github.com/IBM/sarama"
	"time"
)

type BatchHandler[T any] struct {
	//业务逻辑方法函数
	fn func(msgs []*sarama.ConsumerMessage, ts []T) error
	l  logger.LoggerV1
	//用option模式来设置
	batchSize     int
	batchDuration time.Duration
}

func NewBatchHandler[T any](l logger.LoggerV1, fn func(msgs []*sarama.ConsumerMessage, ts []T) error) *BatchHandler[T] {
	return &BatchHandler[T]{
		fn:            fn,
		l:             l,
		batchDuration: time.Second,
		batchSize:     10,
	}
}

// 正常这两个接口你什么都不需要干
func (b *BatchHandler[T]) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (b *BatchHandler[T]) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

// 从kafka中获取消息 进行批量处理
func (b *BatchHandler[T]) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	msgsCh := claim.Messages()
	batchSize := b.batchSize
	for {
		ctx, cancel := context.WithTimeout(context.Background(), b.batchDuration)
		done := false
		msgs := make([]*sarama.ConsumerMessage, 0, batchSize)
		ts := make([]T, 0, batchSize)
		for i := 0; i < batchSize && !done; i++ {
			select {
			case <-ctx.Done():
				done = true
			case msg, ok := <-msgsCh:
				// 再按照 key 或者业务 ID 转发到不同的 channel
				if !ok {
					cancel()
					// 代表消费者被关闭了
					return nil
				}
				var t T
				err := json.Unmarshal(msg.Value, &t)
				if err != nil {
					b.l.Error("反序列化失败",
						logger.Error(err),
						logger.String("topic", msg.Topic),
						logger.Int64("partition", int64(msg.Partition)),
						logger.Int64("offset", msg.Offset))
					continue
				}
				msgs = append(msgs, msg)
				ts = append(ts, t)
			}
		}
		cancel()
		if len(msgs) == 0 {
			continue
		}
		err := b.fn(msgs, ts)
		if err != nil {
			b.l.Error("调用业务批量接口失败",
				logger.Error(err))
			// 你这里整个批次都要记下来

			// 还要继续往前消费
		}
		for _, msg := range msgs {
			// 这样，万无一失
			session.MarkMessage(msg, "")
		}
	}
}

package saramax

import (
	"awesomeProject/webook/pkg/logger"
	"encoding/json"
	"github.com/IBM/sarama"
)

// 创建一个通用的api
type Handler[T any] struct {
	l logger.LoggerV1
	//这是一个业务处理函数，负责在 Kafka 消费到消息后，处理反序列化后的消息体 event。
	fn func(msg *sarama.ConsumerMessage, event T) error
}

func NewHandler[T any](l logger.LoggerV1, fn func(msg *sarama.ConsumerMessage, event T) error) *Handler[T] {
	return &Handler[T]{
		l:  l,
		fn: fn,
	}
}

func (h *Handler[T]) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (h *Handler[T]) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

// 代码的作用是从 Kafka 中读取消息，反序列化消息内容，
// 并调用业务处理逻辑处理每条消息。在处理过程中，它还包含了错误处理和重试机制
func (h *Handler[T]) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	//获取从 Kafka 分区中的消息，，返回的是一个 channel
	msgs := claim.Messages()
	//源源不断的从channel中获取消息
	for msg := range msgs {
		// 在这里调用业务处理逻辑这个泛型就是event结构体----Aid和Uid
		var t T
		err := json.Unmarshal(msg.Value, &t)

		if err != nil {
			h.l.Error("反序列化消息失败",
				logger.Error(err),
				logger.String("topic", msg.Topic),
				logger.Int64("partition", int64(msg.Partition)),
				logger.Int64("offset", msg.Offset))
			continue
		}
		// 在这里执行重试
		for i := 0; i < 3; i++ {
			err = h.fn(msg, t)
			//重试是去获取消息
			if err == nil {
				break
			}
			h.l.Error("处理消息失败",
				logger.Error(err),
				logger.String("topic", msg.Topic),
				logger.Int64("partition", int64(msg.Partition)),
				logger.Int64("offset", msg.Offset))
		}

		if err != nil {
			h.l.Error("处理消息失败-重试次数上限",
				logger.Error(err),
				logger.String("topic", msg.Topic),
				logger.Int32("partition", msg.Partition),
				logger.Int64("offset", msg.Offset))

		}
		session.MarkMessage(msg, "")
	}
	return nil
}

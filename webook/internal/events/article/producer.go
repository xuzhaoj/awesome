package article

import (
	"context"
	"encoding/json"
	"github.com/IBM/sarama"
)

type Producer interface {
	ProduceReadEvent(ctx context.Context, evt ReadEvent) error
}

// *****************************************领域事件只需要传递用户id和文章id就可以
type ReadEvent struct {
	Uid int64
	Aid int64
}

type KafkaProducer struct {
	//同步地将消息发送到kafka，同步的需要等待消息发送成功后才会放回结果
	producer sarama.SyncProducer
}

// 让外面初始化去传递不要自己初始化
func NewKafkaProducer(pc sarama.SyncProducer) Producer {
	return &KafkaProducer{
		producer: pc,
	}
}

// *************************************如果你的重试逻辑很简单你就放在这里
func (k *KafkaProducer) ProduceReadEvent(ctx context.Context, evt ReadEvent) error {
	//需要将这个对象序列化成字节json喜欢的格式转化成JSON格式的字节流
	data, err := json.Marshal(evt)
	if err != nil {
		return err
	}

	_, _, err = k.producer.SendMessage(&sarama.ProducerMessage{
		//代表了这条消息要发送到哪个kafka主题是个命名空间
		Topic: "read_article",
		//key就是在partition中分为多个不同的partition,Key相当于把他们写入不同的分区
		//没有指定key的话，topic会将他们随即分配到分区中
		//相同的key消息会被发送同一个分区，可以保持这些消息的顺序性

		//Value表示要传递的实际内容供给消费者进行消费，kafka的消息是字节流需要进行转化
		//必须使用这个接口设计就是这样
		Value: sarama.ByteEncoder(data),
	})
	return err
}

package sarama

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)
import "github.com/IBM/sarama"

// Kafka 代理的地址列表
var addrs = []string{"localhost:9094"}

// Kafka 同步生产者的消息发送能力

// 测试同步发消息
func TestSyncProducer(t *testing.T) {
	//新的 Sarama 配置对象，用于配置生产者和消费者的行为。
	cfg := sarama.NewConfig()
	//每次消息成功发送后，生产者会返回一个成功结果。
	cfg.Producer.Return.Successes = true
	cfg.Producer.Partitioner = sarama.NewHashPartitioner

	producer, err := sarama.NewSyncProducer(addrs, cfg)
	assert.NoError(t, err)

	//生产者生产一百条消息
	for i := 0; i < 100; i++ {
		//*******************************************************************一条消息，为了演示批量发送生产者的消息我要发送10条
		//这是核心部分，生产者向 Kafka 发送一条消息,SendMessage 函数会阻塞并等待消息被成功发送或出现错误。
		_, _, err = producer.SendMessage(&sarama.ProducerMessage{
			Topic: "read_article",
			// 消息数据体
			// 转 JSON 数据 或 protobuf    字符串消息编码为 Kafka 可以接受的字节流
			Value: sarama.StringEncoder(`{"aid": 1, "uid": 123}`),
			// 会在生产者和消费者之间传递
			//Headers: []sarama.RecordHeader{
			//	{
			//		//键和值
			//		Key:   []byte("trace_id"),
			//		Value: []byte("123456"),
			//	},
			//},

			// 作用于发送过程这个对象不会发送到 Kafka，而是可以用于发送成功后的处理逻辑
			//Metadata: "这是metadata",
		})
		assert.NoError(t, err)
	}

}

// 测试异步
func TestASyncProducer(t *testing.T) {
	cfg := sarama.NewConfig()
	cfg.Producer.Return.Errors = true
	cfg.Producer.Return.Successes = true
	producer, err := sarama.NewAsyncProducer(addrs, cfg)
	require.NoError(t, err)
	msgCh := producer.Input()
	go func() {
		for {
			msgCh <- &sarama.ProducerMessage{
				Topic: "test_topic",
				Key:   sarama.StringEncoder("oid-123"),
				// 消息数据体
				// 转 JSON 数据 或 protobuf    字符串消息编码为 Kafka 可以接受的字节流
				Value: sarama.StringEncoder("Hello, 这是一条消息A"),
				// 会在生产者和消费者之间传递
				Headers: []sarama.RecordHeader{
					{
						//键和值
						Key:   []byte("trace_id"),
						Value: []byte("123456"),
					},
				},
				Metadata: "这是metadata",
			}

		}
	}()

	errCh := producer.Errors()
	succCh := producer.Successes()

	for {
		//两个情况都没用发生就会阻塞
		select {
		case err := <-errCh:
			t.Log("发送出了问题", err.Err)
		case <-succCh:
			t.Log("发送成功")

		}
	}
}

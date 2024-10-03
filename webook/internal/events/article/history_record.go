package article

import (
	"awesomeProject/webook/internal/repository"
	"awesomeProject/webook/pkg/logger"
	"awesomeProject/webook/pkg/saramax"
	"context"
	"github.com/IBM/sarama"
	"time"
)

// 实现了对文章的 阅读计数
type HistoryReadEventConsumer struct {
	//客户端用于连接kafka集群
	client sarama.Client
	//业务代码
	repo repository.InteractiveRepository
	l    logger.LoggerV1
}

func NewHistoryReadEventConsumer(
	client sarama.Client,
	l logger.LoggerV1, repo repository.InteractiveRepository) *HistoryReadEventConsumer {
	return &HistoryReadEventConsumer{
		client: client,
		repo:   repo,
		l:      l,
	}
}

// 启动kafka消费者监听articleredad主题，开始消费信息
func (r *HistoryReadEventConsumer) Start() error {
	//创建消费者组，组名是interactive
	cg, err := sarama.NewConsumerGroupFromClient("interactive", r.client)
	if err != nil {
		return err
	}
	//启动一个kafka消费者消费循环
	go func() {
		//监听kafka主题article_read，
		er := cg.Consume(context.Background(),
			[]string{"read_article"},
			//自定义结构体的工厂模式，要求谁调用谁就传递相应的业务逻辑代码，交给你进行处理了
			//作为一个值去传递而不是立刻去调用
			saramax.NewHandler[ReadEvent](r.l, r.Consume))
		if er != nil {
			r.l.Error("退出消费循环异常", logger.Error(er))
		}
	}()
	return err
}

func (r *HistoryReadEventConsumer) Consume(msg *sarama.ConsumerMessage, t ReadEvent) error {
	//设置超时时间和关闭通道
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return r.repo.AddRecord(ctx, t.Aid, t.Uid)
}

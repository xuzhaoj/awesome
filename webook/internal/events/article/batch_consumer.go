package article

import (
	"awesomeProject/webook/internal/repository"
	"awesomeProject/webook/pkg/logger"
	"awesomeProject/webook/pkg/saramax"
	"context"
	"github.com/IBM/sarama"
	"time"
)

type InteractiveReadEventBatchConsumer struct {

	//客户端用于连接kafka集群
	client sarama.Client
	//业务代码
	repo repository.InteractiveRepository
	l    logger.LoggerV1
}

func NewInteractiveReadEventBatchConsumer(client sarama.Client, repo repository.InteractiveRepository, l logger.LoggerV1) *InteractiveReadEventBatchConsumer {
	return &InteractiveReadEventBatchConsumer{client: client, repo: repo, l: l}
}

// 启动kafka消费者监听articleredad主题，开始消费信息
func (r *InteractiveReadEventBatchConsumer) Start() error {
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
			saramax.NewBatchHandler[ReadEvent](r.l, r.Consume))
		if er != nil {
			r.l.Error("退出消费循环异常", logger.Error(er))
		}
	}()
	return err
}

func (r *InteractiveReadEventBatchConsumer) Consume(msg []*sarama.ConsumerMessage, ts []ReadEvent) error {

	ids := make([]int64, 0, len(ts))
	bizs := make([]string, 0, len(ts))

	for _, evt := range ts {
		ids = append(ids, evt.Aid)
		bizs = append(bizs, "article")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err := r.repo.BatchIncrReadCnt(ctx, bizs, ids)
	if err != nil {
		r.l.Error("批量增加阅读计数失败",
			logger.Field{Key: "ids", Val: ids},
			logger.Error(err))
	}
	return nil
}

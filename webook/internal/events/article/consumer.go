package article

import (
	"awesomeProject/webook/internal/repository"
	"awesomeProject/webook/pkg/logger"
	"awesomeProject/webook/pkg/saramax"
	"context"
	"github.com/IBM/sarama"
	"time"
)

type InteractiveReadEventConsumer struct {
	client sarama.Client
	repo   repository.InteractiveRepository
	l      logger.LoggerV1
}

func NewInteractiveReadEventConsumer(
	client sarama.Client,
	l logger.LoggerV1, repo repository.InteractiveRepository) *InteractiveReadEventConsumer {
	return &InteractiveReadEventConsumer{
		client: client,
		repo:   repo,
		l:      l,
	}
}
func (r *InteractiveReadEventConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("interactive", r.client)
	if err != nil {
		return err
	}
	go func() {
		er := cg.Consume(context.Background(),
			[]string{"article_read"},
			saramax.NewHandler[ReadEvent](r.l, r.Consume))
		if er != nil {
			r.l.Error("退出消费循环异常", logger.Error(er))
		}
	}()
	return err
}

func (r *InteractiveReadEventConsumer) Consume(msg *sarama.ConsumerMessage, t ReadEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return r.repo.IncrReadCnt(ctx, "article", t.Aid)
}

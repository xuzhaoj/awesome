package service

import (
	"awesomeProject/webook/internal/domain"
	"awesomeProject/webook/internal/repository"
	"awesomeProject/webook/pkg/logger"
	"context"
	"golang.org/x/sync/errgroup"
)

//go:generate mockgen -source=./interactive.go -package=svcmocks -destination=mocks/interactive.mock.go InteractiveService
type InteractiveService interface {
	//使用biz和bizId来标识唯一的某个特定的业务

	//实现阅读计数功能
	IncrReadCnt(ctx context.Context, biz string, bizId int64) error
	Like(ctx context.Context, biz string, bizId int64, uid int64) error
	CancelLike(ctx context.Context, biz string, bizId int64, uid int64) error
	Collect(ctx context.Context, biz string, bizId, cid, uid int64) error
	Get(ctx context.Context, biz string, bizId int64, uid int64) (domain.Interactive, error)

	//点赞热榜的实现
	GetByIds(ctx context.Context, biz string, bizIds []int64) (map[int64]domain.Interactive, error)
}
type interactiveService struct {
	repo repository.InteractiveRepository
	l    logger.LoggerV1
}

func newInteractiveService(repo repository.InteractiveRepository, l logger.LoggerV1) InteractiveService {
	return &interactiveService{repo: repo, l: l}

}

func (i *interactiveService) GetByIds(ctx context.Context, biz string, bizIds []int64) (map[int64]domain.Interactive, error) {
	//TODO implement me
	panic("implement me")
}

func NewInteractiveService(r repository.InteractiveRepository, l logger.LoggerV1) InteractiveService {
	return &interactiveService{
		repo: r,
		l:    l,
	}
}

// 实现阅读计数功能
func (i *interactiveService) IncrReadCnt(ctx context.Context, biz string, bizId int64) error {
	return i.repo.IncrReadCnt(ctx, biz, bizId)

}

// 点赞功能的实现
func (i *interactiveService) Like(ctx context.Context, biz string, bizId int64, uid int64) error {
	//"						Article,文章id,用户id"  存储的层面上就是增加点赞数取消点赞数
	return i.repo.IncrLike(ctx, biz, bizId, uid)
}

// 取消点赞的功能实现
func (i *interactiveService) CancelLike(ctx context.Context, biz string, bizId int64, uid int64) error {
	return i.repo.DecrLike(ctx, biz, bizId, uid)
}

// 收藏功能的实现   cid收藏夹的id
func (i *interactiveService) Collect(ctx context.Context,
	biz string, bizId, cid, uid int64) error {
	return i.repo.AddCollectionItem(ctx, biz, bizId, cid, uid)
}

// 读取文章详细信息的时候把文章的点赞收藏阅读记录数据一并返回过来
func (i *interactiveService) Get(ctx context.Context,
	biz string, bizId int64, uid int64) (domain.Interactive, error) {
	//repository层应该完全的把domain的结构体给完整的返回，然后handler中转化成VO返回给前端进行展示
	//通过并发执行完函数
	var (
		eg        errgroup.Group
		intr      domain.Interactive
		liked     bool
		collected bool
	)

	eg.Go(func() error {
		var err error
		intr, err = i.repo.Get(ctx, biz, bizId)
		return err
	})
	eg.Go(func() error {
		var err error
		liked, err = i.repo.Liked(ctx, biz, bizId, uid)
		return err
	})
	eg.Go(func() error {
		var err error
		collected, err = i.repo.Collected(ctx, biz, bizId, uid)
		return err
	})
	err := eg.Wait()
	if err != nil {
		return domain.Interactive{}, err

	}
	intr.Liked = liked
	intr.Collected = collected
	return intr, err

}

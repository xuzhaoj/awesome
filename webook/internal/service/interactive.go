package service

import (
	"awesomeProject/webook/internal/domain"
	"awesomeProject/webook/internal/repository"
	"awesomeProject/webook/pkg/logger"
	"context"
)

type InteractiveService interface {
	//使用biz和bizId来标识唯一的某个特定的业务

	//实现阅读计数功能
	IncrReadCnt(ctx context.Context, biz string, bizId int64) error
	Like(ctx context.Context, biz string, bizId int64, uid int64) error
	CancelLike(ctx context.Context, biz string, bizId int64, uid int64) error
	Collect(ctx context.Context, biz string, bizId, cid, uid int64) error
	Get(ctx context.Context, biz string, bizId int64, uid int64) (domain.Interactive, error)
}
type interactiveService struct {
	repo repository.InteractiveRepository
	l    logger.LoggerV1
}

func NewInteractiveService(r repository.InteractiveRepository, l logger.LoggerV1) InteractiveService {
	return &interactiveService{
		repo: r,
		l:    l,
	}
}

// 实现阅读计数功能
func (i interactiveService) IncrReadCnt(ctx context.Context, biz string, bizId int64) error {
	return i.repo.IncrReadCnt(ctx, biz, bizId)

}

// 点赞功能的实现
func (i interactiveService) Like(ctx context.Context, biz string, bizId int64, uid int64) error {
	//"						Article,文章id,用户id"  存储的层面上就是增加点赞数取消点赞数
	return i.repo.IncrLike(ctx, biz, bizId, uid)
}

// 取消点赞的功能实现
func (i interactiveService) CancelLike(ctx context.Context, biz string, bizId int64, uid int64) error {
	return i.repo.DecrLike(ctx, biz, bizId, uid)
}

func (i interactiveService) Collect(ctx context.Context, biz string, bizId, cid, uid int64) error {
	//TODO implement me
	panic("implement me")
}

func (i interactiveService) Get(ctx context.Context, biz string, bizId int64, uid int64) (domain.Interactive, error) {
	//TODO implement me
	panic("implement me")
}

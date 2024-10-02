package web

import (
	"awesomeProject/webook/internal/domain"
	"awesomeProject/webook/internal/service"
	"awesomeProject/webook/pkg/logger"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
	"net/http"
	"strconv"
	"time"
)

var _ Handler = (*ArticleHandler)(nil)

type ArticleHandler struct {
	svc service.ArticleService
	l   logger.LoggerV1
	//阅读数量、点赞、评论
	intrSvc service.InteractiveService
	biz     string
}

func NewArticleHandler(svc service.ArticleService, l logger.LoggerV1, intrSvc service.InteractiveService) *ArticleHandler {
	return &ArticleHandler{
		svc:     svc,
		l:       l,
		intrSvc: intrSvc,
		biz:     "article",
	}
}
func (h *ArticleHandler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/articles")
	g.POST("/edit", h.Edit)
	g.POST("/publish", h.Publish)
	g.POST("/withdraw", h.Withdraw)
	//创作者的查询接口
	g.POST("/list", h.List)

	//作者查看文章详情的接口
	g.GET("/detail/:id", h.Detail)

	//读者查看文章线上库--------可以同时具有阅读次数的显示
	pub := g.Group("/pub")
	pub.GET("/:id", h.PubDetail)
	pub.POST("/like", h.Like)
	pub.POST("/collect", h.Collect)
}

func (h *ArticleHandler) Detail(ctx *gin.Context) {
	idstr := ctx.Param("id")
	id, err := strconv.ParseInt(idstr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Msg:  "id 参数错误",
			Code: 4,
		})
		h.l.Warn("查询文章失败，id的格式不对",
			logger.String("id", idstr),
			logger.Error(err))
		return
	}
	art, err := h.svc.GetById(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Msg:  "系统错误",
			Code: 5,
		})
		h.l.Error("查询文章失败",
			logger.Int64("id", id),
			logger.Error(err))
		return
	}
	c := ctx.MustGet("claims")
	claims, ok := c.(*UserClaims)
	if !ok {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.l.Error("未发现用户的session")
		return
	}
	//有人在搞系统黑客攻击窃取数据
	if art.Author.Id != claims.Uid {
		ctx.JSON(http.StatusOK, Result{

			Code: 5,
			Msg:  "系统错误",
		})
		h.l.Error("非法查询文章", logger.Int64("id", id),
			logger.Int64("uid", claims.Uid))
		return
	}
	vo := ArticleVO{
		Id:    art.Id,
		Title: art.Title,
		//Abstract: art.Abstract(),

		Content: art.Content,
		//这个vo的author是int类型
		//Author: art.Author.Id,
		// 列表，你不需要
		Status: art.Status.ToUint8(),
		Ctime:  art.Ctime.Format(time.DateTime),
		Utime:  art.Utime.Format(time.DateTime),
	}
	ctx.JSON(http.StatusOK, Result{Data: vo})

}

//前端的抽象

// 接口查询数据
func (h *ArticleHandler) List(ctx *gin.Context) {
	type ListReq struct {
		Offset int `json:"offset"`
		Limit  int `json:"limit"`
	}
	var req ListReq
	//将前端传递过来的bind的结构体绑定到ListReq上
	if err := ctx.Bind(&req); err != nil {

		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	c := ctx.MustGet("claims")
	claims, ok := c.(*UserClaims)
	if !ok {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.l.Error("未发现用户的session")
		return
	}

	//调用接口查询数据，这两个需要从前端传递过来吗目前看来应该是不需要
	arts, err := h.svc.List(ctx, claims.Uid, req.Offset, req.Limit)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "查询数据失败",
		})
	}
	//在列表页,不显示全文，只显示一个摘要简单的几句话，强大的摘要是ai生成的
	ctx.JSON(http.StatusOK, Result{
		Data: slice.Map[domain.Article, ArticleVO](arts, func(idx int, src domain.Article) ArticleVO {
			return ArticleVO{
				Id:       src.Id,
				Title:    src.Title,
				Abstract: src.Abstract(),

				//Content:  src.Content,
				//Author: src.Author.Id,
				// 列表，你不需要
				Status: src.Status.ToUint8(),
				Ctime:  src.Ctime.Format(time.DateTime),
				Utime:  src.Utime.Format(time.DateTime),
			}
		}),
	})
}

//	func mapArticlesToVO(res []domain.Article) []ArticleVO {
//		result := make([]ArticleVO, len(res)) // 创建目标切片，长度与输入相同
//		for idx, src := range res {
//			result[idx] = ArticleVO{
//				Id:       src.Id,
//				Title:    src.Title,
//				Abstract: src.Abstract(), // 只返回摘要，不需要内容字段
//				// Content: src.Content,    // 不需要内容字段
//				// Author: src.Author,      // 自己查看自己的文章不需要这个字段
//				Status: src.Status.ToUint8(),
//				Ctime:  src.Ctime.Format(time.DateTime), // 正确的时间格式化
//				Utime:  src.Utime.Format(time.DateTime), // 正确的时间格式化
//			}
//		}
//		return result
//	}
func (h *ArticleHandler) Withdraw(ctx *gin.Context) {
	//不需要使用整个文章的结构体只需结构体里面有id即可
	type Req struct {
		//这个是对应帖子的id
		Id int64
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}
	c := ctx.MustGet("claims")
	claims, ok := c.(*UserClaims)
	if !ok {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.l.Error("未发现用户的session")
		return
	}
	err := h.svc.Withdraw(ctx, domain.Article{
		//帖子的id
		Id: req.Id,
		Author: domain.Author{
			Id: claims.Uid,
		},
	})
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.l.Error("发表帖子失败", logger.Error(err))
		return
	}
	//模拟的是这里的ok
	ctx.JSON(http.StatusOK, Result{
		Msg: "OK",
	})

}
func (h *ArticleHandler) Publish(ctx *gin.Context) {
	var req ArticleReq
	if err := ctx.Bind(&req); err != nil {
		return
	}
	//获取用户登录的id
	c := ctx.MustGet("claims")
	claims, ok := c.(*UserClaims)
	if !ok {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.l.Error("未发现用户的session")
		return
	}

	id, err := h.svc.Publish(ctx, req.toDomain(claims.Uid))
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.l.Error("发表帖子失败", logger.Error(err))
		return
	}
	//模拟的是这里的ok
	ctx.JSON(http.StatusOK, Result{
		Msg:  "OK",
		Data: id,
	})

}

//1.定义Handle结构体，注册路由，编写路由方法，考虑路由方法需要传递的参数构造响应结构体

func (h *ArticleHandler) Edit(ctx *gin.Context) {

	var req ArticleReq

	if err := ctx.Bind(&req); err != nil {
		return
	}

	//获取用户登录的id
	c := ctx.MustGet("claims")
	claims, ok := c.(*UserClaims)
	if !ok {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.l.Error("未发现用户的session")
		return
	}

	//检测输入
	//调用service代码，文章的id
	id, err := h.svc.Save(ctx, domain.Article{
		Id:      req.Id,
		Title:   req.Title,
		Content: req.Content,
		Author: domain.Author{
			Id: claims.Uid,
		},
	})
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.l.Error("保存帖子失败", logger.Error(err))
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Msg:  "OK",
		Data: id,
	})
}

// 获取单篇文章的详情信息，
func (h *ArticleHandler) PubDetail(ctx *gin.Context) {
	idstr := ctx.Param("id")
	//字符串转整数
	id, err := strconv.ParseInt(idstr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Msg:  "id 参数错误",
			Code: 4,
		})
		h.l.Warn("查询文章失败，id 格式不对",
			logger.String("id", idstr),
			logger.Error(err))
		return
	}

	var eg errgroup.Group
	var art domain.Article
	//获取用户登录的id
	c := ctx.MustGet("claims")
	claims, ok := c.(*UserClaims)
	if !ok {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.l.Error("未发现用户的session")
		return
	}
	//查询出来文章信息了.异步去执行
	eg.Go(func() error {
		art, err = h.svc.GetPublishedById(ctx, id, claims.Uid)
		return err
		//if err != nil {
		//	ctx.JSON(http.StatusOK, Result{
		//		Code: 5,
		//		Msg:  "系统错误",
		//	})
		//	h.l.Error("获得文章的信息失败", logger.Error(err))
		//	return
		//}
	})

	//***********88888888888888*****要在这里获取文章的阅读收藏和点赞******************************
	var intr domain.Interactive
	eg.Go(func() error {

		//查询所有的点赞收藏还有阅读
		intr, err = h.intrSvc.Get(ctx, h.biz, id, claims.Uid)
		//容忍错误的写法
		//if err!=nil{
		//	//记录日志
		//}
		//return nil
		return err

	})

	//要等前面两个结束后才可以开启下面的
	err = eg.Wait()
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}

	//***********************************************************反正你获取好文章后再开异步
	go func() {
		//导致数据库压力大
		//增加阅读计数,开一个goroutine异步去执行就行了
		er := h.intrSvc.IncrReadCnt(ctx, h.biz, art.Id)
		if er != nil {
			h.l.Error("增加阅读计数失败",
				logger.Int64("aid", art.Id),
				logger.Error(er))
		}
	}()

	ctx.JSON(http.StatusOK, Result{
		Data: ArticleVO{
			Id:      art.Id,
			Title:   art.Title,
			Content: art.Content,
			Status:  art.Status.ToUint8(),
			//Author:   art.Author.Id,
			//要把作者信息带上
			Author:     art.Author.Name,
			ReadCnt:    intr.ReadCnt,
			CollectCnt: intr.CollectCnt,
			LikeCnt:    intr.LikeCnt,
			Liked:      intr.Liked,
			Collected:  intr.Collected,
			Ctime:      art.Ctime.Format(time.DateTime),
			Utime:      art.Utime.Format(time.DateTime),
		},
	})
}

// 点赞,取消点赞同时复用接口的实现
func (h *ArticleHandler) Like(ctx *gin.Context) {
	type LikeReq struct {
		//代表文章的id
		Id   int64 `json:"id"`
		Like bool  `json:"like"`
	}
	var req LikeReq
	if err := ctx.Bind(&req); err != nil {
		return
	}
	//获取用户登录的id
	c := ctx.MustGet("claims")
	claims, ok := c.(*UserClaims)
	if !ok {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.l.Error("未发现用户的session")
		return
	}
	var err error
	if req.Like {
		//"Article,文章id,用户id"
		err = h.intrSvc.Like(ctx, h.biz, req.Id, claims.Uid)
	} else {
		err = h.intrSvc.CancelLike(ctx, h.biz, req.Id, claims.Uid)
	}
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
	}
	ctx.JSON(http.StatusOK, Result{
		Msg: "OK",
	})
}

// 收藏进收藏夹的功能
func (h *ArticleHandler) Collect(ctx *gin.Context) {
	type Req struct {
		Id int64 `json:"id"`
		//前端传递收藏夹的id比如说后端开发收藏夹的id为1,前端开发的收藏夹为2这样
		Cid int64 `json:"cid"`
	}
	var req Req
	//把前端的数据绑定到结构体上
	if err := ctx.Bind(&req); err != nil {
		return
	}
	//获取用户登录的id
	c := ctx.MustGet("claims")
	claims, ok := c.(*UserClaims)
	if !ok {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.l.Error("未发现用户的session")
		return
	}
	//进行收藏的功能，就不需要返回什么了把
	err := h.intrSvc.Collect(ctx, h.biz, req.Id, req.Cid, claims.Uid)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5, Msg: "系统错误",
		})
		h.l.Error("收藏失败",
			logger.Error(err),
			logger.Int64("uid", claims.Uid),
			logger.Int64("aid", req.Id))
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Msg: "OK",
	})

}

//func (h *ArticleHandler) Like(c *gin.Context) {
//	type Req struct {
//		Id int64 `json:"id"`
//		// true 是点赞，false 是不点赞
//		Like bool `json:"like"`
//	}
//	var req Req
//	if err := c.Bind(&req); err != nil {
//		return
//	}
//	uc := c.MustGet("user").(jwt.UserClaims)
//	var err error
//	if req.Like {
//		// 点赞
//		err = h.intrSvc.Like(c, h.biz, req.Id, uc.Uid)
//	} else {
//		// 取消点赞
//		err = h.intrSvc.CancelLike(c, h.biz, req.Id, uc.Uid)
//	}
//	if err != nil {
//		c.JSON(http.StatusOK, Result{
//			Code: 5, Msg: "系统错误",
//		})
//		h.l.Error("点赞/取消点赞失败",
//			logger.Error(err),
//			logger.Int64("uid", uc.Uid),
//			logger.Int64("aid", req.Id))
//		return
//	}
//	c.JSON(http.StatusOK, Result{
//		Msg: "OK",
//	})
//}

type ArticleReq struct {
	//修改可以获得文章的id
	Id      int64  `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

// 这就是一个前端传递数据的赋值
func (req ArticleReq) toDomain(uid int64) domain.Article {
	return domain.Article{
		Id:      req.Id,
		Title:   req.Title,
		Content: req.Content,
		Author: domain.Author{
			Id: uid,
		},
	}

}

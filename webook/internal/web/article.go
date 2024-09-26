package web

import (
	"awesomeProject/webook/internal/domain"
	"awesomeProject/webook/internal/service"
	"awesomeProject/webook/pkg/logger"
	"github.com/gin-gonic/gin"
	"net/http"
)

var _ Handler = (*ArticleHandler)(nil)

type ArticleHandler struct {
	svc service.ArticleService
	l   logger.LoggerV1
}

func NewArticleHandler(svc service.ArticleService, l logger.LoggerV1) *ArticleHandler {
	return &ArticleHandler{
		svc: svc,
		l:   l,
	}
}
func (h *ArticleHandler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/articles")
	g.POST("/edit", h.Edit)

}

//1.定义Handle结构体，注册路由，编写路由方法，考虑路由方法需要传递的参数构造响应结构体

func (h *ArticleHandler) Edit(ctx *gin.Context) {
	type Req struct {
		//修改可以获得文章的id
		Id      int64  `json:"id"`
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	var req Req
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

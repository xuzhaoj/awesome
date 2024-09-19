package middleware

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type LoginMiddlewareBuilder struct {
}

func NewLoginMiddlewareBuilder() *LoginMiddlewareBuilder {
	return &LoginMiddlewareBuilder{}
}

// 步骤2
func (l *LoginMiddlewareBuilder) Build() gin.HandlerFunc {
	return func(context *gin.Context) {
		//这些是不需要登录校验的，给我直接登录
		if context.Request.URL.Path == "/users/login" || context.Request.URL.Path == "/users/signup" {
			return

		}

		//获取会话，已经配置完成，存储用户的id
		sess := sessions.Default(context)
		//没有session，前面已经设置好了所以不可能没有new
		//if sess == nil {
		//	context.AbortWithStatus(http.StatusUnauthorized)
		//	return
		//}

		id := sess.Get("userId")
		//没有id
		if id == nil {
			context.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		updateTime := sess.Get("update_time")
		sess.Set("userId", id)
		sess.Options(sessions.Options{
			MaxAge: 60,
		})
		now := time.Now().UnixMilli()
		//说明没有刷新过，刚刚登录
		if updateTime == nil {

			sess.Set("update_time", now)

			sess.Save()
		}
		//以前登录过的，在间隔的时间内就要重新刷新,断言
		updateTimeVal, ok := updateTime.(int64)
		if !ok {
			context.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		//间隔时间
		if now-updateTimeVal > 10*1000 {
			//需要重新刷新
			sess.Set("update_time", now)
			sess.Save()
			return

		}

	}
}

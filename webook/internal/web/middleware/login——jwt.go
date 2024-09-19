package middleware

import (
	"awesomeProject/webook/internal/web"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"log"
	"net/http"
	"strings"
	"time"
)

type LoginJWTMiddlewareBuilder struct {
	paths []string
}

func NewLoginJWTMiddlewareBuilder() *LoginJWTMiddlewareBuilder {
	return &LoginJWTMiddlewareBuilder{}
}
func (l *LoginJWTMiddlewareBuilder) IgnorePaths(path string) *LoginJWTMiddlewareBuilder {
	l.paths = append(l.paths, path)
	return l

}

// CheckLogin 使用了JWT进行登录校验
func (l *LoginJWTMiddlewareBuilder) Build() gin.HandlerFunc {

	return func(ctx *gin.Context) {
		// 忽略掉 登录和注册接口
		for _, path := range l.paths {
			if ctx.Request.URL.Path == path {
				return
			}

		}
		tokenHeader := ctx.GetHeader("Authorization")
		if tokenHeader == "" {

			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		segs := strings.Split(tokenHeader, " ")
		if len(segs) != 2 || segs[0] != "Bearer" {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		tokenStr := segs[1]

		claims := &web.UserClaims{}

		//一定要传入指针不然只是复制一份
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte("iSQXg9EZhWMbSxYkExJaj4zbflnHCppl"), nil
		})
		if err != nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		//会自己校验会不会过期
		if token == nil || !token.Valid || claims.Uid == 0 {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		//******************************************************************匹配安全问题进行系统保护*************************************************************************
		//claims.UserAgent 是用户登录时保存的用户代理信息。
		//ctx.Request.UserAgent() 是当前请求的用户代理信息。
		if claims.UserAgent != ctx.Request.UserAgent() {
			//严重的安全问题
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		now := time.Now()
		//每十秒刷新一次
		if claims.ExpiresAt.Sub(now) < time.Second*50 {
			//没过期需要去刷新一下token
			claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(time.Minute))
			//生成token
			tokenStr, err = token.SignedString([]byte("iSQXg9EZhWMbSxYkExJaj4zbflnHCppl"))
			if err != nil {
				log.Println("jwt续约失败", err)
			}
			ctx.Header("x-jwt-token", tokenStr)

		}

		//fmt.Println(claims.Uid),放置在中间件中
		ctx.Set("claims", claims)

	}

}

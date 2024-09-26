package web

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"time"
)

type jwtHandler struct {
	atKey []byte
	rtKey []byte
}

func NewJwtHandler() jwtHandler {
	return jwtHandler{
		atKey: []byte("iSQXg9EZhWMbSxYkExJaj4zbflnHCppl"),
		rtKey: []byte("iSQXg9EZhWMbSxYkExJaj4zbflnHCppl"),
	}

}

func (h jwtHandler) setJWTToken(context *gin.Context, uid int64) error {
	//*******************************************************************登陆成功****************************************************************************
	//设置jwt登陆状态，生成jwttoken
	//设置登录态，生成token
	//带userID
	claims := UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 30)),
		},
		Uid: uid,
		//标识用户的软件和硬件信息
		UserAgent: context.Request.UserAgent(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	tokenStr, err := token.SignedString(h.atKey)
	if err != nil {
		context.String(http.StatusInternalServerError, "系统错误")
		return err
	}
	//在前响应头中塞进去
	context.Header("x-jwt-token", tokenStr)
	return nil
}

//func (h jwtHandler) setRefreshToken(context *gin.Context, uid int64) error {
//	//*******************************************************************登陆成功****************************************************************************
//	//设置jwt登陆状态，生成jwttoken
//	//设置登录态，生成token
//	//带userID
//	claims := RefreshClaims{
//		RegisteredClaims: jwt.RegisteredClaims{
//			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 30)),
//		},
//		Uid: uid,
//	}
//	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
//	tokenStr, err := token.SignedString(h.rtKey)
//	if err != nil {
//		context.String(http.StatusInternalServerError, "系统错误")
//		return err
//	}
//	//在前响应头中塞进去
//	context.Header("x-refresh-token", tokenStr)
//	return nil
//}
//
//type RefreshClaims struct {
//	jwt.RegisteredClaims
//	//声明要放进去token里面的数据
//	Uid int64
//}

type UserClaims struct {
	jwt.RegisteredClaims
	//声明要放进去token里面的数据
	Uid       int64
	UserAgent string
}

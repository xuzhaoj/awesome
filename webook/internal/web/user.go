package web

import (
	"awesomeProject/webook/internal/domain"
	"awesomeProject/webook/internal/service"
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"regexp"
	"time"
)

// 定义与用户有关的路由
const biz = "login"

type UserHandler struct {
	svc     *service.UserService
	codeSvc *service.CodeService
}

func NewUserHandler(svc *service.UserService, codeSvc *service.CodeService) *UserHandler {

	return &UserHandler{
		svc:     svc,
		codeSvc: codeSvc,
	}

}

func (u *UserHandler) RegisterRoutes(server *gin.Engine) {

	//对于前缀相同的我们可以引入分组路由
	ug := server.Group("/users")
	ug.POST("/signup", u.SingUp)
	ug.POST("/login", u.LoginJWT)
	ug.POST("/edit", u.Edit)
	//发送验证码
	ug.POST("/login_sms/code/send", u.SendLoginSMSCode)
	ug.POST("/login_sms", u.LoginSMS)
	ug.GET("/profile", u.Profile)

}

func (u *UserHandler) LoginSMS(ctx *gin.Context) {
	//前端传递过来的参数映射
	type Req struct {
		Phone string `json:"phone"`
		Code  string `json:"code"`
	}
	var req Req
	err := ctx.Bind(&req)
	if err != nil {
		return
	}
	ok, err := u.codeSvc.Verify(ctx, biz, req.Phone, req.Code)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{

			Code: 5,
			Msg:  "系统错误",
		})
	}
	if !ok {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "验证码错误请重新输入",
		})
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Code: 4,
		Msg:  "验证码校验通过",
	})
}

func (u *UserHandler) SendLoginSMSCode(ctx *gin.Context) {
	//从前端传递过来的值进行json映射
	type Req struct {
		Phone string `json:"phone"`
	}

	var req Req
	//检查前端的json是否映射正确
	if err := ctx.Bind(&req); err != nil {
		return
	}

	//是否是一个合法的手机号码
	if req.Phone == "" {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "输入的手机号有误请重新输入",
		})
		return

	}
	err := u.codeSvc.Send(ctx, biz, req.Phone)
	switch err {
	case nil:
		ctx.JSON(http.StatusOK, Result{
			Msg: "发送验证码成功",
		})
	case service.ErrCodeSendTooMany:
		ctx.JSON(http.StatusOK, Result{
			Msg: "发送频繁，请稍后再试",
		})
	default:
		//ctx.String(http.StatusOK, "系统异常")
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})

	}

}

func (u *UserHandler) SingUp(context *gin.Context) {
	type SignUpReq struct {
		Email           string `json:"email"`
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirmPassword"`
	}
	fmt.Println("可以到这里")
	//初始化结构体对象
	var req SignUpReq
	//Bind方法会根据COntent-type来解析到你的数据到req中,解析失败返回400
	if err := context.Bind(&req); err != nil {
		return
	}
	const (
		emailRegexPattern    = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
		passwordRegexPattern = `^[A-Za-z\d]{8,}$`
	)
	//ok是用来判断是否是True,False的
	//error是用来捕捉异常的,,,,,没错误就是nil
	ok, err := regexp.Match(emailRegexPattern, []byte(req.Email))

	if err != nil {
		context.String(http.StatusOK, "系统错误")
		return

	}
	if !ok {
		context.String(http.StatusOK, "你的邮箱格式不对")
		return
	}

	ok, err = regexp.Match(passwordRegexPattern, []byte(req.Password))

	if err != nil {
		fmt.Println(err)
		context.String(http.StatusOK, "系统错误")
		return
	}
	if !ok {
		context.String(http.StatusOK, "你的密码不对,必须大于八位包含特殊字符")
		return

	}
	//err已经声明过了就不要在:=
	err = u.svc.SignUp(context, domain.User{
		Email:    req.Email,
		Password: req.Password,
	})
	if err == service.ErrUserDuplicateEmail {
		context.String(http.StatusOK, "邮箱冲突，请使用另外一个")
		return

	}
	if err != nil {
		context.String(http.StatusOK, "系统错误")
	}

	if req.ConfirmPassword != req.Password {
		context.String(http.StatusOK, "两次输入的密码不一致,请重新输入")
	} else {
		context.String(http.StatusOK, "注册成功")
		fmt.Printf("%v", req)

	}

}

func (u *UserHandler) Login(context *gin.Context) {
	type LoginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	//这个是个类不是实例对象
	var req LoginReq
	if err := context.Bind(&req); err != nil {
		return
	}

	//登录逻辑调用操作
	user, err := u.svc.Login(context, domain.User{
		Email:    req.Email,
		Password: req.Password,
	})
	if err == service.ErrInvalidUserOrPassword {
		context.String(http.StatusOK, "用户名或密码不对")
		return
	}
	if err != nil {
		context.String(http.StatusOK, "系统错误")
		return
	}

	//登录成功之后设置session
	sess := sessions.Default(context)
	sess.Set("userId", user.Id)
	sess.Options(sessions.Options{
		//Secure: true,
		//HttpOnly: true,
		//一分钟过期
		MaxAge: 60,
	})
	sess.Save()
	context.String(http.StatusOK, "登录成功")
	return

}

func (u *UserHandler) LoginJWT(context *gin.Context) {
	type LoginReq struct {
		//定义一个结构类型用于接受前端传递过来的参数，通过json解析对象
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	//这个是个类不是实例对象，通过赋值将他应用
	var req LoginReq
	if err := context.Bind(&req); err != nil {
		return
	}
	//*********************************************************************************************************************************************
	//登录逻辑调用操作
	user, err := u.svc.Login(context, domain.User{
		Email:    req.Email,
		Password: req.Password,
		//不传递ctime没事，反正数据库dao会进行定义
	})
	//特定的变量从dao 返回的错误字段
	if err == service.ErrInvalidUserOrPassword {
		context.String(http.StatusOK, "用户名或密码不对")
		return
	}
	if err != nil {
		context.String(http.StatusOK, "系统错误")
		return
	}

	//*******************************************************************登陆成功****************************************************************************
	//设置jwt登陆状态，生成jwttoken
	//设置登录态，生成token
	//带userID
	claims := UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 30)),
		},
		Uid: user.Id,
		//标识用户的软件和硬件信息
		UserAgent: context.Request.UserAgent(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	tokenStr, err := token.SignedString([]byte("iSQXg9EZhWMbSxYkExJaj4zbflnHCppl"))
	if err != nil {
		context.String(http.StatusInternalServerError, "系统错误")
		return
	}
	//在前响应头中塞进去
	context.Header("x-jwt-token", tokenStr)
	fmt.Println(user)
	fmt.Println(tokenStr)
	context.String(http.StatusOK, "登录成功JWT")
	return

}

func (u *UserHandler) Edit(context *gin.Context) {

}

func (u *UserHandler) Profile(context *gin.Context) {
	//返回的是any类型,直接从请求头忠拿到的数据，拿到的是接口类型，需要类型断言去转化，所以我要发他转化成结构体指针类型
	c, ok := context.Get("claims")
	if !ok {
		context.String(http.StatusOK, "系统错误")
		return
	}
	//是在对获取的值 c 进行类型断言，尝试将其转换为 *UserClaims 类型的指针。ok 检查类型断言是否成功
	claims, ok := c.(*UserClaims)
	if !ok {
		context.String(http.StatusOK, "系统错误")
		return
	}
	println(claims.Uid)
	context.String(http.StatusOK, "这是你的profile")

}

func (u *UserHandler) Logout(context *gin.Context) {
	sess := sessions.Default(context)
	//sess.Set("userId", user.Id)
	sess.Options(sessions.Options{
		//Secure: true,
		//HttpOnly: true,
		//使cookie过期？
		MaxAge: -1,
	})
	sess.Save()
	context.String(http.StatusOK, "退出登录成功")

}

type UserClaims struct {
	jwt.RegisteredClaims
	//声明要放进去token里面的数据
	Uid       int64
	UserAgent string
}

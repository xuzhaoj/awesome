package logger

// 以前定义的是任何类型
type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

// 值传递一个
func LoggerExample() {
	var l Logger
	phone := "122xxx32323"
	l.Info("用户手机没有注册%s", phone)

}

// 现在定义的是自定义结构体类型
type LoggerV1 interface {
	Debug(msg string, args ...Field)
	Info(msg string, args ...Field)
	Warn(msg string, args ...Field)
	Error(msg string, args ...Field)
}
type Field struct {
	Key string
	Val any
}

// 参数有名字
func exampleV1() {
	var l LoggerV1
	// 这是一个新用户 union_id=123
	l.Info("这是一个新用户", Field{Key: "union_id", Val: 123})
}

type LoggerV2 interface {
	// 它要去 args 必须是偶数，并且是以 key1,value1,key2,value2 的形式传递
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

// 参数有名字更规范
func exampleV2() {
	var l LoggerV2
	// 这是一个新用户 union_id=123
	phone := "122xxx32323"
	l.Info("这是一个新用户", "phone", phone)
}

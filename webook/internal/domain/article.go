package domain

import "time"

type Article struct {
	//这个为什么要在新增的时候添加，那么就是添加的时候不用但是更新的时候需要
	Id      int64
	Title   string
	Content string
	Author  Author
	Status  ArticleStatus
	Ctime   time.Time
	Utime   time.Time
}

// rune 是表示单个 Unicode 字符的类型直接操作 string 类型的字节,一个汉字可以占用多个字节
// 可能会导致字符被切成一半（因为一个汉字可能占多个字节）。使用 rune 可以确保每个字符都被正确地处理。
func (a Article) Abstract() string {
	//摘要我们提取前面几句，中文问题是有可能会将字符切成两半的
	cs := []rune(a.Content)
	if len(cs) < 100 {
		return a.Content
	}
	//取 cs 切片的前 100 个字符，
	return string(cs[:100])

}

// 定义一个延伸类型
type ArticleStatus uint8

// iota 是 Go 语言的一个常量生成器
const (
	ArticleStatusUnknown ArticleStatus = iota
	ArticleStatusUnpublished
	ArticleStatusPublished
	ArticleStatusPrivate
)

func (s ArticleStatus) ToUint8() uint8 {
	return uint8(s)
}
func (s ArticleStatus) Valid() bool {
	return s.ToUint8() > 0
}

// 文章的发布状态
func (s ArticleStatus) NonPublish() bool {
	return s != ArticleStatusPublished
}

func (s ArticleStatus) String() string {
	switch s {
	case ArticleStatusPrivate:
		return "private"

	case ArticleStatusUnpublished:
		return "unpublished"
	case ArticleStatusPublished:
		return "published"
	default:
		return "unknown"
	}
}

// 状态复杂有很多行为需要额外的字段就需要视同这个
type ArticleStatusV1 struct {
	Val  uint8
	Name string
}

var (
	ArticleStatusV1Unkown = ArticleStatusV1{
		Val: 0, Name: "unknown",
	}
)

type Author struct {
	Id   int64
	Name string
}

package web

//和前端打交道的articlevo

type ArticleVO struct {
	Id    int64  `json:"id"`
	Title string `json:"title"`
	//在list里面是不会进行赋值的
	Abstract string `json:"abstract"`
	Content  string `json:"content"`
	Author   string `json:"author"`

	Status uint8  `json:"status"`
	Ctime  string `json:"ctime"`
	Utime  string `json:"utime"`
}

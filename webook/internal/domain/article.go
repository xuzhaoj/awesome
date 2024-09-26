package domain

type Article struct {
	//这个为什么要在新增的时候添加，那么就是添加的时候不用但是更新的时候需要
	Id      int64
	Title   string
	Content string
	Author  Author
}

type Author struct {
	Id   int64
	Name string
}

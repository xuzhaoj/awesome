package web

//和前端打交道的articlevo

type ArticleVO struct {
	Id    int64  `json:"id"`
	Title string `json:"title"`
	//在list里面是不会进行赋值的
	Abstract   string `json:"abstract"`
	Content    string `json:"content"`
	Author     string `json:"author"`
	ReadCnt    int64  `json:"read_cnt"`
	LikeCnt    int64  `json:"like_cnt"`
	CollectCnt int64  `json:"collect_cnt"`
	//我个人是否对文章进行了点赞和收藏
	Liked     bool   `json:"liked"`
	Collected bool   `json:"collected"`
	Status    uint8  `json:"status"`
	Ctime     string `json:"ctime"`
	Utime     string `json:"utime"`
}

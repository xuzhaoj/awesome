package integration

import (
	"awesomeProject/webook/internal/domain"
	"awesomeProject/webook/internal/integration/startup"
	"awesomeProject/webook/internal/repository/dao/article"
	"awesomeProject/webook/internal/web"
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
	"net/http"
	"net/http/httptest"
	"testing"
)

// 测试套件
type ArticleTestSuite struct {
	suite.Suite
	server *gin.Engine
	db     *gorm.DB
}

func (s *ArticleTestSuite) SetupSuite() {
	//实现接口
	//s.server = startup.InitWebServer()
	s.server = gin.Default()
	s.server.Use(func(ctx *gin.Context) {
		ctx.Set("claims", &web.UserClaims{
			//构造用户123
			Uid: 123,
		})
	})
	s.db = startup.InitDB()
	artHdl := startup.InitArticleHandler()
	artHdl.RegisterRoutes(s.server)
}

// 清空数据库表,并且从1开始自增
func (s *ArticleTestSuite) TearDownTest() {
	s.db.Exec("TRUNCATE TABLE articles")
	s.db.Exec("TRUNCATE TABLE published_articles")
}

// 测试编辑功能
func (s *ArticleTestSuite) TestEdit() {
	t := s.T()
	testCases := []struct {
		name string

		//集成测试准备数据
		before func(t *testing.T)
		//集成测试验证数据
		after func(t *testing.T)

		//预期中的输入
		art Article
		//预期返回的result
		wantCode int
		wantRes  Result[int64]
	}{
		{
			name: "新建帖子，保存成功",
			before: func(t *testing.T) {

			},
			after: func(t *testing.T) {
				//绕开你的代码去数据库去验正,因为有可能你的代码都是错误的
				var art article.Article
				err := s.db.Where("id=?", 1).First(&art).Error
				assert.NoError(t, err)
				assert.True(t, art.Ctime > 0)
				assert.True(t, art.Utime > 0)
				art.Ctime = 0
				art.Utime = 0
				assert.Equal(t, article.Article{
					Id:       1,
					Title:    "我的标题",
					Content:  "我的内容",
					AuthorId: 123,
					Status:   domain.ArticleStatusUnpublished.ToUint8(),
				}, art)
			},
			//预期输入
			art: Article{
				Title:   "我的标题",
				Content: "我的内容",
			},
			//期望的输出
			wantCode: http.StatusOK,
			wantRes: Result[int64]{
				Data: 1,
				Msg:  "OK",
			},
		},
		{
			name: "修改已有帖子,并保存下来",
			before: func(t *testing.T) {
				//假设数据库已经有数据了，实际的输出
				err := s.db.Create(article.Article{
					Id:       2,
					Title:    "我的标题",
					Content:  "我的内容",
					AuthorId: 123,
					Ctime:    123,
					Utime:    234,
					Status:   domain.ArticleStatusPublished.ToUint8(),
				}).Error
				//断言这个不会有err必然保存成功
				assert.NoError(t, err)

			},
			after: func(t *testing.T) {
				//绕开你的代码去数据库去验正,因为有可能你的代码都是错误的
				var art article.Article
				err := s.db.Where("id=?", 2).First(&art).Error
				assert.NoError(t, err)

				assert.True(t, art.Utime > 234)
				//art.Ctime = 0如果你不设置 art.Utime = 0，断言时会比较 Utime，而因为它的动态更新特性
				art.Utime = 0
				//这个是数据库中查询出来的数据，相当于模拟数据库，预期的输出
				assert.Equal(t, article.Article{
					Id:       2,
					Title:    "新的标题",
					Content:  "新的内容",
					Ctime:    123,
					AuthorId: 123,
					Status:   domain.ArticleStatusUnpublished.ToUint8(),
				}, art)
			},
			//预期输入
			art: Article{
				Id:      2,
				Title:   "新的标题",
				Content: "新的内容",
			},
			//期望的输出
			wantCode: http.StatusOK,
			wantRes: Result[int64]{
				Data: 2,
				Msg:  "OK",
			},
		},
		{
			name: "修改别人的帖子id",
			before: func(t *testing.T) {
				//假设数据库已经有数据了，实际的输出
				err := s.db.Create(article.Article{
					Id:      3,
					Title:   "我的标题",
					Content: "我的内容",
					//测试模拟的用户是123这里是789
					AuthorId: 789,
					Ctime:    123,
					Utime:    234,
					Status:   domain.ArticleStatusPublished.ToUint8(),
				}).Error
				//断言这个不会有err必然保存成功
				assert.NoError(t, err)

			},
			after: func(t *testing.T) {
				//去数据库去验正,因为有可能你的代码都是错误的
				var art article.Article
				err := s.db.Where("id=?", 3).First(&art).Error
				assert.NoError(t, err)
				//这个是数据库中查询出来的数据，相当于模拟数据库，预期的输出
				assert.Equal(t, article.Article{
					Id:       3,
					Title:    "我的标题",
					Content:  "我的内容",
					AuthorId: 789,
					Ctime:    123,
					Utime:    234,
					Status:   domain.ArticleStatusPublished.ToUint8(),
				}, art)
			},
			//预期输入
			art: Article{
				Id:      3,
				Title:   "新的标题",
				Content: "新的内容",
			},
			//期望的输出
			wantCode: http.StatusOK,
			wantRes: Result[int64]{
				Code: 5,
				Msg:  "系统错误",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			//1.构造请求
			//2.执行
			//3.验证结果
			tc.before(t)
			reqBody, err := json.Marshal(tc.art)
			assert.NoError(t, err)
			req, err := http.NewRequest(http.MethodPost,
				"/articles/edit", bytes.NewBuffer([]byte(reqBody)))
			//踩坑---请求的数据是json
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			//构造的请求或者路径不对会返回err

			//返回数据存储的地方
			resp := httptest.NewRecorder()

			//HTTP进入gin框架的入口
			s.server.ServeHTTP(resp, req)

			assert.Equal(t, tc.wantCode, resp.Code)
			if resp.Code != 200 {
				return
			}
			//验证HTTP请求返回的响应码是否与wancode相同
			var webRes Result[int64]
			err = json.NewDecoder(resp.Body).Decode(&webRes)
			require.NoError(t, err)
			assert.Equal(t, tc.wantRes, webRes)
			tc.after(t)
		})
	}

}

// 测试上线功能
func (s *ArticleTestSuite) TestPublish() {
	t := s.T()

	testCases := []struct {
		name string
		// 要提前准备数据
		before func(t *testing.T)
		// 验证并且删除数据
		after func(t *testing.T)
		req   Article

		// 预期响应
		wantCode   int
		wantResult Result[int64]
	}{
		{
			name: "新建帖子并发表",
			//设置数据库的初始值
			before: func(t *testing.T) {
				// 什么也不需要做
			},
			//它绕过业务逻辑代码，直接查询数据库
			after: func(t *testing.T) {
				// 存储从数据库 查询出来的数据
				var art article.Article
				err := s.db.Where("author_id = ?", 123).First(&art).Error
				assert.NoError(t, err)
				assert.True(t, art.Id > 0)
				assert.True(t, art.Ctime > 0)
				assert.True(t, art.Utime > 0)
				//通过重置这些值，可以确保后续的断言只关注其他重要的字段内容，而不受时间戳和 ID 的影响。
				art.Ctime = 0
				art.Utime = 0
				art.Id = 0
				//后面比较的时候就不比较后面的值
				//断言查询出来的数据的值应该等于
				assert.Equal(t, article.Article{
					Title:    "hello，你好",
					Content:  "随便试试",
					AuthorId: 123,
					Status:   domain.ArticleStatusPublished.ToUint8(),
				}, art)

				var publishedArt article.PublishedArticle
				err = s.db.Where("author_id = ?", 123).First(&publishedArt).Error
				assert.NoError(t, err)
				//可以直接省略，直接对最内层的id进行调用
				assert.True(t, publishedArt.Id > 0)
				assert.True(t, publishedArt.Ctime > 0)
				assert.True(t, publishedArt.Utime > 0)
				publishedArt.Ctime = 0
				publishedArt.Utime = 0
				publishedArt.Id = 0
				assert.Equal(t, article.PublishedArticle{
					article.Article{
						Title:    "hello，你好",
						Content:  "随便试试",
						AuthorId: 123,
						Status:   domain.ArticleStatusPublished.ToUint8(),
					},
				}, publishedArt)
			},
			//测试中传递给 API 的输入数据，模拟实际请求
			req: Article{
				Title:   "hello，你好",
				Content: "随便试试",
			},
			//表示你对 API 请求的预期 HTTP 响应码。
			wantCode: 200,
			//wantResult 是对响应结果的期望。它的类型是 Result[int64]，代表返回的结果格式，
			wantResult: Result[int64]{
				Msg:  "OK",
				Data: 1,
			},
		},
		{
			// 制作库有，但是线上库没有  类似于保存到草稿箱在重新发表
			name: "更新帖子并新发表",
			before: func(t *testing.T) {
				// 模拟已经存在的帖子
				err := s.db.Create(&article.Article{
					Id:       2,
					Title:    "我的标题",
					Content:  "我的内容",
					Ctime:    456,
					Status:   domain.ArticleStatusUnpublished.ToUint8(),
					Utime:    234,
					AuthorId: 123,
				}).Error
				assert.NoError(t, err)
			},

			after: func(t *testing.T) {
				// 验证一下更新之后的数据
				var art article.Article
				s.db.Where("id = ?", 2).First(&art)
				// 更新时间变了，比较不了是当前的时间搓
				assert.True(t, art.Utime > 234)
				art.Utime = 0
				//assert.Equal(t, domain.ArticleStatusPublished.ToUint8(), art.Status)
				assert.Equal(t, article.Article{
					Id: 2,
					// 创建时间没变，本来你的数据库的插入的时间就不会变你更新的只是utime
					Ctime:    456,
					Status:   domain.ArticleStatusPublished.ToUint8(),
					Content:  "新的内容",
					Title:    "新的标题",
					AuthorId: 123,
				}, art)
				//验证一下发布到线上库的文章记录
				var publishedArt article.PublishedArticle
				s.db.Where("id = ?", 2).First(&publishedArt)

				assert.True(t, publishedArt.Ctime > 0)
				assert.True(t, publishedArt.Utime > 0)
				publishedArt.Ctime = 0
				publishedArt.Utime = 0
				assert.Equal(t, article.PublishedArticle{
					Article: article.Article{
						Id: 2,
						// 创建时间发布到线上是当前时间戳你比较不了
						//Ctime:    456,
						Status:   domain.ArticleStatusPublished.ToUint8(),
						Content:  "新的内容",
						Title:    "新的标题",
						AuthorId: 123,
					},
				}, publishedArt)
			},
			req: Article{
				Id:      2,
				Title:   "新的标题",
				Content: "新的内容",
			},
			wantCode: 200,
			wantResult: Result[int64]{
				Msg:  "OK",
				Data: 2,
			},
		},
		{
			name: "更新帖子，并且重新发表",
			before: func(t *testing.T) {
				art := article.Article{
					Id:       3,
					Title:    "我的标题",
					Content:  "我的内容",
					Ctime:    456,
					Status:   domain.ArticleStatusPublished.ToUint8(),
					Utime:    234,
					AuthorId: 123,
				}
				//准备好数据插入表中
				err := s.db.Create(&art).Error
				assert.NoError(t, err)
				part := article.PublishedArticle{
					Article: art,
				}
				//准备好数据插入表中
				err = s.db.Create(&part).Error
				assert.NoError(t, err)

			},
			after: func(t *testing.T) {
				var art article.Article
				err := s.db.Where("id = ?", 3).First(&art).Error
				assert.NoError(t, err)
				assert.True(t, art.Utime > 234)
				//这个art。Utime是一个无法预知的值设置为0不去比较
				art.Utime = 0
				//assert.Equal(t, domain.ArticleStatusPublished.ToUint8(), art.Status)
				assert.Equal(t, article.Article{
					Id: 3,
					// 创建时间没变，本来你的数据库的插入的时间就不会变你更新的只是utime
					Ctime:    456,
					Status:   domain.ArticleStatusPublished.ToUint8(),
					Content:  "新的内容",
					Title:    "新的标题",
					AuthorId: 123,
				}, art)
				var publishedArt article.PublishedArticle
				err = s.db.Where("id = ?", 3).First(&publishedArt).Error
				assert.NoError(t, err)
				assert.True(t, publishedArt.Ctime > 0)
				assert.True(t, publishedArt.Utime > 0)
				publishedArt.Ctime = 0
				publishedArt.Utime = 0
				assert.Equal(t, article.PublishedArticle{
					Article: article.Article{
						Id: 3,
						// 创建时间发布到线上是当前时间戳你比较不了，重新发布创建时间都要变掉
						//Ctime:    456,
						Status:   domain.ArticleStatusPublished.ToUint8(),
						Content:  "新的内容",
						Title:    "新的标题",
						AuthorId: 123,
					},
				}, publishedArt)
			},
			req: Article{
				Id:      3,
				Title:   "新的标题",
				Content: "新的内容",
			},
			wantCode: 200,
			wantResult: Result[int64]{
				Msg:  "OK",
				Data: 3,
			},
		},
		{
			name: "更新别人的帖子，并且发表失败",
			before: func(t *testing.T) {
				art := article.Article{
					//帖子id
					Id:      4,
					Title:   "我的标题",
					Content: "我的内容",
					Ctime:   456,
					Utime:   234,
					Status:  domain.ArticleStatusUnpublished.ToUint8(),
					// 我们是作者123，我们想去模拟篡改789的文章，串改其还未同步发表的地方
					AuthorId: 789,
				}
				s.db.Create(&art)
				part := article.PublishedArticle{
					Article: article.Article{
						Id:       4,
						Title:    "我的标题",
						Content:  "我的内容",
						Ctime:    456,
						Status:   domain.ArticleStatusPublished.ToUint8(),
						Utime:    234,
						AuthorId: 789,
					},
				}
				s.db.Create(&part)
			},
			after: func(t *testing.T) {

				// 更新应该是失败了，数据没有发生变化
				var art article.Article
				s.db.Where("id = ?", 4).First(&art)
				//assert.Equal(t, "我的标题", art.Title)
				//assert.Equal(t, "我的内容", art.Content)
				//assert.Equal(t, int64(456), art.Ctime)
				//assert.Equal(t, int64(234), art.Utime)
				//assert.Equal(t, uint8(1), art.Status)
				//assert.Equal(t, int64(789), art.AuthorId)
				assert.Equal(t, article.Article{
					Id: 4,
					// 创建时间没变，本来你的数据库的插入的时间就不会变你更新的只是utime
					Ctime:    456,
					Status:   domain.ArticleStatusUnpublished.ToUint8(),
					Content:  "我的内容",
					Title:    "我的标题",
					AuthorId: 789,
					Utime:    234,
				}, art)

				var part article.PublishedArticle
				// 数据没有变化
				s.db.Where("id = ?", 4).First(&part)
				//assert.Equal(t, "我的标题", part.Title)
				//assert.Equal(t, "我的内容", part.Content)
				//assert.Equal(t, int64(789), part.AuthorId)
				//assert.Equal(t, uint8(2), part.Status)
				// 创建时间没变
				//assert.Equal(t, int64(456), part.Ctime)
				// 更新时间变了
				//assert.Equal(t, int64(234), part.Utime)
				assert.Equal(t, article.PublishedArticle{
					Article: article.Article{
						Id:       4,
						Ctime:    456,
						Status:   domain.ArticleStatusPublished.ToUint8(),
						Content:  "我的内容",
						Title:    "我的标题",
						AuthorId: 789,
						Utime:    234,
					},
				}, part)
			},
			req: Article{
				Id:      4,
				Title:   "新的标题",
				Content: "新的内容",
			},
			wantCode: 200,
			wantResult: Result[int64]{
				Code: 5,
				Msg:  "系统错误",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			data, err := json.Marshal(tc.req)
			// 不能有 error
			assert.NoError(t, err)
			req, err := http.NewRequest(http.MethodPost,
				"/articles/publish", bytes.NewReader(data))
			assert.NoError(t, err)
			req.Header.Set("Content-Type",
				"application/json")
			recorder := httptest.NewRecorder()

			s.server.ServeHTTP(recorder, req)
			code := recorder.Code
			assert.Equal(t, tc.wantCode, code)
			if code != http.StatusOK {
				return
			}
			// 反序列化为结果
			// 利用泛型来限定结果必须是 int64
			var result Result[int64]
			err = json.Unmarshal(recorder.Body.Bytes(), &result)
			assert.NoError(t, err)
			assert.Equal(t, tc.wantResult, result)
			tc.after(t)
		})
	}

}

// 用来Json序列化（Marshal）和反序列化NewDecode
type Article struct {
	Id      int64  `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

// 反序列化的时候会出现一些问题，所以转泛型
type Result[T any] struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data T      `json:"data"`
}

func (s *ArticleTestSuite) TestABC() {
	s.T().Log("hello,这是测试套件")
}
func TestArticle(t *testing.T) {
	suite.Run(t, &ArticleTestSuite{})
}

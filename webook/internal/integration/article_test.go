package integration

import (
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
}
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

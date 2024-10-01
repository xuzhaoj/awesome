package cache

import (
	"awesomeProject/webook/internal/domain"
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

type ArticleCache interface {
	GetFirstPage(ctx context.Context, author int64) ([]domain.Article, error)
	SetFirstPage(ctx context.Context, author int64, arts []domain.Article) error
	DelFirstPage(ctx context.Context, author int64) error
	Set(ctx context.Context, art domain.Article) error

	SetPub(ctx context.Context, art domain.Article) error
	GetPub(ctx context.Context, id int64) (domain.Article, error)
}

type RedisArticleCache struct {
	client redis.Cmdable
}

// 单个的文章详情进行缓存
func (r *RedisArticleCache) SetPub(ctx context.Context, art domain.Article) error {
	//这种存一个文章的随便看啊
	val, err := json.Marshal(art)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, r.pubKey(art.Id), val, time.Minute*10).Err()
}

// 查询文章详情进行缓存
func (r *RedisArticleCache) GetPub(ctx context.Context, id int64) (domain.Article, error) {
	val, err := r.client.Get(ctx, r.pubKey(id)).Bytes()
	if err != nil {
		return domain.Article{}, err
	}
	var res domain.Article
	err = json.Unmarshal(val, &res)
	return res, err
}
func (r *RedisArticleCache) pubKey(id int64) string {
	return fmt.Sprintf("article:pub:detail:%d", id)
}

func (r *RedisArticleCache) Set(ctx context.Context, art domain.Article) error {
	//将数据转化为json格式便于缓存到redis中
	data, err := json.Marshal(art)
	if err != nil {
		return err
	}
	//过期时间要短，你的预测效果越不好，就要越短
	return r.client.Set(ctx, r.key(art.Id), data, time.Minute).Err()
}
func (r *RedisArticleCache) key(id int64) string {
	//"article:first_page:12345"
	return fmt.Sprintf("article:%d", id)
}

// 得到第一页的缓存数据
func (r *RedisArticleCache) GetFirstPage(ctx context.Context, author int64) ([]domain.Article, error) {
	//json进行数据转化的时候需要传递的是字节数据，你这样返回的是字符串数据
	//data, err := r.client.Get(ctx, r.firstPageKey(author)).Result()

	//将字符串数据转化成字节切片
	bs, err := r.client.Get(ctx, r.firstPageKey(author)).Bytes()
	if err != nil {
		return nil, err
	}
	var ants []domain.Article
	//Redis存储的是JSON格式的字符串，我们需要获取到的json格式反序列化映射到结构体上面
	err = json.Unmarshal(bs, &ants)
	return ants, err

}

func (r *RedisArticleCache) SetFirstPage(ctx context.Context, author int64, arts []domain.Article) error {
	//内容不需要返回太多的，只需要返回一些摘要即可，所以需要重新赋值
	for i := 0; i < len(arts); i++ {
		arts[i].Content = arts[i].Abstract()
	}
	//将数据转化成json格式redis中
	data, err := json.Marshal(arts)
	if err != nil {
		return err
	}
	//设置缓存过期时间
	return r.client.Set(ctx, r.firstPageKey(author), data, time.Minute*10).Err()
}
func (r *RedisArticleCache) firstPageKey(uid int64) string {
	//"article:first_page:12345"
	return fmt.Sprintf("article:first_page:%d", uid)
}

func (r *RedisArticleCache) DelFirstPage(ctx context.Context, author int64) error {
	//实现接口中的方法
	return r.client.Del(ctx, r.firstPageKey(author)).Err()
}

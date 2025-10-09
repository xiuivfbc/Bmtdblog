package models

import (
	"database/sql"
	"fmt"
	"html/template"
	"log/slog"
	"math/rand"
	"time"

	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday"
	"github.com/xiuivfbc/bmtdblog/system"
)

type Post struct {
	ID           uint       `gorm:"primarykey"`
	CreatedAt    *time.Time `gorm:"autoCreateTime"`
	UpdatedAt    *time.Time `gorm:"autoUpdateTime"`
	Title        string     `gorm:"type:text"`
	Body         string     `gorm:"type:longtext"`
	View         int
	IsPublished  bool
	Tags         []*Tag     `gorm:"-"`
	Comments     []*Comment `gorm:"-"`
	CommentTotal int        `gorm:"->"`
}

type QrArchive struct {
	ArchiveDate time.Time //month
	Total       int       //total
	Year        int       // year
	Month       int       // month
}

func (post *Post) Insert() error {
	err := DB.Create(post).Error
	if err != nil {
		return err
	}

	// 清除可能存在的空值缓存
	go func() {
		if err := DeleteNullCache(post.ID); err != nil {
			slog.Error("Failed to delete null cache after insert", "id", post.ID, "error", err)
		}
	}()

	// 清除相关缓存
	go post.ClearRelatedCache()

	return nil
}

func (post *Post) Update() error {
	err := DB.Model(post).Updates(map[string]interface{}{
		"title":        post.Title,
		"body":         post.Body,
		"is_published": post.IsPublished,
	}).Error
	if err != nil {
		return err
	}

	// 清除相关缓存
	go post.ClearRelatedCache()

	return nil
}

func (post *Post) UpdateView() error {
	return DB.Model(post).Updates(map[string]interface{}{
		"view": post.View,
	}).Error
}

func (post *Post) Delete() error {
	err := DB.Delete(post).Error
	if err != nil {
		return err
	}

	// 清除相关缓存
	go post.ClearRelatedCache()

	return nil
}

// 缓存相关常量
const (
	// 单个博文缓存key前缀
	PostCachePrefix = "post"
	// 博文列表缓存key前缀
	PostListCachePrefix = "post_list"
	// 归档缓存key前缀
	PostArchiveCachePrefix = "post_archive"
	// 缓存过期时间
	PostCacheExpiration     = 1 * time.Hour
	PostListCacheExpiration = 30 * time.Minute
	// 防雪崩：随机偏移量最大值
	PostCacheRandomOffset     = 30 * time.Minute // 博文缓存 ±30分钟
	PostListCacheRandomOffset = 10 * time.Minute // 列表缓存 ±10分钟

	// 防穿透：空值缓存相关
	NullCachePrefix     = "null_post"     // 空值缓存key前缀
	NullCacheExpiration = 5 * time.Minute // 空值缓存过期时间（较短）
)

// 生成博文缓存key
func (post *Post) CacheKey() string {
	return system.GenerateKey(PostCachePrefix, post.ID)
}

// 生成随机过期时间（防缓存雪崩）
func getRandomExpiration(baseExpiration time.Duration, randomOffset time.Duration) time.Duration {
	// 生成 -randomOffset 到 +randomOffset 之间的随机偏移量
	offsetRange := int64(randomOffset * 2)
	randomOffsetValue := time.Duration(rand.Int63n(offsetRange)) - randomOffset
	return baseExpiration + randomOffsetValue
}

// 从缓存获取博文
func GetPostFromCache(id uint) (*Post, error) {
	if system.Redis == nil || !system.Redis.IsAvailable() {
		return nil, fmt.Errorf("cache not available")
	}

	key := system.GenerateKey(PostCachePrefix, id)
	var post Post

	err := system.Redis.Get(key, &post)
	if err != nil {
		return nil, err
	}

	slog.Debug("Post loaded from cache", "id", id, "title", post.Title)
	return &post, nil
}

// 将博文存入缓存
func (post *Post) SetCache() error {
	if system.Redis == nil || !system.Redis.IsAvailable() {
		return nil // 缓存不可用时不报错，静默失败
	}

	key := post.CacheKey()
	// 使用随机过期时间防止缓存雪崩
	randomExpiration := getRandomExpiration(PostCacheExpiration, PostCacheRandomOffset)
	err := system.Redis.Set(key, post, randomExpiration)
	if err != nil {
		slog.Error("Failed to cache post", "id", post.ID, "error", err)
		return err
	}

	slog.Debug("Post cached successfully", "id", post.ID, "title", post.Title, "expiration", randomExpiration)
	return nil
}

// 删除博文缓存
func (post *Post) DelCache() error {
	if system.Redis == nil || !system.Redis.IsAvailable() {
		return nil
	}

	key := post.CacheKey()
	err := system.Redis.Del(key)
	if err != nil {
		slog.Error("Failed to delete post cache", "id", post.ID, "error", err)
		return err
	}

	slog.Debug("Post cache deleted", "id", post.ID)
	return nil
}

// 清除所有相关缓存（博文增删改时调用）
func (post *Post) ClearRelatedCache() error {
	if system.Redis == nil || !system.Redis.IsAvailable() {
		return nil
	}

	// 清除自身缓存
	if err := post.DelCache(); err != nil {
		slog.Error("Failed to clear post cache", "id", post.ID, "error", err)
	}

	// 清除列表缓存（影响首页、归档等）
	patterns := []string{
		PostListCachePrefix + "*",
		PostArchiveCachePrefix + "*",
		"index*", // 首页缓存
	}

	for _, pattern := range patterns {
		if err := system.Redis.DelPattern(pattern); err != nil {
			slog.Error("Failed to clear cache pattern", "pattern", pattern, "error", err)
		}
	}

	slog.Debug("Related cache cleared for post", "id", post.ID)
	return nil
}

// 带缓存的获取博文方法（防穿透：空值缓存）
func GetPostByIdWithCache(id uint) (*Post, error) {
	// 1. 先尝试从缓存获取
	if post, err := GetPostFromCache(id); err == nil {
		// 补充关联数据（Tags, Comments等）
		if err := LoadPostRelations(post); err != nil {
			slog.Error("Failed to load post relations from cache", "id", id, "error", err)
		}
		return post, nil
	}

	// 2. 缓存未命中，从数据库获取
	post, err := GetPostById(id)
	if err != nil {
		// 数据库中也不存在，设置空值缓存（避免重复查询）
		go func() {
			if err := SetNullCache(id); err != nil {
				slog.Error("Failed to set null cache async", "id", id, "error", err)
			}
		}()
		return nil, err
	}

	// 3. 异步写入缓存
	go func() {
		if err := post.SetCache(); err != nil {
			slog.Error("Failed to set post cache async", "id", id, "error", err)
		}
	}()

	return post, nil
}

// 加载博文关联数据（标签、评论等）
func LoadPostRelations(post *Post) error {
	// 加载标签
	if tags, err := ListTagByPostId(post.ID); err == nil {
		post.Tags = tags
	}

	// 加载评论
	if comments, err := ListCommentByPostID(post.ID); err == nil {
		post.Comments = comments
		post.CommentTotal = len(comments)
	}

	return nil
}

func (post *Post) Excerpt() template.HTML {
	//you can sanitize, cut it down, add images, etc
	policy := bluemonday.StrictPolicy() //remove all html tags
	sanitized := policy.Sanitize(string(blackfriday.MarkdownCommon([]byte(post.Body))))
	runes := []rune(sanitized)
	if len(runes) > 300 {
		sanitized = string(runes[:300])
	}
	excerpt := template.HTML(sanitized + "...")
	return excerpt
}

func ListPublishedPost(tag string, pageIndex, pageSize int) ([]*Post, error) {
	return _listPost(tag, true, pageIndex, pageSize)
}

func ListAllPost(tag string) ([]*Post, error) {
	return _listPost(tag, false, 0, 0)
}

func _listPost(tagId string, published bool, pageIndex, pageSize int) ([]*Post, error) {
	var posts []*Post
	var err error
	if len(tagId) > 0 {
		var rows *sql.Rows
		if published {
			if pageIndex > 0 {
				rows, err = DB.Raw("select p.* from posts p inner join post_tags pt on p.id = pt.post_id where pt.tag_id = ? and p.is_published = ? order by created_at desc limit ? offset ?", tagId, true, pageSize, (pageIndex-1)*pageSize).Rows()
			} else {
				rows, err = DB.Raw("select p.* from posts p inner join post_tags pt on p.id = pt.post_id where pt.tag_id = ? and p.is_published = ? order by created_at desc", tagId, true).Rows()
			}
		} else {
			rows, err = DB.Raw("select p.* from posts p inner join post_tags pt on p.id = pt.post_id where pt.tag_id = ? order by created_at desc", tagId).Rows()
		}
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		for rows.Next() {
			var post Post
			DB.ScanRows(rows, &post)
			posts = append(posts, &post)
		}
	} else {
		if published {
			if pageIndex > 0 {
				err = DB.Where("is_published = ?", true).Order("created_at desc").Limit(pageSize).Offset((pageIndex - 1) * pageSize).Find(&posts).Error
			} else {
				err = DB.Where("is_published = ?", true).Order("created_at desc").Find(&posts).Error
			}
		} else {
			err = DB.Order("created_at desc").Find(&posts).Error
		}
	}
	return posts, err
}

func MustListMaxReadPost() (posts []*Post) {
	posts, _ = ListMaxReadPost()
	return
}

func ListMaxReadPost() (posts []*Post, err error) {
	err = DB.Where("is_published = ?", true).Order("view desc").Limit(5).Find(&posts).Error
	return
}

func MustListMaxCommentPost() (posts []*Post) {
	posts, _ = ListMaxCommentPost()
	return
}

func ListMaxCommentPost() (posts []*Post, err error) {
	var (
		rows *sql.Rows
	)
	rows, err = DB.Raw("select p.*,c.total comment_total from posts p inner join (select post_id,count(*) total from comments group by post_id) c on p.id = c.post_id order by c.total desc limit 5").Rows()
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var post Post
		DB.ScanRows(rows, &post)
		posts = append(posts, &post)
	}
	return
}

func CountPostByTag(tagId string) (count int, err error) {
	if len(tagId) > 0 {
		err = DB.Raw("select count(*) from posts p inner join post_tags pt on p.id = pt.post_id where pt.tag_id = ? and p.is_published = ?", tagId, true).Row().Scan(&count)
	} else {
		err = DB.Raw("select count(*) from posts p where p.is_published = ?", true).Row().Scan(&count)
	}
	return
}

func CountPost() int64 {
	var count int64
	DB.Model(&Post{}).Count(&count)
	return count
}

func GetPostById(id uint) (*Post, error) {
	var post Post
	err := DB.First(&post, "id = ?", id).Error
	return &post, err
}

func MustListPostArchives() []*QrArchive {
	archives, _ := ListPostArchives()
	return archives
}

func ListPostArchives() ([]*QrArchive, error) {
	var (
		archives []*QrArchive
		querySql string
	)
	querySql = `select date_format(created_at,'%Y-%m') as month,count(*) as total from posts where is_published = ? group by month order by month desc`
	rows, err := DB.Raw(querySql, true).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var archive QrArchive
		var month string
		rows.Scan(&month, &archive.Total)
		//DB.ScanRows(rows, &archive)
		archive.ArchiveDate, _ = time.Parse("2006-01", month)
		archive.Year = archive.ArchiveDate.Year()
		archive.Month = int(archive.ArchiveDate.Month())
		archives = append(archives, &archive)
	}
	return archives, nil
}

func ListPostByArchive(year, month string, pageIndex, pageSize int) ([]*Post, error) {
	var (
		rows     *sql.Rows
		err      error
		querySql string
	)
	if len(month) == 1 {
		month = "0" + month
	}
	condition := fmt.Sprintf("%s-%s", year, month)
	if pageIndex > 0 {
		querySql = `select * from posts where date_format(created_at,'%Y-%m') = ? and is_published = ? order by created_at desc limit ? offset ?`
		rows, err = DB.Raw(querySql, condition, true, pageSize, (pageIndex-1)*pageSize).Rows()
	} else {
		querySql = `select * from posts where date_format(created_at,'%Y-%m') = ? and is_published = ? order by created_at desc`
		rows, err = DB.Raw(querySql, condition, true).Rows()
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	posts := make([]*Post, 0)
	for rows.Next() {
		var post Post
		DB.ScanRows(rows, &post)
		posts = append(posts, &post)
	}
	return posts, nil
}

func CountPostByArchive(year, month string) (count int, err error) {
	var querySql string
	if len(month) == 1 {
		month = "0" + month
	}
	condition := fmt.Sprintf("%s-%s", year, month)
	querySql = `select count(*) from posts where date_format(created_at,'%Y-%m') = ? and is_published = ?`
	err = DB.Raw(querySql, condition, true).Row().Scan(&count)
	return
}

// 缓存辅助方法：为列表数据设置缓存（使用随机过期时间）
func SetListCache(key string, data interface{}) error {
	if system.Redis == nil || !system.Redis.IsAvailable() {
		return nil // 缓存不可用时不报错，静默失败
	}

	// 使用随机过期时间防止缓存雪崩
	randomExpiration := getRandomExpiration(PostListCacheExpiration, PostListCacheRandomOffset)
	err := system.Redis.Set(key, data, randomExpiration)
	if err != nil {
		slog.Error("Failed to cache list data", "key", key, "error", err)
		return err
	}

	slog.Debug("List cache set successfully", "key", key, "expiration", randomExpiration)
	return nil
}

// 缓存辅助方法：获取列表缓存数据
func GetListCache(key string, dest interface{}) error {
	if system.Redis == nil || !system.Redis.IsAvailable() {
		return fmt.Errorf("cache not available")
	}

	err := system.Redis.Get(key, dest)
	if err != nil {
		return err
	}

	slog.Debug("List cache hit", "key", key)
	return nil
}

// ============== 空值缓存相关函数 ==============

// SetNullCache 设置空值缓存
func SetNullCache(postID uint) error {
	if system.Redis == nil || !system.Redis.IsAvailable() {
		return nil
	}

	key := system.GenerateKey(NullCachePrefix, postID)
	// 缓存一个简单的标记，表示该ID不存在
	err := system.Redis.Set(key, "null", NullCacheExpiration)
	if err != nil {
		slog.Error("Failed to set null cache", "post_id", postID, "error", err)
		return err
	}

	slog.Debug("Null cache set", "post_id", postID, "expiration", NullCacheExpiration)
	return nil
}

// DeleteNullCache 删除空值缓存（当博文被创建时调用）
func DeleteNullCache(postID uint) error {
	if system.Redis == nil || !system.Redis.IsAvailable() {
		return nil
	}

	key := system.GenerateKey(NullCachePrefix, postID)
	err := system.Redis.Del(key)
	if err != nil {
		slog.Error("Failed to delete null cache", "post_id", postID, "error", err)
		return err
	}

	slog.Debug("Null cache deleted", "post_id", postID)
	return nil
}

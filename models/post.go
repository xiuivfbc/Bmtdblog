package models

import (
	"database/sql"
	"fmt"
	"html/template"
	"log/slog"
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
)

// 生成博文缓存key
func (post *Post) CacheKey() string {
	return system.GenerateKey(PostCachePrefix, post.ID)
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
	err := system.Redis.Set(key, post, PostCacheExpiration)
	if err != nil {
		slog.Error("Failed to cache post", "id", post.ID, "error", err)
		return err
	}

	slog.Debug("Post cached successfully", "id", post.ID, "title", post.Title)
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

// 带缓存的获取博文方法
func GetPostByIdWithCache(id uint) (*Post, error) {
	// 先尝试从缓存获取
	if post, err := GetPostFromCache(id); err == nil {
		// 补充关联数据（Tags, Comments等）
		if err := LoadPostRelations(post); err != nil {
			slog.Error("Failed to load post relations from cache", "id", id, "error", err)
		}
		return post, nil
	}

	// 缓存未命中，从数据库获取
	post, err := GetPostById(id)
	if err != nil {
		return nil, err
	}

	// 异步写入缓存
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

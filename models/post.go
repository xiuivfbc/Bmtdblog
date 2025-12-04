package models

import (
	"database/sql"
	"fmt"
	"html/template"
	
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
			system.Logger.Error("Failed to delete null cache after insert", "id", post.ID, "error", err)
		}
	}()

	// 同步到ES索引（使用批量队列）
	go func() {
		AddESTask(ESTask{
			Action: "index",
			PostID: post.ID,
			Post:   post,
		})
	}()

	// 清除相关缓存
	go post.ClearRelatedCache()

	return nil
}

func (post *Post) Update() error {
	// 延迟双删策略：第一步 - 先删除缓存
	if err := post.DelayedDoubleDel(); err != nil {
		system.Logger.Error("Failed to initiate delayed double delete", "id", post.ID, "error", err)
	}

	// 更新数据库
	err := DB.Model(post).Updates(map[string]interface{}{
		"title":        post.Title,
		"body":         post.Body,
		"is_published": post.IsPublished,
	}).Error
	if err != nil {
		return err
	}

	// 同步到ES索引（使用批量队列）
	go func() {
		AddESTask(ESTask{
			Action: "index",
			PostID: post.ID,
			Post:   post,
		})
	}()

	// 注意：延迟双删已经处理了缓存清理，不需要再调用ClearRelatedCache

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

	// 从ES索引中删除（使用批量队列）
	go func() {
		AddESTask(ESTask{
			Action: "delete",
			PostID: post.ID,
			Post:   nil,
		})
	}()

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

	// 防击穿：分布式锁相关
	LockCachePrefix   = "lock_post"            // 分布式锁key前缀
	LockCacheTimeout  = 10 * time.Second       // 锁超时时间（防死锁）
	LockRetryInterval = 100 * time.Millisecond // 重试间隔
	LockMaxRetries    = 30                     // 最大重试次数（总等待时间约3秒）
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

	system.Logger.Debug("Post loaded from cache", "id", id, "title", post.Title)
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
		system.Logger.Error("Failed to cache post", "id", post.ID, "error", err)
		return err
	}

	system.Logger.Debug("Post cached successfully", "id", post.ID, "title", post.Title, "expiration", randomExpiration)
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
		system.Logger.Error("Failed to delete post cache", "id", post.ID, "error", err)
		return err
	}

	system.Logger.Debug("Post cache deleted", "id", post.ID)
	return nil
}

// 清除所有相关缓存（博文增删改时调用）
func (post *Post) ClearRelatedCache() error {
	if system.Redis == nil || !system.Redis.IsAvailable() {
		return nil
	}

	// 清除自身缓存
	if err := post.DelCache(); err != nil {
		system.Logger.Error("Failed to clear post cache", "id", post.ID, "error", err)
	}

	// 清除列表缓存（影响首页、归档等）
	patterns := []string{
		PostListCachePrefix + "*",
		PostArchiveCachePrefix + "*",
		"index*", // 首页缓存
	}

	for _, pattern := range patterns {
		if err := system.Redis.DelPattern(pattern); err != nil {
			system.Logger.Error("Failed to clear cache pattern", "pattern", pattern, "error", err)
		}
	}

	system.Logger.Debug("Related cache cleared for post", "id", post.ID)
	return nil
}

// 延迟双删策略：解决缓存一致性问题
func (post *Post) DelayedDoubleDel() error {
	if system.Redis == nil || !system.Redis.IsAvailable() {
		return nil
	}

	// 收集需要删除的缓存键
	keys := []string{
		post.CacheKey(), // 自身缓存
	}

	// 收集相关缓存模式的具体键
	patterns := []string{
		PostListCachePrefix + "*",
		PostArchiveCachePrefix + "*",
		"index*", // 首页缓存
	}

	// 获取匹配模式的具体键名
	for _, pattern := range patterns {
		matchedKeys, err := system.Redis.GetKeysByPattern(pattern)
		if err != nil {
			system.Logger.Error("Failed to get keys by pattern", "pattern", pattern, "error", err)
			continue
		}
		keys = append(keys, matchedKeys...)
	}

	// 执行延迟双删，延迟500ms
	delay := 500 * time.Millisecond
	return system.Redis.DelayedDoubleDelete(keys, delay)
}

// 带缓存的获取博文方法（防穿透+防击穿）
func GetPostByIdWithCache(id uint) (*Post, error) {
	// 1. 检查空值缓存：如果之前确认不存在，直接返回
	if IsNullCached(id) {
		system.Logger.Debug("Null cache hit", "post_id", id)
		return nil, fmt.Errorf("post not found")
	}

	// 2. 先尝试从缓存获取
	if post, err := GetPostFromCache(id); err == nil {
		// 补充关联数据（Tags, Comments等）
		if err := LoadPostRelations(post); err != nil {
			system.Logger.Error("Failed to load post relations from cache", "id", id, "error", err)
		}
		return post, nil
	}

	// 3. 缓存未命中，尝试获取分布式锁（防击穿）
	locked, err := TryLock(id)
	if err != nil {
		system.Logger.Error("Failed to try lock", "post_id", id, "error", err)
		// 锁操作失败，降级为直接查询数据库
		return fallbackGetPost(id)
	}

	if !locked {
		// 未获得锁，说明其他线程正在重建缓存，等待缓存重建完成
		system.Logger.Debug("Lock held by other thread, waiting", "post_id", id)
		if post, err := WaitForLockRelease(id); err == nil {
			// 等待成功，其他线程已重建缓存
			if err := LoadPostRelations(post); err != nil {
				system.Logger.Error("Failed to load post relations after wait", "id", id, "error", err)
			}
			return post, nil
		}
		// 等待超时，降级为直接查询数据库
		system.Logger.Debug("Wait timeout, fallback to direct query", "post_id", id)
		return fallbackGetPost(id)
	}

	// 4. 获得锁，负责重建缓存
	system.Logger.Debug("Got lock, rebuilding cache", "post_id", id)
	defer ReleaseLock(id) // 确保锁被释放

	// 再次检查缓存（双重检查模式）
	if post, err := GetPostFromCache(id); err == nil {
		system.Logger.Debug("Cache rebuilt by other thread (double check)", "post_id", id)
		if err := LoadPostRelations(post); err != nil {
			system.Logger.Error("Failed to load post relations from double check", "id", id, "error", err)
		}
		return post, nil
	}

	// 从数据库获取并重建缓存
	return rebuildCacheAndReturn(id)
}

// fallbackGetPost 降级方案：直接查询数据库（不使用缓存）
func fallbackGetPost(id uint) (*Post, error) {
	post, err := GetPostById(id)
	if err != nil {
		return nil, err
	}

	if err := LoadPostRelations(post); err != nil {
		system.Logger.Error("Failed to load post relations in fallback", "id", id, "error", err)
	}

	return post, nil
}

// rebuildCacheAndReturn 重建缓存并返回数据
func rebuildCacheAndReturn(id uint) (*Post, error) {
	post, err := GetPostById(id)
	if err != nil {
		// 数据库中也不存在，设置空值缓存（避免重复查询）
		go func() {
			if err := SetNullCache(id); err != nil {
				system.Logger.Error("Failed to set null cache async", "id", id, "error", err)
			}
		}()
		return nil, err
	}

	// 同步写入缓存（因为我们持有锁）
	if err := post.SetCache(); err != nil {
		system.Logger.Error("Failed to set post cache sync", "id", id, "error", err)
		// 缓存写入失败不影响返回数据
	}

	// 加载关联数据
	if err := LoadPostRelations(post); err != nil {
		system.Logger.Error("Failed to load post relations after rebuild", "id", id, "error", err)
	}

	system.Logger.Debug("Cache rebuilt successfully", "post_id", id)
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
		system.Logger.Error("Failed to cache list data", "key", key, "error", err)
		return err
	}

	system.Logger.Debug("List cache set successfully", "key", key, "expiration", randomExpiration)
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

	system.Logger.Debug("List cache hit", "key", key)
	return nil
}

// ============== 空值缓存相关函数 ==============

// IsNullCached 检查是否已缓存为空值
func IsNullCached(postID uint) bool {
	if system.Redis == nil || !system.Redis.IsAvailable() {
		return false
	}

	key := system.GenerateKey(NullCachePrefix, postID)
	return system.Redis.Exists(key)
}

// SetNullCache 设置空值缓存
func SetNullCache(postID uint) error {
	if system.Redis == nil || !system.Redis.IsAvailable() {
		return nil
	}

	key := system.GenerateKey(NullCachePrefix, postID)
	// 缓存一个简单的标记，表示该ID不存在
	err := system.Redis.Set(key, "null", NullCacheExpiration)
	if err != nil {
		system.Logger.Error("Failed to set null cache", "post_id", postID, "error", err)
		return err
	}

	system.Logger.Debug("Null cache set", "post_id", postID, "expiration", NullCacheExpiration)
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
		system.Logger.Error("Failed to delete null cache", "post_id", postID, "error", err)
		return err
	}

	system.Logger.Debug("Null cache deleted", "post_id", postID)
	return nil
}

// ============== 分布式锁相关函数（防击穿） ==============

// TryLock 尝试获取分布式锁
func TryLock(postID uint) (bool, error) {
	if system.Redis == nil || !system.Redis.IsAvailable() {
		return false, fmt.Errorf("redis not available")
	}

	key := system.GenerateKey(LockCachePrefix, postID)
	// 使用 SETNX 设置锁，如果key已存在则失败
	// 同时设置过期时间防止死锁
	return system.Redis.SetNX(key, "locked", LockCacheTimeout)
}

// ReleaseLock 释放分布式锁
func ReleaseLock(postID uint) error {
	if system.Redis == nil || !system.Redis.IsAvailable() {
		return nil
	}

	key := system.GenerateKey(LockCachePrefix, postID)
	err := system.Redis.Del(key)
	if err != nil {
		system.Logger.Error("Failed to release lock", "post_id", postID, "error", err)
		return err
	}

	system.Logger.Debug("Lock released", "post_id", postID)
	return nil
}

// WaitForLockRelease 等待锁释放（其他线程重建缓存完成）
func WaitForLockRelease(postID uint) (*Post, error) {
	for i := 0; i < LockMaxRetries; i++ {
		time.Sleep(LockRetryInterval)

		// 尝试从缓存获取数据（其他线程可能已经重建了缓存）
		if post, err := GetPostFromCache(postID); err == nil {
			system.Logger.Debug("Cache rebuilt by other thread", "post_id", postID)
			return post, nil
		}
	}

	// 等待超时，返回错误
	return nil, fmt.Errorf("wait for cache rebuild timeout")
}

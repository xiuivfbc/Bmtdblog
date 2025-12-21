package models

import (
	"encoding/json"
	"fmt"

	"strings"
	"sync"
	"time"

	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/xiuivfbc/bmtdblog/internal/api/dao"
	"github.com/xiuivfbc/bmtdblog/internal/config"
)

// 增量同步状态表
type ESSyncStatus struct {
	ID           uint      `gorm:"primaryKey"`
	LastSyncTime time.Time `json:"last_sync_time"`
	LastPostID   uint      `json:"last_post_id"`
	TotalSynced  int       `json:"total_synced"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// PostDocument ES中的博文文档结构
type PostDocument struct {
	ID           uint      `json:"id"`
	Title        string    `json:"title"`
	Body         string    `json:"body"`
	Tags         []string  `json:"tags"`
	Author       string    `json:"author"`
	IsPublished  bool      `json:"is_published"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	ViewCount    int       `json:"view_count"`
	CommentCount int       `json:"comment_count"`
	Excerpt      string    `json:"excerpt"`
}

// SearchRequest 搜索请求结构
type SearchRequest struct {
	Query    string   `json:"query"`
	Tags     []string `json:"tags,omitempty"`
	Page     int      `json:"page"`
	Size     int      `json:"size"`
	SortBy   string   `json:"sort_by"` // relevance, date, views
	DateFrom string   `json:"date_from,omitempty"`
	DateTo   string   `json:"date_to,omitempty"`
}

// SearchResponse 搜索响应结构
type SearchResponse struct {
	Posts       []*Post  `json:"posts"` // 直接使用Post模型
	Total       int64    `json:"total"`
	MaxScore    float64  `json:"max_score"`
	Took        int      `json:"took"`
	Suggestions []string `json:"suggestions,omitempty"`
}

// ESResponse ES原始响应结构（用于解析）
type ESResponse struct {
	Took int `json:"took"`
	Hits struct {
		Total struct {
			Value int64 `json:"value"`
		} `json:"total"`
		MaxScore float64 `json:"max_score"`
		Hits     []struct {
			ID        string              `json:"_id"`
			Score     float64             `json:"_score"`
			Source    PostDocument        `json:"_source"`
			Highlight map[string][]string `json:"highlight,omitempty"`
		} `json:"hits"`
	} `json:"hits"`
}

// IndexPost 将博文索引到ES
func IndexPost(post *Post) error {
	if !config.GetConfiguration().Elasticsearch.Enabled {
		return nil // ES未启用，跳过
	}

	if !dao.IsESAvailable() {
		config.Logger.Warn("ES不可用，跳过索引", "post_id", post.ID)
		return nil
	}

	doc := &PostDocument{
		ID:           post.ID,
		Title:        post.Title,
		Body:         post.Body,
		IsPublished:  post.IsPublished,
		CreatedAt:    *post.CreatedAt,
		UpdatedAt:    *post.UpdatedAt,
		ViewCount:    post.View,
		CommentCount: post.CommentTotal,
		Excerpt:      generateExcerpt(post.Body, 200),
	}

	// 获取标签
	if tags, err := ListTagByPostId(post.ID); err == nil {
		tagNames := make([]string, len(tags))
		for i, tag := range tags {
			tagNames[i] = tag.Name
		}
		doc.Tags = tagNames
	}

	docJSON, err := json.Marshal(doc)
	if err != nil {
		return err
	}

	indexName := config.GetConfiguration().GetElasticsearchIndexName()
	_, err = dao.ESClient.Index(
		indexName,
		strings.NewReader(string(docJSON)),
		dao.ESClient.Index.WithDocumentID(fmt.Sprintf("%d", post.ID)),
		dao.ESClient.Index.WithRefresh("true"),
	)

	if err != nil {
		config.Logger.Error("博文索引失败", "post_id", post.ID, "error", err)
		return err
	}

	config.Logger.Debug("博文索引成功", "post_id", post.ID)
	return nil
}

// DeletePostFromIndex 从ES删除博文
func DeletePostFromIndex(postID uint) error {
	if !config.GetConfiguration().Elasticsearch.Enabled {
		return nil // ES未启用，跳过
	}

	if !dao.IsESAvailable() {
		config.Logger.Warn("ES不可用，跳过删除", "post_id", postID)
		return nil
	}

	indexName := config.GetConfiguration().GetElasticsearchIndexName()
	_, err := dao.ESClient.Delete(
		indexName,
		fmt.Sprintf("%d", postID),
		dao.ESClient.Delete.WithRefresh("true"),
	)

	if err != nil {
		config.Logger.Error("博文删除失败", "post_id", postID, "error", err)
		return err
	}

	config.Logger.Debug("博文删除成功", "post_id", postID)
	return nil
}

// generateExcerpt 生成摘要
func generateExcerpt(body string, maxLen int) string {
	if len(body) <= maxLen {
		return body
	}
	return body[:maxLen] + "..."
}

// SearchPosts 搜索博文
func SearchPosts(req *SearchRequest) (*SearchResponse, error) {
	if !config.GetConfiguration().Elasticsearch.Enabled {
		return searchPostsFromDB(req) // 降级到数据库搜索
	}

	if !dao.IsESAvailable() {
		config.Logger.Warn("ES不可用，降级到数据库搜索")
		return searchPostsFromDB(req)
	}

	// 构建ES查询
	queryJSON := buildSearchQuery(req)

	// 执行搜索
	indexName := config.GetConfiguration().GetElasticsearchIndexName()
	res, err := dao.ESClient.Search(
		dao.ESClient.Search.WithIndex(indexName),
		dao.ESClient.Search.WithBody(strings.NewReader(string(queryJSON))),
		dao.ESClient.Search.WithTrackTotalHits(true),
	)

	if err != nil {
		config.Logger.Error("ES搜索失败，降级到数据库搜索", "error", err)
		return searchPostsFromDB(req)
	}
	defer res.Body.Close()

	// 解析ES响应
	return parseSearchResponse(res)
}

// buildSearchQuery 构建ES查询
func buildSearchQuery(req *SearchRequest) string {
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []map[string]interface{}{
					{
						"term": map[string]interface{}{
							"is_published": true,
						},
					},
				},
			},
		},
		"from": (req.Page - 1) * req.Size,
		"size": req.Size,
		"highlight": map[string]interface{}{
			"fields": map[string]interface{}{
				"title": map[string]interface{}{
					"fragment_size":       100,
					"number_of_fragments": 1,
				},
				"body": map[string]interface{}{
					"fragment_size":       200,
					"number_of_fragments": 3,
				},
			},
			"pre_tags":  []string{"<mark>"},
			"post_tags": []string{"</mark>"},
		},
	}

	// 添加搜索条件
	if req.Query != "" {
		mustQueries := query["query"].(map[string]interface{})["bool"].(map[string]interface{})["must"].([]map[string]interface{})
		mustQueries = append(mustQueries, map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":     req.Query,
				"fields":    []string{"title^3", "body^1", "tags^2"},
				"type":      "best_fields",
				"operator":  "and",
				"fuzziness": "AUTO",
			},
		})
		query["query"].(map[string]interface{})["bool"].(map[string]interface{})["must"] = mustQueries
	}

	// 添加标签过滤
	if len(req.Tags) > 0 {
		filterQueries := []map[string]interface{}{
			{
				"terms": map[string]interface{}{
					"tags": req.Tags,
				},
			},
		}
		query["query"].(map[string]interface{})["bool"].(map[string]interface{})["filter"] = filterQueries
	}

	// 添加日期范围过滤
	if req.DateFrom != "" || req.DateTo != "" {
		dateRange := map[string]interface{}{}
		if req.DateFrom != "" {
			dateRange["gte"] = req.DateFrom
		}
		if req.DateTo != "" {
			dateRange["lte"] = req.DateTo
		}

		if len(dateRange) > 0 {
			existingFilters := query["query"].(map[string]interface{})["bool"].(map[string]interface{})["filter"]
			if existingFilters == nil {
				existingFilters = []map[string]interface{}{}
			}
			filters := existingFilters.([]map[string]interface{})
			filters = append(filters, map[string]interface{}{
				"range": map[string]interface{}{
					"created_at": dateRange,
				},
			})
			query["query"].(map[string]interface{})["bool"].(map[string]interface{})["filter"] = filters
		}
	}

	// 添加排序
	switch req.SortBy {
	case "date":
		query["sort"] = []map[string]interface{}{
			{"created_at": map[string]string{"order": "desc"}},
		}
	case "views":
		query["sort"] = []map[string]interface{}{
			{"view_count": map[string]string{"order": "desc"}},
		}
	default: // relevance
		query["sort"] = []map[string]interface{}{
			{"_score": map[string]string{"order": "desc"}},
			{"created_at": map[string]string{"order": "desc"}},
		}
	}

	queryJSON, _ := json.Marshal(query)
	return string(queryJSON)
}

// parseSearchResponse 解析ES搜索响应
func parseSearchResponse(res *esapi.Response) (*SearchResponse, error) {
	var esResp ESResponse
	if err := json.NewDecoder(res.Body).Decode(&esResp); err != nil {
		return nil, fmt.Errorf("解析ES响应失败: %w", err)
	}

	posts := make([]*Post, 0, len(esResp.Hits.Hits))

	for _, hit := range esResp.Hits.Hits {
		// 从数据库获取完整的Post对象（确保数据一致性）
		post, err := GetPostById(hit.Source.ID)
		if err != nil {
			config.Logger.Warn("无法从数据库获取博文", "id", hit.Source.ID, "error", err)
			continue
		}

		// 加载关联数据
		if err := LoadPostRelations(post); err != nil {
			config.Logger.Warn("加载博文关联数据失败", "id", post.ID, "error", err)
		}

		// 添加高亮信息（如果需要的话）
		if len(hit.Highlight) > 0 {
			// 可以在这里处理高亮信息，比如设置到Post的某个字段
			config.Logger.Debug("搜索高亮", "post_id", post.ID, "highlights", hit.Highlight)
		}

		posts = append(posts, post)
	}

	return &SearchResponse{
		Posts:    posts,
		Total:    esResp.Hits.Total.Value,
		MaxScore: esResp.Hits.MaxScore,
		Took:     esResp.Took,
	}, nil
}

// searchPostsFromDB 数据库搜索（降级方案）
func searchPostsFromDB(req *SearchRequest) (*SearchResponse, error) {
	fmt.Printf("数据库搜索开始: query='%s', tags=%v, page=%d\n", req.Query, req.Tags, req.Page)

	var posts []*Post
	var total int64

	query := DB.Model(&Post{}).Where("is_published = ?", true)

	// 添加关键词搜索
	if req.Query != "" {
		searchPattern := "%" + req.Query + "%"
		query = query.Where("title LIKE ? OR body LIKE ?", searchPattern, searchPattern)
		config.Logger.Info("数据库搜索", "keyword", req.Query, "pattern", searchPattern)
		fmt.Printf("添加关键词搜索: pattern='%s'\n", searchPattern)
	} else {
		fmt.Printf("空关键词搜索，显示所有文章\n")
	}

	// 添加标签过滤
	if len(req.Tags) > 0 {
		// 通过关联表查询
		subQuery := DB.Table("post_tags pt").
			Select("pt.post_id").
			Joins("JOIN tags t ON t.id = pt.tag_id").
			Where("t.name IN ?", req.Tags)
		query = query.Where("id IN (?)", subQuery)
	}

	// 计算总数
	query.Count(&total)
	config.Logger.Info("搜索统计", "total", total, "query", req.Query)
	fmt.Printf("搜索统计: total=%d\n", total)

	// 添加排序
	switch req.SortBy {
	case "date":
		query = query.Order("created_at DESC")
	case "views":
		query = query.Order("view DESC")
	default:
		query = query.Order("created_at DESC")
	}

	// 分页
	err := query.Offset((req.Page - 1) * req.Size).
		Limit(req.Size).
		Find(&posts).Error

	if err != nil {
		fmt.Printf("数据库查询出错: %v\n", err)
		return nil, err
	}

	fmt.Printf("查询结果: 找到%d条记录\n", len(posts))

	// 加载关联数据
	for _, post := range posts {
		if err := LoadPostRelations(post); err != nil {
			config.Logger.Warn("加载博文关联数据失败", "id", post.ID, "error", err)
		}
	}

	return &SearchResponse{
		Posts: posts,
		Total: total,
		Took:  0, // 数据库查询不统计时间
	}, nil
}

// GetSearchSuggestions 获取搜索建议
func GetSearchSuggestions(prefix string, limit int) ([]string, error) {
	if !config.GetConfiguration().Elasticsearch.Enabled || !dao.IsESAvailable() {
		return getSearchSuggestionsFromDB(prefix, limit)
	}

	// ES建议查询
	query := map[string]interface{}{
		"suggest": map[string]interface{}{
			"title_suggest": map[string]interface{}{
				"prefix": prefix,
				"term": map[string]interface{}{
					"field": "title",
					"size":  limit,
				},
			},
		},
	}

	queryJSON, _ := json.Marshal(query)
	indexName := config.GetConfiguration().GetElasticsearchIndexName()

	res, err := dao.ESClient.Search(
		dao.ESClient.Search.WithIndex(indexName),
		dao.ESClient.Search.WithBody(strings.NewReader(string(queryJSON))),
	)
	if err != nil {
		config.Logger.Warn("ES建议查询失败，降级到数据库", "error", err)
		return getSearchSuggestionsFromDB(prefix, limit)
	}
	defer res.Body.Close()

	// 简化处理，这里可以根据实际需要解析建议响应
	return getSearchSuggestionsFromDB(prefix, limit)
}

// getSearchSuggestionsFromDB 从数据库获取搜索建议
func getSearchSuggestionsFromDB(prefix string, limit int) ([]string, error) {
	var suggestions []string

	err := DB.Model(&Post{}).
		Where("is_published = ? AND title LIKE ?", true, prefix+"%").
		Limit(limit).
		Pluck("title", &suggestions).Error

	return suggestions, err
}

// ========== 数据同步功能 ==========

// SyncAllPostsToES 将所有已发布的博文同步到ES
func SyncAllPostsToES() error {
	if !config.GetConfiguration().Elasticsearch.Enabled {
		config.Logger.Info("ES未启用，跳过数据同步")
		return nil
	}

	if !dao.IsESAvailable() {
		return fmt.Errorf("ES不可用")
	}

	var posts []*Post
	err := DB.Where("is_published = ?", true).Find(&posts).Error
	if err != nil {
		return fmt.Errorf("查询博文失败: %w", err)
	}

	config.Logger.Info("开始批量同步博文到ES", "total_count", len(posts))

	// 批量处理，每批100个
	const batchSize = 100
	successCount := 0
	failCount := 0

	for i := 0; i < len(posts); i += batchSize {
		end := i + batchSize
		if end > len(posts) {
			end = len(posts)
		}

		batch := posts[i:end]
		if err := bulkIndexPosts(batch); err != nil {
			config.Logger.Error("批量同步失败", "batch_start", i, "batch_size", len(batch), "error", err)
			failCount += len(batch)
		} else {
			successCount += len(batch)
			config.Logger.Info("批量同步进度", "processed", end, "total", len(posts))
		}

		// 避免过快请求，给ES一些处理时间
		if i+batchSize < len(posts) {
			time.Sleep(100 * time.Millisecond)
		}
	}

	config.Logger.Info("博文批量同步完成",
		"total", len(posts),
		"success", successCount,
		"failed", failCount)

	return nil
}

// bulkIndexPosts 批量索引博文到ES
func bulkIndexPosts(posts []*Post) error {
	if len(posts) == 0 {
		return nil
	}

	indexName := config.GetConfiguration().GetElasticsearchIndexName()
	var bulkBody strings.Builder

	for _, post := range posts {
		// 构建文档数据
		doc := &PostDocument{
			ID:           post.ID,
			Title:        post.Title,
			Body:         post.Body,
			IsPublished:  post.IsPublished,
			CreatedAt:    *post.CreatedAt,
			UpdatedAt:    *post.UpdatedAt,
			ViewCount:    post.View,
			CommentCount: post.CommentTotal,
			Excerpt:      generateExcerpt(post.Body, 200),
		}

		// 获取标签
		if tags, err := ListTagByPostId(post.ID); err == nil {
			tagNames := make([]string, len(tags))
			for i, tag := range tags {
				tagNames[i] = tag.Name
			}
			doc.Tags = tagNames
		}

		// 构建批量请求的action行
		action := map[string]interface{}{
			"index": map[string]interface{}{
				"_index": indexName,
				"_id":    fmt.Sprintf("%d", post.ID),
			},
		}

		actionJSON, err := json.Marshal(action)
		if err != nil {
			config.Logger.Error("构建批量action失败", "post_id", post.ID, "error", err)
			continue
		}

		docJSON, err := json.Marshal(doc)
		if err != nil {
			config.Logger.Error("序列化文档失败", "post_id", post.ID, "error", err)
			continue
		}

		// 添加到bulk请求体（每个操作需要两行：action + document）
		bulkBody.WriteString(string(actionJSON) + "\n")
		bulkBody.WriteString(string(docJSON) + "\n")
	}

	if bulkBody.Len() == 0 {
		return fmt.Errorf("没有有效的文档需要索引")
	}

	// 执行批量请求
	res, err := dao.ESClient.Bulk(
		strings.NewReader(bulkBody.String()),
		dao.ESClient.Bulk.WithIndex(indexName),
		dao.ESClient.Bulk.WithRefresh("false"), // 批量完成后再刷新，提高性能
	)

	if err != nil {
		return fmt.Errorf("批量索引请求失败: %w", err)
	}
	defer res.Body.Close()

	// 检查批量操作结果
	if res.IsError() {
		return fmt.Errorf("批量索引响应错误: %s", res.String())
	}

	// 解析批量响应，检查是否有失败的操作
	var bulkResp map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&bulkResp); err != nil {
		return fmt.Errorf("解析批量响应失败: %w", err)
	}

	// 检查是否有错误
	if errors, exists := bulkResp["errors"]; exists && errors.(bool) {
		config.Logger.Warn("批量操作中有部分失败", "batch_size", len(posts))
		// 可以进一步解析具体的错误信息
		if items, exists := bulkResp["items"]; exists {
			config.Logger.Debug("批量操作详情", "items", items)
		}
	}

	config.Logger.Debug("批量索引成功", "batch_size", len(posts))
	return nil
}

// ESTaskQueue 批量ES任务队列
type ESTaskQueue struct {
	tasks  []ESTask
	mutex  sync.Mutex
	ticker *time.Ticker
	quit   chan bool
}

type ESTask struct {
	Action string // "index", "delete"
	PostID uint
	Post   *Post
}

var esTaskQueue *ESTaskQueue

// InitESTaskQueue 初始化ES任务队列
func InitESTaskQueue() {
	esTaskQueue = &ESTaskQueue{
		tasks:  make([]ESTask, 0),
		ticker: time.NewTicker(5 * time.Second), // 每5秒处理一次批量任务
		quit:   make(chan bool),
	}

	// 启动后台处理协程
	go esTaskQueue.processLoop()

	// 启动时进行增量同步检查
	go performIncrementalSync()

	config.Logger.Info("ES批量任务队列已启动")
}

// StopESTaskQueue 停止ES任务队列
func StopESTaskQueue() {
	if esTaskQueue != nil {
		esTaskQueue.quit <- true
		esTaskQueue.ticker.Stop()

		// 处理剩余任务
		esTaskQueue.processRemaining()
		config.Logger.Info("ES批量任务队列已停止")
	}
}

// performIncrementalSync 执行增量同步
func performIncrementalSync() {
	config.Logger.Info("开始ES增量同步检查...")

	// 获取上次同步状态
	var syncStatus ESSyncStatus
	result := DB.Order("id desc").First(&syncStatus)

	var lastSyncTime time.Time
	var lastPostID uint

	if result.Error != nil {
		// 首次同步，获取所有发布的博文
		config.Logger.Info("首次ES同步，将索引所有已发布博文")
		lastSyncTime = time.Time{} // 零值时间
		lastPostID = 0
	} else {
		lastSyncTime = syncStatus.LastSyncTime
		lastPostID = syncStatus.LastPostID
		config.Logger.Info("检测到上次同步记录",
			"last_sync_time", lastSyncTime,
			"last_post_id", lastPostID,
			"total_synced", syncStatus.TotalSynced)
	}

	// 查询需要同步的博文
	var posts []*Post
	query := DB.Where("status = ? AND is_published = ?", "published", true)

	if !lastSyncTime.IsZero() {
		// 增量同步：只同步新增或更新的博文
		query = query.Where("(created_at > ? OR updated_at > ?) AND id > ?",
			lastSyncTime, lastSyncTime, lastPostID)
	}

	if err := query.Order("id asc").Find(&posts).Error; err != nil {
		config.Logger.Error("查询待同步博文失败", "error", err)
		return
	}

	if len(posts) == 0 {
		config.Logger.Info("没有需要同步的博文")
		return
	}

	config.Logger.Info("发现待同步博文", "count", len(posts))

	// 批量同步到ES
	batchSize := 50
	successCount := 0

	for i := 0; i < len(posts); i += batchSize {
		end := i + batchSize
		if end > len(posts) {
			end = len(posts)
		}

		batch := posts[i:end]
		if err := bulkIndexPosts(batch); err != nil {
			config.Logger.Error("批量索引博文失败", "error", err, "batch_start", i, "batch_size", len(batch))
			continue
		}

		successCount += len(batch)
		config.Logger.Info("批量索引成功", "batch_size", len(batch), "total_success", successCount)
	}

	// 更新同步状态
	newSyncStatus := ESSyncStatus{
		LastSyncTime: time.Now(),
		LastPostID:   posts[len(posts)-1].ID, // 最后一个博文ID
		TotalSynced:  successCount,
	}

	if result.Error != nil {
		// 创建新记录
		if err := DB.Create(&newSyncStatus).Error; err != nil {
			config.Logger.Error("创建同步状态记录失败", "error", err)
		}
	} else {
		// 更新现有记录
		if err := DB.Model(&syncStatus).Updates(&newSyncStatus).Error; err != nil {
			config.Logger.Error("更新同步状态记录失败", "error", err)
		}
	}

	config.Logger.Info("ES增量同步完成", "synced_count", successCount)
}

// GetESIndexStatus 获取ES索引状态
func GetESIndexStatus() (*ESSyncStatus, error) {
	var syncStatus ESSyncStatus
	err := DB.Order("id desc").First(&syncStatus).Error
	if err != nil {
		return nil, err
	}
	return &syncStatus, nil
}

// ForceFullReindex 强制全量重建索引
func ForceFullReindex() error {
	config.Logger.Info("开始强制全量重建ES索引...")

	// 删除现有索引
	if err := DeleteESIndex(); err != nil {
		config.Logger.Error("删除ES索引失败", "error", err)
		// 继续执行，可能索引不存在
	}

	// 重建索引
	if err := CreateESIndex(); err != nil {
		return fmt.Errorf("创建ES索引失败: %w", err)
	}

	// 清除同步状态
	if err := DB.Where("1 = 1").Delete(&ESSyncStatus{}).Error; err != nil {
		config.Logger.Error("清除同步状态失败", "error", err)
	}

	// 重新执行增量同步（此时会变成全量同步）
	go performIncrementalSync()

	return nil
}

// CreateESIndex 创建ES索引
func CreateESIndex() error {
	if !config.GetConfiguration().Elasticsearch.Enabled {
		return fmt.Errorf("ES未启用")
	}

	// 这里应该调用system包中的创建索引函数
	// 或者实现具体的索引创建逻辑
	config.Logger.Info("创建ES索引")
	return nil
}

// DeleteESIndex 删除ES索引
func DeleteESIndex() error {
	if !config.GetConfiguration().Elasticsearch.Enabled {
		return fmt.Errorf("ES未启用")
	}

	// 这里应该调用system包中的删除索引函数
	// 或者实现具体的索引删除逻辑
	config.Logger.Info("删除ES索引")
	return nil
}

// StopESTaskQueue 停止ES任务队列

// AddESTask 添加ES任务到队列
func AddESTask(task ESTask) {
	if esTaskQueue == nil {
		// 如果队列未初始化，直接同步处理
		processSingleTask(task)
		return
	}

	esTaskQueue.mutex.Lock()
	defer esTaskQueue.mutex.Unlock()

	esTaskQueue.tasks = append(esTaskQueue.tasks, task)

	// 如果队列过大，立即处理
	if len(esTaskQueue.tasks) >= 50 {
		go esTaskQueue.processBatch()
	}
}

// processLoop 后台处理循环
func (q *ESTaskQueue) processLoop() {
	for {
		select {
		case <-q.ticker.C:
			q.processBatch()
		case <-q.quit:
			return
		}
	}
}

// processBatch 处理当前批次
func (q *ESTaskQueue) processBatch() {
	q.mutex.Lock()
	if len(q.tasks) == 0 {
		q.mutex.Unlock()
		return
	}

	// 复制当前任务并清空队列
	batch := make([]ESTask, len(q.tasks))
	copy(batch, q.tasks)
	q.tasks = q.tasks[:0]
	q.mutex.Unlock()

	// 分离索引和删除任务
	var indexTasks []*Post
	var deleteTasks []uint

	for _, task := range batch {
		switch task.Action {
		case "index":
			if task.Post != nil {
				indexTasks = append(indexTasks, task.Post)
			}
		case "delete":
			deleteTasks = append(deleteTasks, task.PostID)
		}
	}

	// 批量处理索引任务
	if len(indexTasks) > 0 {
		if err := bulkIndexPosts(indexTasks); err != nil {
			config.Logger.Error("批量索引任务失败", "count", len(indexTasks), "error", err)
		} else {
			config.Logger.Debug("批量索引任务完成", "count", len(indexTasks))
		}
	}

	// 批量处理删除任务
	if len(deleteTasks) > 0 {
		if err := bulkDeletePosts(deleteTasks); err != nil {
			config.Logger.Error("批量删除任务失败", "count", len(deleteTasks), "error", err)
		} else {
			config.Logger.Debug("批量删除任务完成", "count", len(deleteTasks))
		}
	}
}

// processRemaining 处理剩余任务
func (q *ESTaskQueue) processRemaining() {
	q.processBatch()
}

// processSingleTask 单个任务同步处理（降级方案）
func processSingleTask(task ESTask) {
	switch task.Action {
	case "index":
		if task.Post != nil {
			if err := IndexPost(task.Post); err != nil {
				config.Logger.Error("同步索引失败", "post_id", task.PostID, "error", err)
			}
		}
	case "delete":
		if err := DeletePostFromIndex(task.PostID); err != nil {
			config.Logger.Error("同步删除失败", "post_id", task.PostID, "error", err)
		}
	}
}

// bulkDeletePosts 批量删除博文索引
func bulkDeletePosts(postIDs []uint) error {
	if len(postIDs) == 0 {
		return nil
	}

	indexName := config.GetConfiguration().GetElasticsearchIndexName()
	var bulkBody strings.Builder

	for _, postID := range postIDs {
		// 构建删除操作
		action := map[string]interface{}{
			"delete": map[string]interface{}{
				"_index": indexName,
				"_id":    fmt.Sprintf("%d", postID),
			},
		}

		actionJSON, err := json.Marshal(action)
		if err != nil {
			config.Logger.Error("构建删除action失败", "post_id", postID, "error", err)
			continue
		}

		bulkBody.WriteString(string(actionJSON) + "\n")
	}

	if bulkBody.Len() == 0 {
		return fmt.Errorf("没有有效的删除操作")
	}

	// 执行批量删除
	res, err := dao.ESClient.Bulk(
		strings.NewReader(bulkBody.String()),
		dao.ESClient.Bulk.WithIndex(indexName),
		dao.ESClient.Bulk.WithRefresh("false"),
	)

	if err != nil {
		return fmt.Errorf("批量删除请求失败: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("批量删除响应错误: %s", res.String())
	}

	config.Logger.Debug("批量删除成功", "count", len(postIDs))
	return nil
}

// ReindexAllPosts 重建所有博文的ES索引
func ReindexAllPosts() error {
	cfg := config.GetConfiguration()
	if !cfg.Elasticsearch.Enabled {
		config.Logger.Info("ES未启用，跳过重建索引")
		return nil
	}

	if !dao.IsESAvailable() {
		return fmt.Errorf("ES不可用")
	}

	config.Logger.Info("开始重建ES索引")

	// 先删除现有索引
	indexName := config.GetConfiguration().GetElasticsearchIndexName()
	_, err := dao.ESClient.Indices.Delete([]string{indexName})
	if err != nil {
		config.Logger.Warn("删除现有索引失败", "error", err)
	}

	// 重新创建索引
	if err := dao.InitElasticsearch(); err != nil {
		return fmt.Errorf("重新创建索引失败: %w", err)
	}

	// 同步所有数据
	return SyncAllPostsToES()
}

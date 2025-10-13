package models

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/xiuivfbc/bmtdblog/system"
)

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
	if !system.GetConfiguration().Elasticsearch.Enabled {
		return nil // ES未启用，跳过
	}

	if !system.IsESAvailable() {
		slog.Warn("ES不可用，跳过索引", "post_id", post.ID)
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

	indexName := system.GetConfiguration().GetElasticsearchIndexName()
	_, err = system.ESClient.Index(
		indexName,
		strings.NewReader(string(docJSON)),
		system.ESClient.Index.WithDocumentID(fmt.Sprintf("%d", post.ID)),
		system.ESClient.Index.WithRefresh("true"),
	)

	if err != nil {
		slog.Error("博文索引失败", "post_id", post.ID, "error", err)
		return err
	}

	slog.Debug("博文索引成功", "post_id", post.ID)
	return nil
}

// DeletePostFromIndex 从ES删除博文
func DeletePostFromIndex(postID uint) error {
	if !system.GetConfiguration().Elasticsearch.Enabled {
		return nil // ES未启用，跳过
	}

	if !system.IsESAvailable() {
		slog.Warn("ES不可用，跳过删除", "post_id", postID)
		return nil
	}

	indexName := system.GetConfiguration().GetElasticsearchIndexName()
	_, err := system.ESClient.Delete(
		indexName,
		fmt.Sprintf("%d", postID),
		system.ESClient.Delete.WithRefresh("true"),
	)

	if err != nil {
		slog.Error("博文删除失败", "post_id", postID, "error", err)
		return err
	}

	slog.Debug("博文删除成功", "post_id", postID)
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
	if !system.GetConfiguration().Elasticsearch.Enabled {
		return searchPostsFromDB(req) // 降级到数据库搜索
	}

	if !system.IsESAvailable() {
		slog.Warn("ES不可用，降级到数据库搜索")
		return searchPostsFromDB(req)
	}

	// 构建ES查询
	queryJSON := buildSearchQuery(req)

	// 执行搜索
	indexName := system.GetConfiguration().GetElasticsearchIndexName()
	res, err := system.ESClient.Search(
		system.ESClient.Search.WithIndex(indexName),
		system.ESClient.Search.WithBody(strings.NewReader(string(queryJSON))),
		system.ESClient.Search.WithTrackTotalHits(true),
	)

	if err != nil {
		slog.Error("ES搜索失败，降级到数据库搜索", "error", err)
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
			slog.Warn("无法从数据库获取博文", "id", hit.Source.ID, "error", err)
			continue
		}

		// 加载关联数据
		if err := LoadPostRelations(post); err != nil {
			slog.Warn("加载博文关联数据失败", "id", post.ID, "error", err)
		}

		// 添加高亮信息（如果需要的话）
		if len(hit.Highlight) > 0 {
			// 可以在这里处理高亮信息，比如设置到Post的某个字段
			slog.Debug("搜索高亮", "post_id", post.ID, "highlights", hit.Highlight)
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
		slog.Info("数据库搜索", "keyword", req.Query, "pattern", searchPattern)
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
	slog.Info("搜索统计", "total", total, "query", req.Query)
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
			slog.Warn("加载博文关联数据失败", "id", post.ID, "error", err)
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
	if !system.GetConfiguration().Elasticsearch.Enabled || !system.IsESAvailable() {
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
	indexName := system.GetConfiguration().GetElasticsearchIndexName()

	res, err := system.ESClient.Search(
		system.ESClient.Search.WithIndex(indexName),
		system.ESClient.Search.WithBody(strings.NewReader(string(queryJSON))),
	)
	if err != nil {
		slog.Warn("ES建议查询失败，降级到数据库", "error", err)
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
	if !system.GetConfiguration().Elasticsearch.Enabled {
		slog.Info("ES未启用，跳过数据同步")
		return nil
	}

	if !system.IsESAvailable() {
		return fmt.Errorf("ES不可用")
	}

	var posts []*Post
	err := DB.Where("is_published = ?", true).Find(&posts).Error
	if err != nil {
		return fmt.Errorf("查询博文失败: %w", err)
	}

	slog.Info("开始同步博文到ES", "count", len(posts))

	successCount := 0
	failCount := 0

	for _, post := range posts {
		if err := IndexPost(post); err != nil {
			slog.Error("同步博文到ES失败", "id", post.ID, "title", post.Title, "error", err)
			failCount++
		} else {
			successCount++
			slog.Debug("博文同步成功", "id", post.ID, "title", post.Title)
		}
	}

	slog.Info("博文同步完成",
		"total", len(posts),
		"success", successCount,
		"failed", failCount)

	return nil
}

// ReindexAllPosts 重建所有博文的ES索引
func ReindexAllPosts() error {
	if !system.GetConfiguration().Elasticsearch.Enabled {
		slog.Info("ES未启用，跳过重建索引")
		return nil
	}

	if !system.IsESAvailable() {
		return fmt.Errorf("ES不可用")
	}

	slog.Info("开始重建ES索引")

	// 先删除现有索引
	indexName := system.GetConfiguration().GetElasticsearchIndexName()
	_, err := system.ESClient.Indices.Delete([]string{indexName})
	if err != nil {
		slog.Warn("删除现有索引失败", "error", err)
	}

	// 重新创建索引
	if err := system.InitElasticsearch(); err != nil {
		return fmt.Errorf("重新创建索引失败: %w", err)
	}

	// 同步所有数据
	return SyncAllPostsToES()
}

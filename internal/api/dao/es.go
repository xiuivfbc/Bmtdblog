package dao

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/xiuivfbc/bmtdblog/internal/config"
)

// 全局ES客户端实例
var ESClient *elasticsearch.Client

// InitElasticsearch 初始化ElasticSearch客户端
func InitElasticsearch() error {
	myCfg := config.GetConfiguration()
	cfg := elasticsearch.Config{
		Addresses: []string{
			myCfg.Elasticsearch.URL,
		},
		Username: myCfg.Elasticsearch.Username, // 可选认证
		Password: myCfg.Elasticsearch.Password, // 可选认证
	}

	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return fmt.Errorf("创建ES客户端失败: %w", err)
	}

	ESClient = client

	// 测试连接
	res, err := ESClient.Info()
	if err != nil {
		return fmt.Errorf("ES连接测试失败: %w", err)
	}
	defer res.Body.Close()

	config.Logger.Info("ElasticSearch连接成功")

	return nil
}

// createPostIndex 创建博文索引
func CreatePostIndex(indexName string) error {
	if ESClient == nil {
		return fmt.Errorf("ES客户端未初始化")
	}

	// 检查索引是否存在
	res, err := ESClient.Indices.Exists([]string{indexName})
	if err != nil {
		return err
	}
	defer res.Body.Close()

	// 如果索引已存在，跳过创建
	if res.StatusCode == 200 {
		slog.Info("索引已存在", "index", indexName)
		return nil
	}

	// 索引映射配置
	mapping := `{
		"settings": {
			"number_of_shards": 1,
			"number_of_replicas": 0,
			"analysis": {
				"analyzer": {
					"ik_max_word": {
						"type": "standard"
					},
					"ik_smart": {
						"type": "standard"
					}
				}
			}
		},
		"mappings": {
			"properties": {
				"id": {"type": "long"},
				"title": {
					"type": "text",
					"analyzer": "ik_max_word",
					"search_analyzer": "ik_smart",
					"fields": {
						"keyword": {"type": "keyword"}
					}
				},
				"body": {
					"type": "text", 
					"analyzer": "ik_max_word",
					"search_analyzer": "ik_smart"
				},
				"tags": {
					"type": "keyword"
				},
				"author": {"type": "keyword"},
				"is_published": {"type": "boolean"},
				"created_at": {"type": "date"},
				"updated_at": {"type": "date"},
				"view_count": {"type": "long"},
				"comment_count": {"type": "long"},
				"excerpt": {
					"type": "text",
					"analyzer": "ik_max_word"
				}
			}
		}
	}`

	// 创建索引
	res, err = ESClient.Indices.Create(
		indexName,
		ESClient.Indices.Create.WithBody(strings.NewReader(mapping)),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("创建索引失败: %s", res.String())
	}

	slog.Info("索引创建成功", "index", indexName)
	return nil
}

// IsESAvailable 检查ES是否可用
func IsESAvailable() bool {
	if ESClient == nil {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := ESClient.Ping(ESClient.Ping.WithContext(ctx))
	return err == nil
}

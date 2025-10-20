package system

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"golang.org/x/crypto/acme/autocert"
)

type (
	Backup struct {
		Enabled   bool   `mapstructure:"enabled"`
		BackupKey string `mapstructure:"backup_key"`
	}

	Database struct {
		Dialect string `mapstructure:"dialect"`
		DSN     string `mapstructure:"dsn"`
	}

	Author struct {
		Name  string `mapstructure:"name"`
		Email string `mapstructure:"email"`
	}

	Seo struct {
		Description string `mapstructure:"description"`
		Author      Author `mapstructure:"author"`
	}

	Qiniu struct {
		Enabled    bool   `mapstructure:"enabled"`
		AccessKey  string `mapstructure:"accesskey"`
		SecretKey  string `mapstructure:"secretkey"`
		FileServer string `mapstructure:"fileserver"`
		Bucket     string `mapstructure:"bucket"`
	}

	Smms struct {
		Enabled bool   `mapstructure:"enabled"`
		ApiUrl  string `mapstructure:"apiurl"`
		ApiKey  string `mapstructure:"apikey"`
	}

	RedisConfig struct {
		Enabled  bool   `mapstructure:"enabled"`
		Addr     string `mapstructure:"addr"`
		Password string `mapstructure:"password"`
		DB       int    `mapstructure:"db"`
		PoolSize int    `mapstructure:"pool_size"`
	}

	ElasticsearchConfig struct {
		Enabled   bool   `mapstructure:"enabled"`
		URL       string `mapstructure:"url"`
		Username  string `mapstructure:"username"`
		Password  string `mapstructure:"password"`
		IndexName string `mapstructure:"index_name"`
	}

	Github struct {
		Enabled      bool   `mapstructure:"enabled"`
		ClientId     string `mapstructure:"clientid"`
		ClientSecret string `mapstructure:"clientsecret"`
		RedirectURL  string `mapstructure:"redirecturl"`
		AuthUrl      string `mapstructure:"authurl"`
		TokenUrl     string `mapstructure:"tokenurl"`
		Scope        string `mapstructure:"scope"`
	}

	Smtp struct {
		Enabled  bool   `mapstructure:"enabled"`
		Username string `mapstructure:"username"`
		Password string `mapstructure:"password"`
		Host     string `mapstructure:"host"`
	}

	Navigator struct {
		Title  string `mapstructure:"title"`
		Url    string `mapstructure:"url"`
		Target string `mapstructure:"target"`
	}

	TLSConfig struct {
		Enabled  bool   `mapstructure:"enabled"`
		AutoCert bool   `mapstructure:"auto_cert"`
		Domain   string `mapstructure:"domain"`
		Email    string `mapstructure:"email"`
		CertFile string `mapstructure:"cert_file"`
		KeyFile  string `mapstructure:"key_file"`
		CertDir  string `mapstructure:"cert_dir"`
	}

	Configuration struct {
		Addr          string              `mapstructure:"addr"`
		SignupEnabled bool                `mapstructure:"signup_enabled"`
		Title         string              `mapstructure:"title"`
		SessionSecret string              `mapstructure:"session_secret"`
		Domain        string              `mapstructure:"domain"`
		FileServer    string              `mapstructure:"file_server"`
		NotifyEmails  string              `mapstructure:"notify_emails"`
		PageSize      int                 `mapstructure:"page_size"`
		PublicDir     string              `mapstructure:"public"`
		ViewDir       string              `mapstructure:"view"`
		Dir           string              `mapstructure:"dir"`
		Database      Database            `mapstructure:"database"`
		Seo           Seo                 `mapstructure:"seo"`
		Qiniu         Qiniu               `mapstructure:"qiniu"`
		Smms          Smms                `mapstructure:"smms"`
		Redis         RedisConfig         `mapstructure:"redis"`
		Elasticsearch ElasticsearchConfig `mapstructure:"elasticsearch"`
		Github        Github              `mapstructure:"github"`
		Smtp          Smtp                `mapstructure:"smtp"`
		TLS           TLSConfig           `mapstructure:"tls"`
		Navigators    []Navigator         `mapstructure:"navigators"`
		Backup        Backup              `mapstructure:"backup"`
	}
)

// Redis客户端结构体
type RedisCacheClient struct {
	client *redis.Client
	ctx    context.Context
}

var (
	Logger        *slog.Logger
	Redis         *RedisCacheClient
	ESClient      *elasticsearch.Client
	configuration *Configuration
)

func (a Author) String() string {
	return fmt.Sprintf("%s,%s", a.Name, a.Email)
}

func LoadConfiguration(path string) error {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("toml")

	if err := v.ReadInConfig(); err != nil {
		return err
	}

	var config Configuration
	if err := v.Unmarshal(&config); err != nil {
		return err
	}

	configuration = &config
	return nil
}

func GetConfiguration() *Configuration {
	return configuration
}

// GetElasticsearchURL 获取ES连接URL
func (cfg *Configuration) GetElasticsearchURL() string {
	if cfg.Elasticsearch.URL != "" {
		return cfg.Elasticsearch.URL
	}
	return "http://localhost:9200" // 默认值
}

// GetElasticsearchIndexName 获取ES索引名
func (cfg *Configuration) GetElasticsearchIndexName() string {
	if cfg.Elasticsearch.IndexName != "" {
		return cfg.Elasticsearch.IndexName
	}
	return "bmtdblog_posts" // 默认值
}

// Redis相关方法

// 初始化Redis连接
func InitRedis() error {
	cfg := GetConfiguration()

	if !cfg.Redis.Enabled {
		slog.Info("Redis is disabled")
		return nil
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
		PoolSize: cfg.Redis.PoolSize,
	})

	// 测试连接
	ctx := context.Background()
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		slog.Error("Failed to connect to Redis", "error", err)
		return err
	}

	Redis = &RedisCacheClient{
		client: rdb,
		ctx:    ctx,
	}

	slog.Info("Redis connected successfully",
		"addr", cfg.Redis.Addr,
		"db", cfg.Redis.DB)

	return nil
}

// 检查Redis是否可用
func (r *RedisCacheClient) IsAvailable() bool {
	return r != nil && r.client != nil
}

// 设置缓存
func (r *RedisCacheClient) Set(key string, value interface{}, expiration time.Duration) error {
	if !r.IsAvailable() {
		return fmt.Errorf("redis is not available")
	}

	jsonValue, err := json.Marshal(value)
	if err != nil {
		return err
	}

	err = r.client.Set(r.ctx, key, jsonValue, expiration).Err()
	if err != nil {
		slog.Error("Failed to set cache", "key", key, "error", err)
		return err
	}

	slog.Debug("Cache set successfully", "key", key, "expiration", expiration)
	return nil
}

// 获取缓存
func (r *RedisCacheClient) Get(key string, dest interface{}) error {
	if !r.IsAvailable() {
		return fmt.Errorf("redis is not available")
	}

	val, err := r.client.Get(r.ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("cache miss")
		}
		slog.Error("Failed to get cache", "key", key, "error", err)
		return err
	}

	err = json.Unmarshal([]byte(val), dest)
	if err != nil {
		slog.Error("Failed to unmarshal cache", "key", key, "error", err)
		return err
	}

	slog.Debug("Cache hit", "key", key)
	return nil
}

// 删除缓存
func (r *RedisCacheClient) Del(keys ...string) error {
	if !r.IsAvailable() {
		return fmt.Errorf("redis is not available")
	}

	err := r.client.Del(r.ctx, keys...).Err()
	if err != nil {
		slog.Error("Failed to delete cache", "keys", keys, "error", err)
		return err
	}

	slog.Debug("Cache deleted successfully", "keys", keys)
	return nil
}

// 延迟双删策略：先删缓存→更新数据库→延迟删除缓存
func (r *RedisCacheClient) DelayedDoubleDelete(keys []string, delay time.Duration) error {
	if !r.IsAvailable() {
		return fmt.Errorf("redis is not available")
	}

	// 第一次删除：删除缓存
	err := r.Del(keys...)
	if err != nil {
		slog.Error("Failed to delete cache in first phase", "keys", keys, "error", err)
		return err
	}

	// 延迟后第二次删除：确保数据一致性
	go func() {
		time.Sleep(delay)
		if delErr := r.Del(keys...); delErr != nil {
			slog.Error("Failed to delete cache in second phase", "keys", keys, "error", delErr)
		} else {
			slog.Debug("Delayed double delete completed", "keys", keys, "delay", delay)
		}
	}()

	slog.Debug("Delayed double delete initiated", "keys", keys, "delay", delay)
	return nil
}

// 根据模式获取所有匹配的键
func (r *RedisCacheClient) GetKeysByPattern(pattern string) ([]string, error) {
	if !r.IsAvailable() {
		return nil, fmt.Errorf("redis is not available")
	}

	keys, err := r.client.Keys(r.ctx, pattern).Result()
	if err != nil {
		slog.Error("Failed to get keys by pattern", "pattern", pattern, "error", err)
		return nil, err
	}

	return keys, nil
}

// 检查key是否存在
func (r *RedisCacheClient) Exists(key string) bool {
	if !r.IsAvailable() {
		return false
	}

	result, err := r.client.Exists(r.ctx, key).Result()
	if err != nil {
		slog.Error("Failed to check cache existence", "key", key, "error", err)
		return false
	}

	return result > 0
}

// SetNX 设置缓存，仅当key不存在时（分布式锁）
func (r *RedisCacheClient) SetNX(key string, value interface{}, expiration time.Duration) (bool, error) {
	if !r.IsAvailable() {
		return false, fmt.Errorf("redis is not available")
	}

	result, err := r.client.SetNX(r.ctx, key, value, expiration).Result()
	if err != nil {
		slog.Error("Failed to setnx", "key", key, "error", err)
		return false, err
	}

	if result {
		slog.Debug("SetNX successful", "key", key, "expiration", expiration)
	} else {
		slog.Debug("SetNX failed, key already exists", "key", key)
	}

	return result, nil
}

// 设置过期时间
func (r *RedisCacheClient) Expire(key string, expiration time.Duration) error {
	if !r.IsAvailable() {
		return fmt.Errorf("redis is not available")
	}

	err := r.client.Expire(r.ctx, key, expiration).Err()
	if err != nil {
		slog.Error("Failed to set expiration", "key", key, "error", err)
		return err
	}

	return nil
}

// 获取缓存TTL
func (r *RedisCacheClient) TTL(key string) (time.Duration, error) {
	if !r.IsAvailable() {
		return 0, fmt.Errorf("redis is not available")
	}

	ttl, err := r.client.TTL(r.ctx, key).Result()
	if err != nil {
		slog.Error("Failed to get TTL", "key", key, "error", err)
		return 0, err
	}

	return ttl, nil
}

// 清除所有缓存
func (r *RedisCacheClient) FlushAll() error {
	if !r.IsAvailable() {
		return fmt.Errorf("redis is not available")
	}

	err := r.client.FlushAll(r.ctx).Err()
	if err != nil {
		slog.Error("Failed to flush all cache", "error", err)
		return err
	}

	slog.Info("All cache flushed successfully")
	return nil
}

// 关闭Redis连接
func (r *RedisCacheClient) Close() error {
	if r.client != nil {
		return r.client.Close()
	}
	return nil
}

// 生成缓存key的辅助函数
func GenerateKey(prefix string, args ...interface{}) string {
	if len(args) == 0 {
		return prefix
	}
	return fmt.Sprintf("%s:%v", prefix, fmt.Sprintf("%v", args[0]))
}

// 批量删除key（支持模式匹配）
func (r *RedisCacheClient) DelPattern(pattern string) error {
	if !r.IsAvailable() {
		return fmt.Errorf("redis is not available")
	}

	keys, err := r.client.Keys(r.ctx, pattern).Result()
	if err != nil {
		slog.Error("Failed to get keys by pattern", "pattern", pattern, "error", err)
		return err
	}

	if len(keys) == 0 {
		slog.Debug("No keys found for pattern", "pattern", pattern)
		return nil
	}

	err = r.client.Del(r.ctx, keys...).Err()
	if err != nil {
		slog.Error("Failed to delete keys by pattern", "pattern", pattern, "error", err)
		return err
	}

	slog.Debug("Keys deleted by pattern", "pattern", pattern, "count", len(keys))
	return nil
}

// ================= ElasticSearch 相关代码 =================

// InitElasticsearch 初始化ElasticSearch客户端
func InitElasticsearch() error {
	cfg := elasticsearch.Config{
		Addresses: []string{
			GetConfiguration().GetElasticsearchURL(),
		},
		Username: GetConfiguration().Elasticsearch.Username, // 可选认证
		Password: GetConfiguration().Elasticsearch.Password, // 可选认证
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

	slog.Info("ElasticSearch连接成功")

	// 创建索引
	if err := createPostIndex(); err != nil {
		return fmt.Errorf("创建索引失败: %w", err)
	}

	return nil
}

// createPostIndex 创建博文索引
func createPostIndex() error {
	indexName := GetConfiguration().GetElasticsearchIndexName()

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

// CreateAutoCertManager 创建自动证书管理器
func CreateAutoCertManager(domain, email, certDir string) *autocert.Manager {
	if certDir == "" {
		certDir = "certs"
	}

	// 确保证书目录存在
	if err := os.MkdirAll(certDir, 0755); err != nil {
		Logger.Error("创建证书目录失败", "dir", certDir, "err", err)
		return nil
	}

	Logger.Info("配置自动证书管理", "domain", domain, "email", email, "certDir", certDir)

	return &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(domain),
		Cache:      autocert.DirCache(certDir),
		Email:      email,
	}
}

// StartWithAutoCert 使用自动证书启动HTTPS服务器
func StartWithAutoCert(srv *http.Server, domain, email, certDir string) error {
	m := CreateAutoCertManager(domain, email, certDir)
	if m == nil {
		return fmt.Errorf("failed to create autocert manager")
	}

	srv.TLSConfig = &tls.Config{
		GetCertificate: m.GetCertificate,
		NextProtos:     []string{"h2", "http/1.1"},
		MinVersion:     tls.VersionTLS12,
	}

	// 启动HTTP挑战服务器（端口80）
	go func() {
		Logger.Info("启动HTTP挑战服务器在端口80")
		if err := http.ListenAndServe(":80", m.HTTPHandler(nil)); err != nil {
			Logger.Error("HTTP挑战服务器错误", "err", err)
		}
	}()

	Logger.Info("启动HTTPS服务器", "addr", srv.Addr, "domain", domain)
	return srv.ListenAndServeTLS("", "")
}

// GeneratePrivateKeyScript 生成私钥生成脚本
func GeneratePrivateKeyScript(keyFile string) error {
	if keyFile == "" {
		keyFile = "server.key"
	}

	scriptContent := fmt.Sprintf(`#!/bin/bash
# 自动生成RSA私钥脚本
# 生成日期: %s

KEY_FILE="%s"
KEY_SIZE=2048

echo "正在生成RSA私钥..."
echo "密钥文件: $KEY_FILE"
echo "密钥长度: $KEY_SIZE bits"

# 检查是否已存在私钥文件
if [ -f "$KEY_FILE" ]; then
    echo "警告: 私钥文件 $KEY_FILE 已存在"
    read -p "是否覆盖现有文件? (y/N): " confirm
    if [[ $confirm != [yY] ]]; then
        echo "操作已取消"
        exit 0
    fi
fi

# 生成RSA私钥
openssl genrsa -out "$KEY_FILE" $KEY_SIZE

if [ $? -eq 0 ]; then
    echo "私钥生成成功: $KEY_FILE"
    
    # 设置安全权限 (仅所有者可读写)
    chmod 600 "$KEY_FILE"
    echo "私钥文件权限已设置为600"
    
    # 显示私钥信息
    echo "私钥信息:"
    openssl rsa -in "$KEY_FILE" -text -noout | head -10
    
    echo ""
    echo "注意事项:"
    echo "1. 请妥善保管此私钥文件"
    echo "2. 不要将私钥文件上传到公共代码仓库"
    echo "3. 定期备份私钥文件"
    echo "4. 如果私钥泄露，请立即重新生成"
else
    echo "错误: 私钥生成失败"
    exit 1
fi
`, time.Now().Format("2006-01-02 15:04:05"), keyFile)

	scriptFile := "generate_private_key.sh"
	if err := os.WriteFile(scriptFile, []byte(scriptContent), 0755); err != nil {
		return fmt.Errorf("生成私钥脚本失败: %v", err)
	}

	Logger.Info("私钥生成脚本已创建", "script", scriptFile)
	return nil
}

// GenerateSelfSignedCert 生成自签名证书（用于开发测试）
func GenerateSelfSignedCert(domain, certFile, keyFile string) error {
	if certFile == "" {
		certFile = "server.crt"
	}
	if keyFile == "" {
		keyFile = "server.key"
	}

	scriptContent := fmt.Sprintf(`#!/bin/bash
# 生成自签名证书脚本
# 生成日期: %s

DOMAIN="%s"
CERT_FILE="%s"
KEY_FILE="%s"
DAYS=365

echo "正在生成自签名SSL证书..."
echo "域名: $DOMAIN"
echo "证书文件: $CERT_FILE"
echo "私钥文件: $KEY_FILE"
echo "有效期: $DAYS 天"

# 生成私钥
echo "1. 生成RSA私钥..."
openssl genrsa -out "$KEY_FILE" 2048

if [ $? -ne 0 ]; then
    echo "错误: 私钥生成失败"
    exit 1
fi

# 生成自签名证书
echo "2. 生成自签名证书..."
openssl req -new -x509 -key "$KEY_FILE" -out "$CERT_FILE" -days $DAYS \
    -subj "/C=CN/ST=State/L=City/O=Organization/OU=OrgUnit/CN=$DOMAIN"

if [ $? -eq 0 ]; then
    echo "证书生成成功!"
    
    # 设置文件权限
    chmod 600 "$KEY_FILE"
    chmod 644 "$CERT_FILE"
    
    echo "文件权限已设置"
    echo ""
    echo "证书信息:"
    openssl x509 -in "$CERT_FILE" -text -noout | grep -E "(Subject:|Not Before|Not After|DNS:)"
    
    echo ""
    echo "使用方法:"
    echo "在配置文件中设置:"
    echo "[tls]"
    echo "enabled = true"
    echo "cert_file = \"$CERT_FILE\""
    echo "key_file = \"$KEY_FILE\""
    echo ""
    echo "注意: 自签名证书仅适用于开发测试环境"
    echo "      生产环境请使用权威CA签发的证书"
else
    echo "错误: 证书生成失败"
    exit 1
fi
`, time.Now().Format("2006-01-02 15:04:05"), domain, certFile, keyFile)

	scriptFile := "generate_self_signed_cert.sh"
	if err := os.WriteFile(scriptFile, []byte(scriptContent), 0755); err != nil {
		return fmt.Errorf("生成证书脚本失败: %v", err)
	}

	Logger.Info("自签名证书生成脚本已创建", "script", scriptFile)
	return nil
}

// StartServer 启动HTTP/HTTPS服务器
func StartServer(srv *http.Server) error {
	cfg := GetConfiguration()

	if !cfg.TLS.Enabled {
		Logger.Info("启动HTTP服务", "addr", cfg.Addr)
		return srv.ListenAndServe()
	}

	return startTLSServer(srv, cfg)
}

// startTLSServer 启动TLS服务器
func startTLSServer(srv *http.Server, cfg *Configuration) error {
	if cfg.TLS.AutoCert && cfg.TLS.Domain != "" {
		// 自动证书模式
		Logger.Info("启动自动HTTPS服务", "addr", cfg.Addr, "domain", cfg.TLS.Domain)
		return StartWithAutoCert(srv, cfg.TLS.Domain, cfg.TLS.Email, cfg.TLS.CertDir)
	}

	if cfg.TLS.CertFile != "" && cfg.TLS.KeyFile != "" {
		// 手动证书模式
		Logger.Info("启动HTTPS服务", "addr", cfg.Addr)
		return srv.ListenAndServeTLS(cfg.TLS.CertFile, cfg.TLS.KeyFile)
	}

	// 配置不完整，回退到HTTP
	Logger.Error("TLS配置错误: 需要配置自动证书(auto_cert+domain)或手动证书(cert_file+key_file)")
	Logger.Info("TLS配置不完整，回退到HTTP服务", "addr", cfg.Addr)
	return srv.ListenAndServe()
}

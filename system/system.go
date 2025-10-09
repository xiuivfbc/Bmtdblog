package system

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
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

	Configuration struct {
		Addr          string      `mapstructure:"addr"`
		SignupEnabled bool        `mapstructure:"signup_enabled"`
		Title         string      `mapstructure:"title"`
		SessionSecret string      `mapstructure:"session_secret"`
		Domain        string      `mapstructure:"domain"`
		FileServer    string      `mapstructure:"file_server"`
		NotifyEmails  string      `mapstructure:"notify_emails"`
		PageSize      int         `mapstructure:"page_size"`
		PublicDir     string      `mapstructure:"public"`
		ViewDir       string      `mapstructure:"view"`
		Dir           string      `mapstructure:"dir"`
		Database      Database    `mapstructure:"database"`
		Seo           Seo         `mapstructure:"seo"`
		Qiniu         Qiniu       `mapstructure:"qiniu"`
		Smms          Smms        `mapstructure:"smms"`
		Redis         RedisConfig `mapstructure:"redis"`
		Github        Github      `mapstructure:"github"`
		Smtp          Smtp        `mapstructure:"smtp"`
		Navigators    []Navigator `mapstructure:"navigators"`
		Backup        Backup      `mapstructure:"backup"`
	}
)

// Redis客户端结构体
type RedisCacheClient struct {
	client *redis.Client
	ctx    context.Context
}

var (
	Logger *slog.Logger
	Redis  *RedisCacheClient
)

func (a Author) String() string {
	return fmt.Sprintf("%s,%s", a.Name, a.Email)
}

var configuration *Configuration

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

package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/xiuivfbc/bmtdblog/internal/common/log"
	"github.com/xiuivfbc/bmtdblog/internal/config"
)

// Redis客户端结构体
type RedisCacheClient struct {
	client *redis.Client
	ctx    context.Context
}

// 全局Redis实例
var redisclient *RedisCacheClient

// 初始化Redis连接
func InitRedis(conf config.RedisConfig) error {

	password := conf.Password

	rdb := redis.NewClient(&redis.Options{
		Addr:     conf.Addr,
		Password: password,
		DB:       conf.DB,
		PoolSize: conf.PoolSize,
	})

	// 测试连接
	ctx := context.Background()
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Error("Failed to connect to Redis", "error", err)
		return err
	}

	redisclient = &RedisCacheClient{
		client: rdb,
		ctx:    ctx,
	}

	log.Info("Redis connected successfully",
		"addr", conf.Addr,
		"db", conf.DB)

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
		log.Error("Failed to set cache", "key", key, "error", err)
		return err
	}

	log.Debug("Cache set successfully", "key", key, "expiration", expiration)
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
		log.Error("Failed to get cache", "key", key, "error", err)
		return err
	}

	err = json.Unmarshal([]byte(val), dest)
	if err != nil {
		log.Error("Failed to unmarshal cache", "key", key, "error", err)
		return err
	}

	log.Debug("Cache hit", "key", key)
	return nil
}

// 删除缓存
func (r *RedisCacheClient) Del(keys ...string) error {
	if !r.IsAvailable() {
		return fmt.Errorf("redis is not available")
	}

	err := r.client.Del(r.ctx, keys...).Err()
	if err != nil {
		log.Error("Failed to delete cache", "keys", keys, "error", err)
		return err
	}

	log.Debug("Cache deleted successfully", "keys", keys)
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
		log.Error("Failed to delete cache in first phase", "keys", keys, "error", err)
		return err
	}

	// 延迟后第二次删除：确保数据一致性
	go func() {
		time.Sleep(delay)
		if delErr := r.Del(keys...); delErr != nil {
			log.Error("Failed to delete cache in second phase", "keys", keys, "error", delErr)
		} else {
			log.Debug("Delayed double delete completed", "keys", keys, "delay", delay)
		}
	}()

	log.Debug("Delayed double delete initiated", "keys", keys, "delay", delay)
	return nil
}

// 根据模式获取所有匹配的键
func (r *RedisCacheClient) GetKeysByPattern(pattern string) ([]string, error) {
	if !r.IsAvailable() {
		return nil, fmt.Errorf("redis is not available")
	}

	keys, err := r.client.Keys(r.ctx, pattern).Result()
	if err != nil {
		log.Error("Failed to get keys by pattern", "pattern", pattern, "error", err)
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
		log.Error("Failed to check cache existence", "key", key, "error", err)
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
		log.Error("Failed to setnx", "key", key, "error", err)
		return false, err
	}

	if result {
		log.Debug("SetNX successful", "key", key, "expiration", expiration)
	} else {
		log.Debug("SetNX failed, key already exists", "key", key)
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
		log.Error("Failed to set expiration", "key", key, "error", err)
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
		log.Error("Failed to get TTL", "key", key, "error", err)
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
		log.Error("Failed to flush all cache", "error", err)
		return err
	}

	log.Info("All cache flushed successfully")
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
		log.Error("Failed to get keys by pattern", "pattern", pattern, "error", err)
		return err
	}

	if len(keys) == 0 {
		log.Debug("No keys found for pattern", "pattern", pattern)
		return nil
	}

	err = r.client.Del(r.ctx, keys...).Err()
	if err != nil {
		log.Error("Failed to delete keys by pattern", "pattern", pattern, "error", err)
		return err
	}

	log.Debug("Keys deleted by pattern", "pattern", pattern, "count", len(keys))
	return nil
}

func GetRedis() *RedisCacheClient {
	return redisclient
}

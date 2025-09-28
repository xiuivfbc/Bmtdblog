# Redis缓存实现说明

## 功能概述

本次实现为Bmtdblog项目添加了Redis缓存功能，主要用于博文数据的缓存，以减少数据库压力和提升响应速度。Redis功能已集成到`system`包中，统一管理系统组件。

## 实现的功能

### 1. Redis配置
- 在 `system/system.go` 中添加了Redis配置结构体和Redis客户端
- 在 `conf/conf.toml` 中添加了Redis连接配置
- 支持Redis连接池配置

### 2. 缓存服务 (`system/system.go`)
- Redis连接管理
- 基本缓存操作（Get, Set, Del, Exists等）
- 支持键值过期时间设置
- 模式匹配批量删除
- 错误处理和日志记录

### 3. 博文缓存逻辑 (`models/post.go`)
- `GetPostByIdWithCache()`: 缓存优先的博文获取方法
- `SetCache()`: 将博文数据存入缓存
- `DelCache()`: 删除博文缓存
- `ClearRelatedCache()`: 清除相关缓存（列表、归档等）
- 在增删改操作中自动处理缓存更新

### 4. 控制器优化
- 修改 `PostGet`, `PostEdit`, `PostToggle` 使用缓存版本
- 异步更新浏览数，避免影响缓存性能
- 评论相关控制器也使用缓存优先逻辑

## 配置说明

### Redis配置项 (conf/conf.toml)
```toml
[redis]
enabled = true              # 是否启用Redis缓存
addr = '127.0.0.1:6379'    # Redis服务器地址
password = ''               # Redis密码（可选）
db = 0                      # 数据库编号
pool_size = 10              # 连接池大小
```

## 缓存策略

### 1. 缓存Key设计
- 单个博文: `post:{{id}}`
- 博文列表: `post_list:*`
- 归档页面: `post_archive:*`

### 2. 过期时间
- 单个博文缓存: 1小时
- 列表缓存: 30分钟

### 3. 缓存更新策略
- **写入时清除** (Write-Through): 博文增删改时自动清除相关缓存
- **延迟加载** (Lazy Loading): 缓存未命中时从数据库获取并异步写入缓存
- **异步更新**: 浏览数更新等操作异步执行，不影响缓存性能

## 使用方法

### 1. 启动Redis服务
```bash
# Ubuntu/Debian
sudo service redis-server start

# 或使用Docker
docker run -d -p 6379:6379 redis:alpine
```

### 2. 修改配置
在 `conf/conf.toml` 中启用Redis：
```toml
[redis]
enabled = true
addr = '127.0.0.1:6379'
```

### 3. 重启应用
重启Bmtdblog应用，Redis缓存将自动生效。

## 缓存效果

### 性能提升预期
- **首次访问**: 从数据库加载（正常速度）
- **缓存命中**: 响应时间提升 50-90%
- **数据库压力**: 减少 60-80% 的查询请求

### 监控方法
应用日志中会记录：
- `Cache hit`: 缓存命中
- `Cache miss`: 缓存未命中
- `Cache set/delete`: 缓存操作日志

## 故障处理

### 1. Redis连接失败
- 应用不会因Redis连接失败而退出
- 自动降级为直接数据库访问
- 日志中会记录Redis连接错误

### 2. 缓存数据异常
手动清空所有缓存：
```bash
redis-cli FLUSHALL
```

### 3. 性能调优
根据实际情况调整：
- `pool_size`: 连接池大小
- `PostCacheExpiration`: 缓存过期时间
- 缓存key的设计模式

## 扩展建议

### 1. 列表缓存
为首页文章列表、标签页面等添加缓存支持

### 2. 缓存预热
应用启动时预加载热门文章到缓存

### 3. 分布式缓存
使用Redis Cluster支持水平扩展

### 4. 缓存监控
添加缓存命中率、响应时间等监控指标

## 注意事项

1. **数据一致性**: 缓存更新策略确保数据一致性
2. **内存使用**: 根据博文数量调整Redis内存配置
3. **网络延迟**: Redis服务器尽量部署在同一网络
4. **备份策略**: Redis数据建议定期备份

## 测试验证

### 验证缓存工作
1. 访问博文页面，查看日志是否有 "Post loaded from cache"
2. 使用 `redis-cli KEYS "*"` 查看缓存key
3. 修改博文后验证相关缓存是否清除

### 性能测试
可使用ab、wrk等工具测试缓存前后的性能差异。
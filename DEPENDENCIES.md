# 依赖更新说明

添加Redis缓存功能后，需要更新Go依赖包。

## 需要添加的依赖

在项目根目录运行以下命令添加Redis客户端：

```bash
go get github.com/redis/go-redis/v9
```

## 更新后的依赖

项目现在依赖：
- `github.com/redis/go-redis/v9` - Redis客户端库
- 原有的所有依赖保持不变

## 验证依赖安装

运行以下命令验证：
```bash
go mod tidy
go build .
```

如果编译成功，说明依赖安装正确。
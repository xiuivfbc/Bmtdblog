# 多阶段构建
# 第一阶段：构建应用
FROM golang:1.24-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装必要的工具
RUN apk add --no-cache git

# 复制go mod文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bmtdblog .

# 第二阶段：运行时镜像
FROM alpine:latest

# 安装ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates tzdata

# 设置时区
ENV TZ=Asia/Shanghai

WORKDIR /root/

# 从构建阶段复制二进制文件
COPY --from=builder /app/bmtdblog .

# 复制配置文件和静态资源
COPY --from=builder /app/conf ./conf
COPY --from=builder /app/views ./views
COPY --from=builder /app/static ./static

# 创建必要的目录
RUN mkdir -p logs slog

# 暴露端口
EXPOSE 8090

# 启动命令
CMD ["./bmtdblog", "-C", "conf/conf_docker.toml"]
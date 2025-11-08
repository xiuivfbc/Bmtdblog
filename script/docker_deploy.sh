#!/bin/bash

echo "🚀 开始Docker部署bmtdblog..."

# 停止现有容器
echo "📦 停止现有容器..."
docker-compose down

# 清理旧镜像（可选）
echo "🧹 清理旧镜像..."
docker image prune -f

# 构建并启动服务
echo "🔨 构建并启动服务..."
docker-compose up --build -d

# 检查服务状态
echo "📊 检查服务状态..."
sleep 10
docker-compose ps

# 查看日志
echo "📝 服务日志："
docker-compose logs --tail=20

echo "✅ 部署完成！"
echo "🌐 访问地址: http://115.120.208.110:8090"
echo "📊 监控命令: docker-compose logs -f bmtdblog"
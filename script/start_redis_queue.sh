#!/bin/bash

# Redis邮件队列启动脚本
# 使用优化的持久化配置启动Redis

REDIS_CONF="../conf/redis-email-queue.conf"
REDIS_LOG="../logs/redis-email-queue.log"
REDIS_PID="./redis.pid"

# 创建日志目录
mkdir -p logs

echo "🚀 启动Redis邮件队列服务..."
echo "配置文件: $REDIS_CONF"
echo "日志文件: $REDIS_LOG"

# 检查Redis是否已经运行
if [ -f "$REDIS_PID" ]; then
    PID=$(cat $REDIS_PID)
    if ps -p $PID > /dev/null 2>&1; then
        echo "⚠ Redis已经在运行 (PID: $PID)"
        exit 1
    else
        echo "清理过期的PID文件..."
        rm -f $REDIS_PID
    fi
fi

# 启动Redis
redis-server $REDIS_CONF --daemonize yes --pidfile $REDIS_PID

# 等待启动
sleep 2

# 检查启动状态
if [ -f "$REDIS_PID" ]; then
    PID=$(cat $REDIS_PID)
    if ps -p $PID > /dev/null 2>&1; then
        echo "✅ Redis启动成功 (PID: $PID)"
        echo "📊 检查持久化配置..."
        
        # 检查AOF配置
        AOF_STATUS=$(redis-cli CONFIG GET appendonly | tail -n 1)
        echo "   AOF持久化: $AOF_STATUS"
        
        if [ "$AOF_STATUS" = "yes" ]; then
            FSYNC_STATUS=$(redis-cli CONFIG GET appendfsync | tail -n 1)
            echo "   同步策略: $FSYNC_STATUS"
        fi
        
        # 检查混合持久化
        HYBRID_STATUS=$(redis-cli CONFIG GET aof-use-rdb-preamble | tail -n 1)
        echo "   混合持久化: $HYBRID_STATUS"
        
        echo ""
        echo "🎯 Redis邮件队列服务已启动，数据持久化已配置！"
        echo "💡 使用 'redis-cli shutdown' 停止服务"
        
    else
        echo "❌ Redis启动失败"
        exit 1
    fi
else
    echo "❌ Redis启动失败，未找到PID文件"
    exit 1
fi
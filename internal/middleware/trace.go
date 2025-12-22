package middleware

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/xiuivfbc/bmtdblog/internal/common/log"
	"github.com/xiuivfbc/bmtdblog/internal/config"
	"go.uber.org/zap"
)

// 定义上下文键
type contextKey string

const (
	TraceIDKey contextKey = "trace_id"
	GinTraceID string     = "X-Trace-ID"
)

// TraceMiddleware 追踪中间件，为每个请求生成唯一的 trace_id
func TraceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := "NULL"

		if config.GetConfiguration().Zap.EnableTraceID {
			traceID = uuid.New().String()
		}

		// 设置到 Gin 上下文
		c.Set(string(TraceIDKey), traceID)

		// 记录请求开始
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// 请求开始日志
		log.DaemonInfo("middleware", "trace", "Request started",
			zap.String("trace_id", traceID),
			zap.String("method", method),
			zap.String("path", path),
			zap.String("client_ip", c.ClientIP()),
		)

		c.Next()

		// 请求结束日志
		latency := time.Since(start)
		status := c.Writer.Status()

		log.DaemonInfo("middleware", "trace", "Request completed",
			zap.String("trace_id", traceID),
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", status),
			zap.String("latency", latency.String()),
		)
	}
}

// GetTraceID 从 Gin 上下文获取 trace_id
func GetTraceID(c *gin.Context) string {
	if traceID, exists := c.Get(string(TraceIDKey)); exists {
		return traceID.(string)
	}
	return ""
}

// GetTraceIDFromContext 从标准 context 获取 trace_id
func GetTraceIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if traceID, ok := ctx.Value(TraceIDKey).(string); ok {
		return traceID
	}
	return ""
}

// WithTraceID 将 trace_id 注入到标准 context
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, TraceIDKey, traceID)
}

// ContextFromGin 从 Gin 上下文创建带 trace_id 的标准 context
func ContextFromGin(c *gin.Context) context.Context {
	traceID := GetTraceID(c)
	return WithTraceID(c.Request.Context(), traceID)
}

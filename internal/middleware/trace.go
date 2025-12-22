package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/xiuivfbc/bmtdblog/internal/common/log"
	"github.com/xiuivfbc/bmtdblog/internal/config"
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

		c.Set(string(TraceIDKey), traceID)

		start := time.Now()

		// 创建带trace_id的context并设置到goroutine局部存储
		ctx := ContextFromGin(c)
		log.SetGoroutineContext(ctx)
		defer log.DeleteGoroutineContext()

		c.Next()

		// 获取请求参数
		paramStr := ""
		if c.Request.Method == "GET" {
			paramStr = c.Request.URL.RawQuery
		} else {
			c.Request.ParseForm()
			paramStr = c.Request.Form.Encode()
		}

		log.Info(fmt.Sprintf(
			"Request complete: method=%s path=%s | status=%d | cost=%s | params=%s",
			c.Request.Method,
			c.Request.URL.Path,
			c.Writer.Status(),
			time.Since(start),
			paramStr,
		))
	}
}

// ContextFromGin 从 Gin 上下文创建带 trace_id 的标准 context
func ContextFromGin(c *gin.Context) context.Context {
	traceID := GetTraceID(c)
	return WithTraceID(c.Request.Context(), traceID)
}

// GetTraceID 从 Gin 上下文获取 trace_id
func GetTraceID(c *gin.Context) string {
	if traceID, exists := c.Get(string(TraceIDKey)); exists {
		return traceID.(string)
	}
	return ""
}

// WithTraceID 将 trace_id 注入到标准 context
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, TraceIDKey, traceID)
}

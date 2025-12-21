package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
		// 优先从请求头获取 trace_id，支持跨服务传递
		traceID := c.GetHeader(GinTraceID)
		if traceID == "" {
			traceID = uuid.New().String()
		}

		// 设置到 Gin 上下文
		c.Set(string(TraceIDKey), traceID)

		// 设置响应头，便于调试
		c.Header(GinTraceID, traceID)

		// 记录请求开始
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// 请求开始日志
		config.Logger.Info("Request started",
			"trace_id", traceID,
			"method", method,
			"path", path,
			"client_ip", c.ClientIP(),
		)

		c.Next()

		// 请求结束日志
		latency := time.Since(start)
		status := c.Writer.Status()

		config.Logger.Info("Request completed",
			"trace_id", traceID,
			"method", method,
			"path", path,
			"status", status,
			"latency", latency.String(),
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

// LogWithTrace 带 trace_id 的日志记录器
func LogWithTrace(c *gin.Context) *slog.Logger {
	traceID := GetTraceID(c)
	if traceID == "" {
		return config.Logger
	}
	return config.Logger.With("trace_id", traceID)
}

// LogWithContext 从标准 context 获取带 trace_id 的日志记录器
func LogWithContext(ctx context.Context) *slog.Logger {
	traceID := GetTraceIDFromContext(ctx)
	if traceID == "" {
		return config.Logger
	}
	return config.Logger.With("trace_id", traceID)
}

// getCallerInfo 获取调用者信息（文件路径、函数名、行号）
// skip 表示跳过的调用栈层数
func getCallerInfo(skip int) string {
	pc, filePath, line, ok := runtime.Caller(skip)
	if !ok {
		return "unknown/unknown/0"
	}

	// 获取函数名
	var funcName string
	fn := runtime.FuncForPC(pc)
	if fn != nil {
		funcName = filepath.Base(fn.Name())
	} else {
		funcName = "unknown"
	}

	// 获取相对路径（只保留文件名）
	file := filepath.Base(filePath)

	// 格式：文件名/函数名/行号
	return fmt.Sprintf("%s/%s/%d", file, funcName, line)
}

// addCallerArgs 添加调用者信息到参数列表
func addCallerArgs(skip int, args []any) []any {
	location := getCallerInfo(skip)
	return append(args, "location", location)
}

// LogInfo 带 trace_id 和调用者信息的 Info 日志
func LogInfo(c *gin.Context, msg string, args ...any) {
	args = addCallerArgs(2, args)
	LogWithTrace(c).Info(msg, args...)
}

// LogError 带 trace_id 和调用者信息的 Error 日志
func LogError(c *gin.Context, msg string, args ...any) {
	args = addCallerArgs(2, args)
	LogWithTrace(c).Error(msg, args...)
}

// LogWarn 带 trace_id 和调用者信息的 Warn 日志
func LogWarn(c *gin.Context, msg string, args ...any) {
	args = addCallerArgs(2, args)
	LogWithTrace(c).Warn(msg, args...)
}

// LogDebug 带 trace_id 和调用者信息的 Debug 日志
func LogDebug(c *gin.Context, msg string, args ...any) {
	args = addCallerArgs(2, args)
	LogWithTrace(c).Debug(msg, args...)
}

// ========== Context 版本的日志快捷函数 ==========

// LogInfoCtx 带 trace_id 和调用者信息的 Info 日志（Context版本）
func LogInfoCtx(ctx context.Context, msg string, args ...any) {
	args = addCallerArgs(2, args)
	LogWithContext(ctx).Info(msg, args...)
}

// LogErrorCtx 带 trace_id 和调用者信息的 Error 日志（Context版本）
func LogErrorCtx(ctx context.Context, msg string, args ...any) {
	args = addCallerArgs(2, args)
	LogWithContext(ctx).Error(msg, args...)
}

// LogWarnCtx 带 trace_id 和调用者信息的 Warn 日志（Context版本）
func LogWarnCtx(ctx context.Context, msg string, args ...any) {
	args = addCallerArgs(2, args)
	LogWithContext(ctx).Warn(msg, args...)
}

// LogDebugCtx 带 trace_id 和调用者信息的 Debug 日志（Context版本）
func LogDebugCtx(ctx context.Context, msg string, args ...any) {
	args = addCallerArgs(2, args)
	LogWithContext(ctx).Debug(msg, args...)
}

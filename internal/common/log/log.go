package log

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/xiuivfbc/bmtdblog/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// 定义上下文键
type contextKey string

const (
	TraceIDKey contextKey = "trace_id"
	GinTraceID string     = "X-Trace-ID"
)


// GLSData 存储goroutine的上下文数据
type GLSData struct {
	TraceID string
}

// glsMap 全局goroutine局部存储，以goroutine ID为键
var (
	Logger *zap.Logger

	glsMap = make(map[uint64]*GLSData)
	glsMu  sync.RWMutex

	DaemonDebug = newDaemonLogger(zapcore.DebugLevel, 1)
	DaemonInfo  = newDaemonLogger(zapcore.InfoLevel, 1)
	DaemonWarn  = newDaemonLogger(zapcore.WarnLevel, 1)
	DaemonError = newDaemonLogger(zapcore.ErrorLevel, 1)
	DaemonFatal = newDaemonLogger(zapcore.FatalLevel, 1)

	Debug  = newLogger(zapcore.DebugLevel, 1)
	Info   = newLogger(zapcore.InfoLevel, 1)
	Warn   = newLogger(zapcore.WarnLevel, 1)
	Error  = newLogger(zapcore.ErrorLevel, 1)
	Fatal  = newLogger(zapcore.FatalLevel, 1)
	DPanic = newLogger(zapcore.DPanicLevel, 1)
)


// Init 初始化日志系统
func Init() error {
	cfg := config.GetConfiguration()
	zapCfg := cfg.Zap

	// 创建日志目录
	// if err := os.MkdirAll(zapCfg.OutputPath, 0755); err != nil {
	// 	return err
	// }

	// 日志文件名 - 每次启动覆写原文件
	logFile := filepath.Join(zapCfg.OutputPath, "Bmtdblog.log")

	// 设置日志级别
	var level zapcore.Level
	switch zapCfg.Level {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	default:
		level = zapcore.InfoLevel
	}

	// 配置编码器
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder, // 显示相对路径+函数+行数
	}

	// 创建文件写入器 - 每次启动覆写原文件
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	// 配置Core - 只输出到文件
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(file),
		level,
	)

	// 创建Logger
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	// 设置全局Logger
	Logger = logger
	zap.ReplaceGlobals(logger)

	return nil
}


func processArgs(args []interface{}) []zap.Field {
	fields := make([]zap.Field, len(args))
	for i, arg := range args {
		// 如果是 zap.Field 类型，直接使用；否则用 zap.Any 包装，key 为索引
		if f, ok := arg.(zap.Field); ok {
			fields[i] = f
		} else {
			fields[i] = zap.Any(fmt.Sprintf("%d", i), arg)
		}
	}
	return fields
}

func newDaemonLogger(level zapcore.Level, skip int) func(daemonName, daemonID, msg string, args ...interface{}) {
	return func(daemonName, daemonID, msg string, args ...interface{}) {
		var fields []zap.Field
		if len(args) > 0 {
			fields = processArgs(args)
		}

		daemonFields := []zap.Field{
			zap.String("daemon_name", daemonName),
			zap.String("daemon_id", daemonID),
		}
		fields = append(daemonFields, fields...)

		Logger.WithOptions(zap.AddCallerSkip(skip)).Check(level, msg).Write(fields...)
	}
}

func newLogger(level zapcore.Level, skip int) func(msg string, args ...interface{}) {
	return func(msg string, args ...interface{}) {
		var fields []zap.Field
		if len(args) > 0 {
			fields = processArgs(args)
		}

		// 自动从goroutine上下文获取trace_id
		if data := GetGoroutineContext(); data != nil && data.TraceID != "" {
			fields = append(fields, zap.String("trace_id", data.TraceID))
		}

		Logger.WithOptions(zap.AddCallerSkip(skip)).Check(level, msg).Write(fields...)
	}
}

// GetGoroutineContext 获取goroutine上下文
func GetGoroutineContext() *GLSData {
	gid := getGoroutineID()
	glsMu.RLock()
	defer glsMu.RUnlock()
	return glsMap[gid]
}

// getGoroutineID 获取当前goroutine ID
func getGoroutineID() uint64 {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	idField := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
	id, _ := strconv.ParseUint(idField, 10, 64)
	return id
}

// SetGoroutineContext 设置goroutine上下文
func SetGoroutineContext(ctx context.Context) {
	if traceID := GetTraceIDFromContext(ctx); traceID != "" {
		gid := getGoroutineID()
		glsMu.Lock()
		glsMap[gid] = &GLSData{TraceID: traceID}
		glsMu.Unlock()
	}
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


// DeleteGoroutineContext 删除goroutine上下文
func DeleteGoroutineContext() {
	gid := getGoroutineID()
	glsMu.Lock()
	delete(glsMap, gid)
	glsMu.Unlock()
}

// GetLogger 获取日志记录器
func GetLogger() *zap.Logger {
	return Logger
}
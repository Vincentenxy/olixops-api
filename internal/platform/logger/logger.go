// Package logger 提供基于 Zap 的日志器封装。
//
// 业务代码通过 logger.L() 获取全局 logger,也可以从 context 中取出带 trace_id 的 logger。
package logger

import (
	"context"
	"fmt"
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"olixops/internal/config"
)

type ctxKey struct{}

var (
	global *zap.Logger
)

// mergedCore 包装 zapcore.Core,在 Write 时把累积的 fields 和本次调用的 fields 合并。
//
// zap v1.27.0 重构了 With 语义: ioCore.With 把 fields 存到 c.fields 但 ioCore.Write
// 不再合并 c.fields 直接传给 encoder,导致 With(...) 加的字段全部丢失。
// 本包装层把 mergedCore 自身注册到 ce.cores,在 Write 时手动合并,恢复旧版行为。
type mergedCore struct {
	zapcore.Core
	fields []zapcore.Field
}

func (c *mergedCore) With(fields []zapcore.Field) zapcore.Core {
	merged := make([]zapcore.Field, 0, len(c.fields)+len(fields))
	merged = append(merged, c.fields...)
	merged = append(merged, fields...)
	return &mergedCore{Core: c.Core.With(fields), fields: merged}
}

// Check 把 mergedCore 自己注册到 ce.cores,这样写入时才走 mergedCore.Write。
// 注意:不能调 c.Core.Check,否则 ioCore 会注册进去,Write 就跳过了 mergedCore。
func (c *mergedCore) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Core.Enabled(ent.Level) {
		return ce.AddCore(ent, c)
	}
	return ce
}

func (c *mergedCore) Write(ent zapcore.Entry, fields []zapcore.Field) error {
	merged := make([]zapcore.Field, 0, len(c.fields)+len(fields))
	merged = append(merged, c.fields...)
	merged = append(merged, fields...)
	return c.Core.Write(ent, merged)
}

// New 根据配置创建 Zap logger。
func New(cfg config.LogConfig, appName, env string) (*zap.Logger, error) {
	level, err := parseLevel(cfg.Level)
	if err != nil {
		return nil, err
	}

	encoder := buildEncoder(cfg.Format)
	writer, err := buildWriter(cfg)
	if err != nil {
		return nil, err
	}

	core := zapcore.NewCore(encoder, writer, level)
	// 用 mergedCore 包装,把 app/env 作为初始 fields。
	// 后续 logger.With(...) 加的字段也会累积进来,在 Write 时合并传给 encoder。
	wrapped := &mergedCore{
		Core: core,
		fields: []zapcore.Field{
			zap.String("app", appName),
			zap.String("env", env),
		},
	}

	logger := zap.New(wrapped,
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
	global = logger
	return logger, nil
}

// L 返回全局 logger,未初始化时返回 NopLogger。
func L() *zap.Logger {
	if global == nil {
		return zap.NewNop()
	}
	return global
}

// WithContext 将 logger 注入 context。
func WithContext(ctx context.Context, l *zap.Logger) context.Context {
	if l == nil {
		return ctx
	}
	return context.WithValue(ctx, ctxKey{}, l)
}

// FromContext 从 context 取出 logger,缺省返回全局 logger。
func FromContext(ctx context.Context) *zap.Logger {
	if ctx == nil {
		return L()
	}
	if l, ok := ctx.Value(ctxKey{}).(*zap.Logger); ok && l != nil {
		return l
	}
	return L()
}

// Sync 在程序退出前刷盘。
func Sync() {
	if global != nil {
		_ = global.Sync()
	}
}

func parseLevel(s string) (zapcore.Level, error) {
	switch strings.ToLower(s) {
	case "debug":
		return zapcore.DebugLevel, nil
	case "info", "":
		return zapcore.InfoLevel, nil
	case "warn", "warning":
		return zapcore.WarnLevel, nil
	case "error":
		return zapcore.ErrorLevel, nil
	default:
		return zapcore.InfoLevel, fmt.Errorf("unknown log level: %s", s)
	}
}

func buildEncoder(format string) zapcore.Encoder {
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "ts"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderCfg.EncodeDuration = zapcore.MillisDurationEncoder
	encoderCfg.EncodeCaller = zapcore.ShortCallerEncoder

	if strings.EqualFold(format, "console") {
		encoderCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
		// 自定义 console encoder: 把 trace_id 字段从结构化输出里抽出来,
		// 单独放在 caller 之后、msg 之前,缺省时输出 []。
		return newTraceConsoleEncoder(encoderCfg)
	}
	encoderCfg.EncodeLevel = zapcore.LowercaseLevelEncoder
	return zapcore.NewJSONEncoder(encoderCfg)
}

func buildWriter(cfg config.LogConfig) (zapcore.WriteSyncer, error) {
	switch strings.ToLower(cfg.Output) {
	case "stderr":
		return zapcore.AddSync(os.Stderr), nil
	case "file":
		if cfg.FilePath == "" {
			return nil, fmt.Errorf("log.file_path is required when log.output=file")
		}
		f, err := os.OpenFile(cfg.FilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			return nil, fmt.Errorf("open log file: %w", err)
		}
		return zapcore.AddSync(f), nil
	default:
		return zapcore.AddSync(os.Stdout), nil
	}
}

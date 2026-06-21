// 演示 internal/platform/logger 的用法。
//
// 用法:
//
//	go run ./cmd/logger-demo
//
// 该示例不依赖配置文件,通过环境变量或默认值控制输出:
//
//	OLIXOPS_LOG_LEVEL=debug \
//	OLIXOPS_LOG_FORMAT=console \
//	OLIXOPS_LOG_OUTPUT=stdout \
//	go run ./cmd/logger-demo
package main

import (
	"context"
	"errors"
	"os"

	"go.uber.org/zap"

	"olixops/internal/config"
	"olixops/internal/platform/logger"
)

func main() {
	// 1) 加载配置并初始化 logger。
	cfg, err := config.Load("")
	if err != nil {
		// 配置加载阶段 logger 还没就绪,只能退化到标准库。
		os.Stderr.WriteString("load config: " + err.Error() + "\n")
		os.Exit(1)
	}

	if _, err := logger.New(cfg.Log, cfg.App.Name, cfg.App.Env); err != nil {
		os.Stderr.WriteString("init logger: " + err.Error() + "\n")
		os.Exit(1)
	}
	defer logger.Sync() // 程序退出前刷盘,避免丢失缓冲日志。

	// 2) 直接使用全局 logger: 简单场景够用。
	demoGlobalLogger()

	// 3) 基于 context 传递 logger: 业务/接口层推荐写法。
	demoContextLogger()

	// 4) 模拟一个带 trace-id 的请求链路。
	demoRequestFlow()
}

// demoGlobalLogger 演示直接通过 L() 拿全局 logger。
// 适用:启动期初始化、一次性脚本、不接收 context 的工具函数。
func demoGlobalLogger() {
	log := logger.L()

	log.Info("server boot complete",
		zap.Int("pid", os.Getpid()),
	)

	log.Debug("debug message example",
		zap.String("hint", "只有 OLIXOPS_LOG_LEVEL=debug 才看得见"),
	)
}

// demoContextLogger 演示 context 注入的典型用法:
//
//   - service / repository 内部统一用 FromContext(ctx) 拿 logger。
//   - 这样中间件可以把 trace-id 等字段挂在 logger 上,业务代码不感知。
func demoContextLogger() {
	// 构造一个携带业务字段的子 logger。
	log := logger.L().With(
		zap.String("module", "user"),
		zap.String("action", "create"),
	)

	ctx := logger.WithContext(context.Background(), log)

	// 在更深层调用中,FromContext 会拿到这个带字段的 logger。
	logger.FromContext(ctx).Info("user create start",
		zap.String("userId", "u-1001"),
	)

	// 模拟一段耗时操作后再打一条。
	simulateWork(ctx)
}

func simulateWork(ctx context.Context) {
	// 没有传额外参数:FromContext 仍然能拿到上一步注入的 logger。
	logger.FromContext(ctx).Info("user create persist ok",
		zap.String("userId", "u-1001"),
		zap.Int64("costMs", 23),
	)
}

// demoRequestFlow 模拟一个 HTTP 请求的日志链:
//
//   - 入口中间件生成 trace-id,挂到 logger 上塞进 context。
//   - 业务层只调 FromContext(ctx),每条日志自动带 trace_id。
//   - 发生错误时 Error 级日志,带 stacktrace(由 logger.go 配置 zap.AddStacktrace(ErrorLevel) 自动开启)。
func demoRequestFlow() {
	traceID := "trace-7c2a8b9e-1f4d-4d11-bb62-3a5e9f0c1a22"

	log := logger.L().With(
		zap.String("trace_id", traceID),
		zap.String("method", "POST"),
		zap.String("path", "/api/v1/projects/create"),
	)
	ctx := logger.WithContext(context.Background(), log)

	logger.FromContext(ctx).Info("request received",
		zap.String("remote", "10.0.0.8"),
	)

	// 业务正常路径。
	logger.FromContext(ctx).Info("project created",
		zap.String("projectId", "p-9001"),
	)

	// 业务异常路径:Error 级会附带 stacktrace。
	if err := errors.New("namespace quota exceeded"); err != nil {
		logger.FromContext(ctx).Error("project create failed",
			zap.Error(err),
			zap.String("projectId", "p-9001"),
		)
	}

	// 警告示例:可降级但需要关注。
	logger.FromContext(ctx).Warn("slow dependency",
		zap.String("dep", "harbor"),
		zap.Int64("latencyMs", 1820),
	)
}

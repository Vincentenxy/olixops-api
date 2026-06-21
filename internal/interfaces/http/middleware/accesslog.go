package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"olixops/internal/platform/logger"
)

// AccessLog 返回 gin 中间件,记录每个 HTTP 请求的访问日志。
//
// 日志字段:
//   - method   请求方法
//   - path     请求路径 (不含 query string, 避免敏感参数落盘)
//   - status   HTTP 响应状态码
//   - costMs   处理耗时 (毫秒, 保留微秒精度)
//   - clientIp 客户端 IP
//   - bytes    响应字节数 (handler 未写 body 时为 0)
//
// 挂载要求: 必须在 Trace 中间件之后挂载, 这样 logger.FromContext 才能拿到
// 带 trace_id 的子 logger, 日志自动带上 trace-id。
//
// 使用 defer 确保 handler panic 时也能记录 status (gin 默认 panic 状态为 500)。
func AccessLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		defer func() {
			cost := time.Since(start)
			status := c.Writer.Status()
			bytes := c.Writer.Size()
			if bytes < 0 {
				bytes = 0
			}

			// 用 logger.FromContext 拿带 trace_id 的子 logger, 日志自动有 trace-id
			lg := logger.FromContext(c.Request.Context())
			lg.Info("http access",
				zap.String("method", method),
				zap.String("path", path),
				zap.Int("status", status),
				zap.Float64("costMs", float64(cost.Microseconds())/1000.0),
				zap.String("clientIp", c.ClientIP()),
				zap.Int("bytes", bytes),
			)
		}()

		c.Next()
	}
}

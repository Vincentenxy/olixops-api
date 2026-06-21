// Package middleware 提供 HTTP 中间件。
//
// trace 中间件负责:
//   - 读取前端传入的 X-Request-Id, 缺省时生成 UUID 作为 trace_id
//   - 把 trace_id 写入 gin context (httpx 从这里取)、request context (logger 从这里取)
//   - 把 trace_id 写入 response header X-Trace-Id, 让前端也能拿到同链路 ID
package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"olixops/internal/platform/logger"
)

const (
	// HeaderRequestID 前端传入的 trace ID header (业界主流命名)。
	HeaderRequestID = "X-Request-Id"
	// HeaderTraceID 后端响应里回写的 trace ID header。
	// 与 logger / 日志规范的 trace_id 字段对应。
	HeaderTraceID = "X-Trace-Id"

	// ginContextKey 与 pkg/httpx.response.go 里的 traceIDKey 保持一致,
	// 保证 httpx.OK/Error 等 envelope 能取到 trace_id。
	ginContextKey = "trace_id"
)

// Trace 返回 gin 中间件,实现 trace_id 注入。
//
// 调用顺序: 应当是 router 上第一个挂载的中间件,
// 让 accesslog / recover / auth 等后续中间件和 handler 都能拿到 trace_id。
func Trace() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1) 优先取前端传入的 X-Request-Id, 没有则生成 UUID。
		traceID := c.GetHeader(HeaderRequestID)
		if traceID == "" {
			traceID = uuid.NewString()
		}

		// 2) 写 gin context (httpx 从这里读)。
		c.Set(ginContextKey, traceID)

		// 3) 构造带 trace_id 字段的子 logger, 塞进 request context。
		//    handler 通过 logger.FromContext(ctx) 拿到的就是这个子 logger,
		//    每条日志自动带 trace_id 字段 (符合 logging-standards 规范)。
		l := logger.L().With(zap.String("trace_id", traceID))
		ctx := logger.WithContext(c.Request.Context(), l)
		c.Request = c.Request.WithContext(ctx)

		// 4) response header 回写, 让前端拿到同链路 ID。
		c.Header(HeaderTraceID, traceID)

		c.Next()
	}
}

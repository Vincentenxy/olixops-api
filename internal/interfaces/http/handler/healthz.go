package handler

import (
	"github.com/gin-gonic/gin"

	"olixops/internal/platform/logger"
	"olixops/pkg/httpx"
)

// Healthz GET /api/pub/v1/healthz
//
// 探针接口:打日志,返回 envelope。后续 trace-id 中间件接入后,
// logger.FromContext 取出的 logger 会自带 trace_id,日志自动满足规范。
func Healthz(c *gin.Context) {
	logger.FromContext(c.Request.Context()).Info("healthz hit")
	httpx.OK(c, gin.H{"status": "ok"})
}

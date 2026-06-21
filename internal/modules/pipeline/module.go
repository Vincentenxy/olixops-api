// Package pipeline 集成 Tekton 流水线。
//
// 待实现:
//   - Pipeline / Task 模板与参数化
//   - PipelineRun 触发、取消、重试、日志、状态
//   - 触发方式(手动 / Webhook / 定时 / 需求驱动)
//   - 凭据管理(Git / Registry / Kube / Secret)
//   - 结果回写(镜像 Tag / 测试报告 / 发布记录)
package pipeline

import "github.com/gin-gonic/gin"

// Handler 是 Pipeline 模块的 HTTP 入口占位。
type Handler struct{}

// NewHandler 构造 handler。
func NewHandler() *Handler { return &Handler{} }

// Register 注册路由。
func (h *Handler) Register(r *gin.RouterGroup) {
	tpl := r.Group("/pipelines")
	tpl.GET("", notImplemented("list pipeline templates"))
	tpl.POST("", notImplemented("create pipeline template"))

	runs := r.Group("/pipeline-runs")
	runs.GET("", notImplemented("list pipeline runs"))
	runs.POST("", notImplemented("trigger pipeline run"))
	runs.POST("/:id/cancel", notImplemented("cancel pipeline run"))
	runs.GET("/:id/logs", notImplemented("get pipeline run logs"))
}

func notImplemented(feature string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(501, gin.H{"code": "NOT_IMPLEMENTED", "message": feature + " not implemented yet"})
	}
}

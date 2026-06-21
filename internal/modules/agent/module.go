// Package agent 管理智能体应用、沙箱、Session 与工具调用审计。
//
// 待实现:
//   - Agent 应用注册(模型、工具、知识库、上下文策略)
//   - 沙箱模板与生命周期(隔离、超时、回收)
//   - Agent Session 与执行记录
//   - 工具调用、文件访问、外部网络访问审计
//   - 并发、配额、成本统计
package agent

import "github.com/gin-gonic/gin"

// Handler 是 Agent 模块的 HTTP 入口占位。
type Handler struct{}

// NewHandler 构造 handler。
func NewHandler() *Handler { return &Handler{} }

// Register 注册路由。
func (h *Handler) Register(r *gin.RouterGroup) {
	g := r.Group("/agents")
	g.GET("", notImplemented("list agent applications"))
	g.POST("", notImplemented("create agent application"))

	s := r.Group("/sandboxes")
	s.GET("", notImplemented("list sandboxes"))
	s.POST("", notImplemented("create sandbox"))
}

func notImplemented(feature string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(501, gin.H{"code": "NOT_IMPLEMENTED", "message": feature + " not implemented yet"})
	}
}

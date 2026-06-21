// Package release 管理需求、发布单、审批流与环境流转。
//
// 待实现:
//   - 需求 / 变更管理
//   - 发布单(发布范围、目标环境、审批人、风险等级)
//   - 环境流转 dev -> test -> sim -> prod
//   - 审批流程
//   - 变更记录与发布看板
package release

import "github.com/gin-gonic/gin"

// Handler 是 Release 模块的 HTTP 入口占位。
type Handler struct{}

// NewHandler 构造 handler。
func NewHandler() *Handler { return &Handler{} }

// Register 注册路由。
func (h *Handler) Register(r *gin.RouterGroup) {
	req := r.Group("/requirements")
	req.GET("", notImplemented("list requirements"))
	req.POST("", notImplemented("create requirement"))

	orders := r.Group("/release-orders")
	orders.GET("", notImplemented("list release orders"))
	orders.POST("", notImplemented("create release order"))
	orders.POST("/:id/approve", notImplemented("approve release order"))
	orders.POST("/:id/execute", notImplemented("execute release order"))
}

func notImplemented(feature string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(501, gin.H{"code": "NOT_IMPLEMENTED", "message": feature + " not implemented yet"})
	}
}

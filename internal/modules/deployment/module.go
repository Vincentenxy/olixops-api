// Package deployment 管理应用的部署、发布与回滚。
//
// 待实现:
//   - 部署单元(Deployment / StatefulSet / DaemonSet / Job / CronJob)
//   - 发布策略(滚动、蓝绿、金丝雀、暂停/继续)
//   - 部署记录、版本、回滚
//   - 运行状态、Pod 事件、日志入口
package deployment

import "github.com/gin-gonic/gin"

// Handler 是部署模块的 HTTP 入口占位。
type Handler struct{}

// NewHandler 构造 handler。
func NewHandler() *Handler { return &Handler{} }

// Register 注册路由。
func (h *Handler) Register(r *gin.RouterGroup) {
	g := r.Group("/applications/:app_id/deployments")
	g.GET("", notImplemented("list deployments"))
	g.POST("", notImplemented("create deployment"))
	g.POST("/:id/rollback", notImplemented("rollback deployment"))
}

func notImplemented(feature string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(501, gin.H{"code": "NOT_IMPLEMENTED", "message": feature + " not implemented yet"})
	}
}

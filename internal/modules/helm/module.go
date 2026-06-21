// Package helm 集成 Helm Chart 仓库与 Release 生命周期。
//
// 待实现:
//   - Chart 仓库接入(OCI / HTTP / 本地)
//   - Chart 元数据、Values 管理(按环境保存、差异对比)
//   - Release 安装、升级、回滚、卸载
//   - 模板渲染与发布前检查
//   - 依赖治理
package helm

import "github.com/gin-gonic/gin"

// Handler 是 Helm 模块的 HTTP 入口占位。
type Handler struct{}

// NewHandler 构造 handler。
func NewHandler() *Handler { return &Handler{} }

// Register 注册路由。
func (h *Handler) Register(r *gin.RouterGroup) {
	repos := r.Group("/helm/repositories")
	repos.GET("", notImplemented("list helm repositories"))
	repos.POST("", notImplemented("add helm repository"))

	releases := r.Group("/helm/releases")
	releases.GET("", notImplemented("list helm releases"))
	releases.POST("", notImplemented("install release"))
	releases.POST("/:id/rollback", notImplemented("rollback release"))
}

func notImplemented(feature string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(501, gin.H{"code": "NOT_IMPLEMENTED", "message": feature + " not implemented yet"})
	}
}

// Package registry 集成镜像仓库与制品仓库。
//
// 待实现:
//   - Registry 接入(Harbor / Docker Registry / 云厂商)
//   - 项目与仓库同步(项目 / 镜像 / Tag / Digest / 扫描结果)
//   - 镜像选择(按分支 / Commit / Tag)
//   - 安全扫描与阻断策略
//   - 镜像晋级与清理策略
package registry

import "github.com/gin-gonic/gin"

// Handler 是 Registry 模块的 HTTP 入口占位。
type Handler struct{}

// NewHandler 构造 handler。
func NewHandler() *Handler { return &Handler{} }

// Register 注册路由。
func (h *Handler) Register(r *gin.RouterGroup) {
	g := r.Group("/registries")
	g.GET("", notImplemented("list registries"))
	g.POST("", notImplemented("register registry"))
	g.GET("/:id/repositories", notImplemented("list repositories"))
	g.GET("/:id/repositories/:repo/tags", notImplemented("list tags"))
}

func notImplemented(feature string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(501, gin.H{"code": "NOT_IMPLEMENTED", "message": feature + " not implemented yet"})
	}
}

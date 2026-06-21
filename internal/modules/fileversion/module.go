// Package fileversion 管理配置文件、脚本、部署文件、需求附件等版本化资源。
//
// 待实现:
//   - 文件仓库(按项目 / 应用 / 环境 / 目录组织)
//   - 版本记录(版本号、提交人、说明、时间、Hash)
//   - 差异对比(文本 / 配置 / 二进制元数据)
//   - 发布绑定(发布单 / 流水线 / 部署记录)
//   - 存储后端(本地 / 对象存储 / Git)
//   - 权限与审计
package fileversion

import "github.com/gin-gonic/gin"

// Handler 是 FileVersion 模块的 HTTP 入口占位。
type Handler struct{}

// NewHandler 构造 handler。
func NewHandler() *Handler { return &Handler{} }

// Register 注册路由。
func (h *Handler) Register(r *gin.RouterGroup) {
	g := r.Group("/files")
	g.GET("", notImplemented("list file assets"))
	g.POST("", notImplemented("upload file asset"))
	g.GET("/:id/versions", notImplemented("list file versions"))
	g.GET("/:id/versions/:version", notImplemented("get file version"))
	g.GET("/:id/diff", notImplemented("diff file versions"))
}

func notImplemented(feature string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(501, gin.H{"code": "NOT_IMPLEMENTED", "message": feature + " not implemented yet"})
	}
}

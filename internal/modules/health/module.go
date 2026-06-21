// Package health 是基础设施探针模块,提供 /api/pub/v1/healthz 接口。
//
// 本模块没有业务逻辑,主要用于演示 ModuleRoutes 契约;
// 未来其他 12 个业务模块按相同结构 (Module struct + New() + Register*) 复制即可。
package health

import (
	"github.com/gin-gonic/gin"

	"olixops/internal/interfaces/http/handler"
	"olixops/internal/interfaces/http/router"
)

// Module 是 health 模块的入口,实现 router.ModuleRoutes 契约。
type Module struct{}

// New 构造 health 模块。
func New() *Module { return &Module{} }

// 编译期断言:确保 health 模块永远满足 router.ModuleRoutes 契约。
var _ router.ModuleRoutes = (*Module)(nil)

// Name 返回模块名,用于启动日志。
func (m *Module) Name() string { return "health" }

// RegisterPub 把 healthz 挂到 /api/pub/v1。
func (m *Module) RegisterPub(g *gin.RouterGroup) {
	g.GET("/healthz", handler.Healthz)
}

// RegisterPrivate 空实现 — health 是基础设施模块,无业务接口。
func (m *Module) RegisterPrivate(_ *gin.RouterGroup) {}

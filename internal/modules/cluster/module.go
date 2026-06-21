// Package cluster 是 cluster 模块对外的路由注册入口, 实现 router.ModuleRoutes。
//
// 路径设计 (按 [[api-path-layout]] / [[api-http-methods]] 规范):
//
//	/api/v1/clusters                                  GET/POST
//	/api/v1/clusters/:id                              GET
//	/api/v1/clusters/:id/update                       POST
//	/api/v1/clusters/:id/refresh                      POST
//	/api/v1/clusters/:id/delete                       POST
//	/api/v1/clusters/:id/namespaces                   GET/POST
//	/api/v1/clusters/:id/namespaces/:ns              GET
//	/api/v1/clusters/:id/namespaces/:ns/delete       POST
//	/api/v1/clusters/:id/nodes                        GET
//	/api/v1/clusters/:id/nodes/:name                  GET
//	/api/v1/clusters/:id/deployments                  GET
//	/api/v1/clusters/:id/deployments/:ns/:name        GET
//	/api/v1/clusters/:id/services                     GET
//	/api/v1/clusters/:id/services/:ns/:name          GET
//	/api/v1/clusters/:id/ingresses                    GET
//	/api/v1/clusters/:id/ingresses/:ns/:name         GET
//
// 注意: 严禁 PUT / DELETE, [[api-http-methods]] 规范禁止。
package cluster

import (
	"github.com/gin-gonic/gin"

	"olixops/internal/interfaces/http/router"
	"olixops/internal/modules/cluster/handler"
)

// Handlers 聚合 cluster 模块所有子 handler, 减少 Module 字段数。
// 后续 NewModule 接受这个聚合结构, 方便 app.go 一次性注入。
type Handlers struct {
	Cluster  *handler.ClusterHandler
	Ns       *handler.NamespaceHandler
	Node     *handler.NodeHandler
	Workload *handler.WorkloadHandler
}

// Module 是 cluster 模块对外的路由注册入口, 实现 router.ModuleRoutes。
type Module struct {
	h *Handlers
}

// NewModule 构造 cluster 模块。
func NewModule(h *Handlers) *Module {
	return &Module{h: h}
}

// 编译期断言: Module 必须满足 router.ModuleRoutes。
var _ router.ModuleRoutes = (*Module)(nil)

// Name 返回模块名, 用于启动日志。
func (m *Module) Name() string { return "cluster" }

// RegisterPub cluster 没有 pub 路由, 空实现。
func (m *Module) RegisterPub(_ *gin.RouterGroup) {}

// RegisterPrivate 挂载需鉴权路由到 /api/v1。
//
// TODO 实现完所有 handler 后, 把以下注释展开为真实路由注册:
//
//	cl := g.Group("/clusters")
//	cl.GET("", m.h.Cluster.list)
//	cl.POST("", m.h.Cluster.create)
//	cl.GET("/:id", m.h.Cluster.get)
//	cl.POST("/:id/update", m.h.Cluster.update)
//	cl.POST("/:id/refresh", m.h.Cluster.refresh)
//	cl.POST("/:id/delete", m.h.Cluster.delete)
//
//	cl.GET("/:id/namespaces", m.h.Ns.list)
//	cl.POST("/:id/namespaces", m.h.Ns.create)
//	cl.GET("/:id/namespaces/:ns", m.h.Ns.get)
//	cl.POST("/:id/namespaces/:ns/delete", m.h.Ns.delete)
//
//	cl.GET("/:id/nodes", m.h.Node.list)
//	cl.GET("/:id/nodes/:name", m.h.Node.get)
//
//	cl.GET("/:id/deployments", m.h.Workload.listDeployments)
//	cl.GET("/:id/deployments/:ns/:name", m.h.Workload.getDeployment)
//	cl.GET("/:id/services", m.h.Workload.listServices)
//	cl.GET("/:id/services/:ns/:name", m.h.Workload.getService)
//	cl.GET("/:id/ingresses", m.h.Workload.listIngresses)
//	cl.GET("/:id/ingresses/:ns/:name", m.h.Workload.getIngress)
func (m *Module) RegisterPrivate(g *gin.RouterGroup) {
	_ = g
}

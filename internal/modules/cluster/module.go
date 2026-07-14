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
	"context"
	"olixops/internal/config"
	"olixops/internal/interfaces/http/router"
	"olixops/internal/modules/cluster/adapter/k8s"
	"olixops/internal/modules/cluster/handler"
	"olixops/internal/modules/cluster/repository"
	"olixops/internal/modules/cluster/service"
	"olixops/internal/platform/audit"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Handlers 聚合 cluster 模块所有子 handler, 减少 Module 字段数。
// 后续 NewModule 接受这个聚合结构, 方便 app.go 一次性注入。
type Handlers struct {
	ClusterHandler  *handler.ClusterHandler
	NsHandler       *handler.NamespaceHandler
	NodeHandler     *handler.NodeHandler
	WorkloadHandler *handler.WorkloadHandler
}

func NewHandlers(clusterHandler *handler.ClusterHandler,
	nsHandler *handler.NamespaceHandler,
	nodeHandler *handler.NodeHandler,
	workloadHandler *handler.WorkloadHandler) *Handlers {
	return &Handlers{
		ClusterHandler:  clusterHandler,
		NsHandler:       nsHandler,
		NodeHandler:     nodeHandler,
		WorkloadHandler: workloadHandler,
	}
}

// Module 是 cluster 模块对外的路由注册入口, 实现 router.ModuleRoutes。
type Module struct {
	h *Handlers
}

// NewModule 构造 cluster 模块
func NewModule(h *Handlers) *Module {
	return &Module{h: h}
}

func Assemble(ctx context.Context, db *gorm.DB, k8sConfig *config.K8sConfig, recoder audit.Recorder) (*Module, error) {

	repo := repository.NewClusterRepo(db)
	factory := k8s.NewFactory(ctx, k8sConfig, repo)

	// cluster handler init
	svc := service.NewClusterService(repo, factory)
	clusterHandler := handler.NewClusterHandler(svc, k8sConfig)

	// ns handler init
	nssvc := service.NewNamespaceService(repo, factory)
	nsHandler := handler.NewNamespaceHandler(nssvc)

	// node handler
	nodesvc := service.NewNodeService(repo, factory)
	nodeHandler := handler.NewNodeHandler(nodesvc)

	// workload handler init
	// TODO

	h := NewHandlers(
		clusterHandler,
		nsHandler,
		nodeHandler,
		nil,
	)
	return NewModule(h), nil
}

// Name 返回模块名, 用于启动日志
func (m *Module) Name() string { return "cluster" }

// RegisterPub cluster 没有 pub 路由, 空实现。
func (m *Module) RegisterPub(_ *gin.RouterGroup) {}

func (m *Module) RegisterPrivate(g *gin.RouterGroup) {
	v1 := g.Group("/v1")
	{
		k8s := v1.Group("/k8s")
		{
			cl := k8s.Group("/cluster")
			{
				cl.POST("/create", m.h.ClusterHandler.Create)
				cl.POST("/list", m.h.ClusterHandler.List)
			}

			node := k8s.Group("/node")
			{
				node.POST("/list", m.h.NodeHandler.List)
			}

			ns := k8s.Group("/ns")
			{
				ns.POST("/list", m.h.NsHandler.List)
			}

		}
	}
}

// 编译期断言: Module 必须满足 router.ModuleRoutes
var _ router.ModuleRoutes = (*Module)(nil)

// Package helm 集成 Helm Chart 仓库与 Release 生命周期。
//
// 实现范围:
//   - Chart 仓库接入 (HTTP / OCI)
//   - Chart 元数据查询 (列表、版本、默认 values)
//   - Release 安装、升级、回滚、卸载
//   - Release 版本历史、资源查询
//
// 模块分层 (按 modular monolith 规范):
//
//	domain/       → 实体 + 领域错误 (ChartRepo, Release, ChartMeta, ChartVersion)
//	repository/   → GORM 持久化 (HelmRepo 接口)
//	adapter/      → Helm SDK 封装 (HelmClient + Factory 接口)
//	service/      → 业务编排 (RepoService, ReleaseService)
//	handler/      → HTTP 入口 (RepoHandler, ReleaseHandler)
//
// 路由设计 (按 api-path-layout 规范):
//
//	/api/v1/helm/repo/add                                  POST
//	/api/v1/helm/repo/list                                 POST
//	/api/v1/helm/repo/:id/detail                           POST
//	/api/v1/helm/repo/:id/update                           POST
//	/api/v1/helm/repo/:id/delete                           POST
//	/api/v1/helm/repo/:id/sync                             POST
//	/api/v1/helm/repo/:id/charts                           POST
//	/api/v1/helm/repo/:id/charts/:chart_name/versions      POST
//	/api/v1/helm/repo/:id/charts/:chart_name/values        POST
//
//	/api/v1/helm/release/install                           POST
//	/api/v1/helm/release/list                              POST
//	/api/v1/helm/release/:id/detail                        POST
//	/api/v1/helm/release/:id/upgrade                       POST
//	/api/v1/helm/release/:id/rollback                      POST
//	/api/v1/helm/release/:id/uninstall                     POST
//	/api/v1/helm/release/:id/history                       POST
//	/api/v1/helm/release/:id/resources                     POST
//
// 注意: 严禁 PUT / DELETE, api-http-methods 规范禁止。
package helm

import (
	"olixops/internal/interfaces/http/router"
	"olixops/internal/modules/helm/adapter"
	"olixops/internal/modules/helm/handler"
	"olixops/internal/modules/helm/repository"
	"olixops/internal/modules/helm/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Handlers 聚合 helm 模块所有子 handler, 减少 Module 字段数。
type Handlers struct {
	RepoHandler    *handler.RepoHandler
	ReleaseHandler *handler.ReleaseHandler
}

// NewHandlers 构造 Handlers。
func NewHandlers(
	repoHandler *handler.RepoHandler,
	releaseHandler *handler.ReleaseHandler,
) *Handlers {
	return &Handlers{
		RepoHandler:    repoHandler,
		ReleaseHandler: releaseHandler,
	}
}

// Module 是 helm 模块对外的路由注册入口, 实现 router.ModuleRoutes。
type Module struct {
	h *Handlers
}

// NewModule 构造 helm 模块。
func NewModule(h *Handlers) *Module {
	return &Module{h: h}
}

// ─────────────────────────────────────────────────────────────────────
// Assemble 是 helm 模块的依赖装配入口。
//
// 装配顺序: repository → adapter.Factory → service → handler → Module。
//
// 参数说明:
//   - db:               GORM DB 实例, 用于 repository 持久化
//   - helmFactory:      Helm SDK Factory, 封装 Helm action (install/upgrade/...)
//   - clusterProvider:  从 cluster 模块获取 kubeconfig 的接口 (跨模块解耦)
//
// TODO: clusterProvider 的实现需要在 cluster 模块中提供,
//
//	可以在 app.go 的 Run() 中构造后传入。
//	初期可以先传 nil, 然后只实现 Repo 相关功能 (不依赖 kubeconfig)。
//
// ─────────────────────────────────────────────────────────────────────
func Assemble(
	db *gorm.DB,
	helmFactory adapter.Factory,
	clusterProvider service.ClusterKubeconfigProvider,
) (*Module, error) {

	// 1. repository
	repo := repository.NewHelmRepo(db)

	// 2. repo service + handler
	repoSvc := service.NewRepoService(repo, helmFactory)
	repoHandler := handler.NewRepoHandler(repoSvc)

	// 3. release service + handler
	releaseSvc := service.NewReleaseService(repo, helmFactory, clusterProvider)
	releaseHandler := handler.NewReleaseHandler(releaseSvc)

	// 4. 聚合
	h := NewHandlers(repoHandler, releaseHandler)
	return NewModule(h), nil
}

// ─────────────────────────────────────────────────────────────────────
// ModuleRoutes 接口实现
// ─────────────────────────────────────────────────────────────────────

// Name 返回模块名, 用于启动日志。
func (m *Module) Name() string { return "helm" }

// RegisterPub helm 模块没有 pub 路由, 空实现。
func (m *Module) RegisterPub(_ *gin.RouterGroup) {}

// RegisterPrivate 注册 helm 模块的鉴权路由。
//
// 路由分两组:
//   - /v1/helm/repo/*    仓库管理
//   - /v1/helm/release/* Release 管理
func (m *Module) RegisterPrivate(g *gin.RouterGroup) {
	v1 := g.Group("/v1")
	{
		helmGroup := v1.Group("/helm")
		{
			// ─────────────────────────────────────────────
			// Chart 仓库管理
			// ─────────────────────────────────────────────
			repoGroup := helmGroup.Group("/repo")
			{
				repoGroup.POST("/add", m.h.RepoHandler.Add)
				repoGroup.POST("/list", m.h.RepoHandler.List)
				repoGroup.POST("/:id/detail", m.h.RepoHandler.Get)
				repoGroup.POST("/:id/update", m.h.RepoHandler.Update)
				repoGroup.POST("/:id/delete", m.h.RepoHandler.Delete)
				repoGroup.POST("/:id/sync", m.h.RepoHandler.Sync)

				// Chart 查询 (基于仓库)
				repoGroup.POST("/:id/charts", m.h.RepoHandler.ListCharts)
				repoGroup.POST("/:id/charts/:chart_name/versions", m.h.RepoHandler.ListChartVersions)
				repoGroup.POST("/:id/charts/:chart_name/values", m.h.RepoHandler.GetChartDefaultValues)
			}

			// ─────────────────────────────────────────────
			// Release 生命周期管理
			// ─────────────────────────────────────────────
			releaseGroup := helmGroup.Group("/release")
			{
				releaseGroup.POST("/install", m.h.ReleaseHandler.Install)
				releaseGroup.POST("/list", m.h.ReleaseHandler.List)
				releaseGroup.POST("/:id/detail", m.h.ReleaseHandler.Get)
				releaseGroup.POST("/:id/upgrade", m.h.ReleaseHandler.Upgrade)
				releaseGroup.POST("/:id/rollback", m.h.ReleaseHandler.Rollback)
				releaseGroup.POST("/:id/uninstall", m.h.ReleaseHandler.Uninstall)
				releaseGroup.POST("/:id/history", m.h.ReleaseHandler.History)
				releaseGroup.POST("/:id/resources", m.h.ReleaseHandler.Resources)
			}
		}
	}
}

// 编译期断言: Module 必须满足 router.ModuleRoutes。
var _ router.ModuleRoutes = (*Module)(nil)

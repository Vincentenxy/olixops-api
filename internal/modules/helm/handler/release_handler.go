package handler

import (
	"olixops/internal/modules/helm/service"
	"olixops/pkg/errs"
	"olixops/pkg/httpx"

	"github.com/gin-gonic/gin"
)

// ReleaseHandler Helm Release HTTP 入口。
type ReleaseHandler struct {
	svc *service.ReleaseService
}

// NewReleaseHandler 构造 handler。
func NewReleaseHandler(svc *service.ReleaseService) *ReleaseHandler {
	return &ReleaseHandler{svc: svc}
}

// ─────────────────────────────────────────────────────────────────────
// 请求结构体
// ─────────────────────────────────────────────────────────────────────

// InstallReq 安装 Release 请求体。
type InstallReq struct {
	Name      string         `json:"name" binding:"required,min=2,max=128"`
	Namespace string         `json:"namespace" binding:"required,min=2,max=128"`
	ClusterID string         `json:"cluster_id" binding:"required,uuid"`
	AppID     string         `json:"app_id" binding:"omitempty,uuid"`
	EnvID     string         `json:"env_id" binding:"omitempty,uuid"`
	RepoID    string         `json:"repo_id" binding:"required,uuid"`
	ChartName string         `json:"chart_name" binding:"required,min=2,max=256"`
	ChartVer  string         `json:"chart_ver" binding:"required,min=1,max=64"`
	Values    map[string]any `json:"values"`
	Wait      bool           `json:"wait"`
	Timeout   int            `json:"timeout"`
}

// UpgradeReq 升级 Release 请求体。
type UpgradeReq struct {
	ChartVer string         `json:"chart_ver" binding:"required,min=1,max=64"`
	Values   map[string]any `json:"values"`
	Wait     bool           `json:"wait"`
	Timeout  int            `json:"timeout"`
}

// RollbackReq 回滚 Release 请求体。
type RollbackReq struct {
	Version int  `json:"version"` // 0 = 回滚到上一个
	Wait    bool `json:"wait"`
	Timeout int  `json:"timeout"`
}

// UninstallReq 卸载 Release 请求体。
type UninstallReq struct {
	KeepHistory bool `json:"keep_history"`
}

// ListReleaseReq 列表查询 Release 请求体。
type ListReleaseReq struct {
	ClusterID string `json:"cluster_id"`
	Namespace string `json:"namespace"`
	AppID     string `json:"app_id"`
	Status    string `json:"status"`
	ChartName string `json:"chart_name"`
}

// ─────────────────────────────────────────────────────────────────────
// 路由处理方法
// ─────────────────────────────────────────────────────────────────────

// Install POST /api/v1/helm/release/install
//
// 安装一个新的 Helm Release。
//
// 实现步骤:
//   1. ShouldBindJSON(&req) 解析请求
//   2. 构造 service.InstallInput
//   3. svc.Install(ctx, input)
//   4. httpx.OK(c, release) 返回
func (h *ReleaseHandler) Install(c *gin.Context) {
	var req InstallReq
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Fail(c, errs.InvalidArg("invalid request: %v", err))
		return
	}

	// TODO: 实现
	// release, err := h.svc.Install(c.Request.Context(), service.InstallInput{
	//     Name:      req.Name,
	//     Namespace: req.Namespace,
	//     ClusterID: req.ClusterID,
	//     AppID:     req.AppID,
	//     EnvID:     req.EnvID,
	//     RepoID:    req.RepoID,
	//     ChartName: req.ChartName,
	//     ChartVer:  req.ChartVer,
	//     Values:    req.Values,
	//     Wait:      req.Wait,
	//     Timeout:   req.Timeout,
	// })
	// if err != nil {
	//     httpx.Fail(c, err)
	//     return
	// }
	// httpx.OK(c, release)

	_ = req
	httpx.Fail(c, errs.InvalidArg("not implemented"))
}

// Upgrade POST /api/v1/helm/release/:id/upgrade
//
// 升级已安装的 Release (Chart 版本或 values 变更)。
//
// 实现步骤:
//   1. c.Param("id") 获取 Release ID
//   2. ShouldBindJSON(&req)
//   3. svc.Upgrade(ctx, id, UpgradeInput{...})
//   4. httpx.OK(c, release)
func (h *ReleaseHandler) Upgrade(c *gin.Context) {
	// TODO: 实现
	// id := c.Param("id")
	// var req UpgradeReq
	// if err := c.ShouldBindJSON(&req); err != nil { ... }
	// release, err := h.svc.Upgrade(ctx, id, service.UpgradeInput{...})
	// httpx.OK(c, release)

	_ = c
	httpx.Fail(c, errs.InvalidArg("not implemented"))
}

// Rollback POST /api/v1/helm/release/:id/rollback
//
// 回滚到指定 revision。
//
// 实现步骤:
//   1. c.Param("id") 获取 Release ID
//   2. ShouldBindJSON(&req)
//   3. svc.Rollback(ctx, id, RollbackInput{...})
//   4. httpx.OK(c, release)
func (h *ReleaseHandler) Rollback(c *gin.Context) {
	// TODO: 实现
	_ = c
	httpx.Fail(c, errs.InvalidArg("not implemented"))
}

// Uninstall POST /api/v1/helm/release/:id/uninstall
//
// 卸载一个 Release (删除 K8s 资源)。
//
// 实现步骤:
//   1. c.Param("id") 获取 Release ID
//   2. ShouldBindJSON(&req) (可选, keep_history)
//   3. svc.Uninstall(ctx, id, req.KeepHistory)
//   4. httpx.OK[any](c, nil)
func (h *ReleaseHandler) Uninstall(c *gin.Context) {
	// TODO: 实现
	_ = c
	httpx.Fail(c, errs.InvalidArg("not implemented"))
}

// Get POST /api/v1/helm/release/:id/detail
//
// 查询单个 Release 详情。
//
// 实现步骤:
//   1. c.Param("id")
//   2. svc.Get(ctx, id)
//   3. httpx.OK(c, release)
func (h *ReleaseHandler) Get(c *gin.Context) {
	// TODO: 实现
	_ = c
	httpx.Fail(c, errs.InvalidArg("not implemented"))
}

// List POST /api/v1/helm/release/list
//
// 列表查询 Release (按集群/namespace/应用/状态过滤)。
//
// 实现步骤:
//   1. ShouldBindJSON(&req)
//   2. 构造 domain.ReleaseListFilter
//   3. svc.List(ctx, filter)
//   4. httpx.OK(c, list)
func (h *ReleaseHandler) List(c *gin.Context) {
	// TODO: 实现
	_ = c
	httpx.Fail(c, errs.InvalidArg("not implemented"))
}

// History POST /api/v1/helm/release/:id/history
//
// 获取 Release 的版本历史 (所有 revision, 从 Helm storage 拉取)。
//
// 实现步骤:
//   1. c.Param("id")
//   2. svc.History(ctx, id)
//   3. httpx.OK(c, []*adapter.ReleaseInfo)
func (h *ReleaseHandler) History(c *gin.Context) {
	// TODO: 实现
	_ = c
	httpx.Fail(c, errs.InvalidArg("not implemented"))
}

// Resources POST /api/v1/helm/release/:id/resources
//
// 获取 Release 部署的 K8s 资源列表 (Deployment/Service/Ingress 等)。
//
// 实现步骤:
//   1. c.Param("id")
//   2. svc.GetResources(ctx, id)
//   3. httpx.OK(c, []adapter.ResourceInfo)
func (h *ReleaseHandler) Resources(c *gin.Context) {
	// TODO: 实现
	_ = c
	httpx.Fail(c, errs.InvalidArg("not implemented"))
}

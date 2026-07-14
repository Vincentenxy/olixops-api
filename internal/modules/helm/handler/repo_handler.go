package handler

import (
	"olixops/internal/modules/helm/domain"
	"olixops/internal/modules/helm/service"
	"olixops/internal/platform/logger"
	"olixops/pkg/errs"
	"olixops/pkg/httpx"
	"olixops/pkg/pagination"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RepoHandler Chart 仓库 HTTP 入口。
type RepoHandler struct {
	svc *service.RepoService
}

// NewRepoHandler 构造 handler。
func NewRepoHandler(svc *service.RepoService) *RepoHandler {
	return &RepoHandler{svc: svc}
}

// ─────────────────────────────────────────────────────────────────────
// 请求/响应结构体
// ─────────────────────────────────────────────────────────────────────

// AddRepoReq 添加仓库请求体。
type AddRepoReq struct {
	Name        string `json:"name" binding:"required,min=2,max=128"`
	URL         string `json:"url" binding:"required,url"`
	Type        string `json:"type" binding:"required,oneof=http oci"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	Description string `json:"description"`
}

// UpdateRepoReq 更新仓库请求体。
type UpdateRepoReq struct {
	URL         *string `json:"url,omitempty" binding:"omitempty,url"`
	Username    *string `json:"username,omitempty"`
	Password    *string `json:"password,omitempty"`
	Description *string `json:"description,omitempty"`
}

// ListRepoReq 列表查询仓库请求体。
type ListRepoReq struct {
	pagination.Query
	Name   string `json:"name"`
	Type   string `json:"type"`
	Status string `json:"status"`
}

// ─────────────────────────────────────────────────────────────────────
// 路由处理方法
// ─────────────────────────────────────────────────────────────────────

// Add POST /api/v1/helm/repo/add
//
// 添加一个新的 Chart 仓库。
//
// 实现步骤:
//  1. ShouldBindJSON(&req) 解析请求
//  2. 调用 svc.Add(ctx, AddRepoInput{...})
//  3. httpx.OK(c, chartRepo) 返回
func (h *RepoHandler) Add(c *gin.Context) {
	var req AddRepoReq
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Fail(c, errs.InvalidArg("invalid request: %v", err))
		return
	}

	chartRepo, err := h.svc.Add(c.Request.Context(), service.AddRepoInput{
		Name:        req.Name,
		URL:         req.URL,
		Type:        domain.RepoType(req.Type),
		Username:    req.Username,
		Password:    req.Password,
		Description: req.Description,
	})
	if err != nil {
		logger.L().Error("add repo error", zap.Error(err))
		httpx.Fail(c, err)
		return
	}
	httpx.OK(c, chartRepo)
}

// Sync POST /api/v1/helm/repo/:id/sync
//
// 手动触发仓库索引同步 (拉取 index.yaml, 更新 Chart 列表)。
//
// 实现步骤:
//  1. 从 URL 参数获取 id: c.Param("id")
//  2. 调用 svc.Sync(ctx, id)
//  3. httpx.OK(c, nil)
func (h *RepoHandler) Sync(c *gin.Context) {
	// TODO: 实现
	// id := c.Param("id")
	// if err := h.svc.Sync(c.Request.Context(), id); err != nil {
	//     httpx.Fail(c, err)
	//     return
	// }
	// httpx.OK[any](c, nil)
	_ = c
	httpx.Fail(c, errs.InvalidArg("not implemented"))
}

// Get GET /api/v1/helm/repo/:id
//
// 查询单个仓库详情。
//
// 实现步骤:
//  1. c.Param("id") 获取 ID
//  2. svc.Get(ctx, id)
//  3. httpx.OK(c, chartRepo)
func (h *RepoHandler) Get(c *gin.Context) {
	// TODO: 实现
	_ = c
	httpx.Fail(c, errs.InvalidArg("not implemented"))
}

// List POST /api/v1/helm/repo/list
//
// 列表查询仓库 (分页 + 过滤)。
//
// 实现步骤:
//  1. ShouldBindJSON(&req)
//  2. req.Normalize()
//  3. 构造 domain.RepoListFilter
//  4. svc.List(ctx, filter)
//  5. httpx.OK(c, pagination.PageResult{...})
func (h *RepoHandler) List(c *gin.Context) {
	// TODO: 实现
	var req ListRepoReq
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Fail(c, errs.InvalidArg("invalid request: %v", err))
		return
	}
	// req.Normalize()
	//
	// filter := &domain.RepoListFilter{
	//     Query:  pagination.Query{PageNum: req.PageNum, PageSize: req.PageSize},
	//     Name:   req.Name,
	//     Type:   domain.RepoType(req.Type),
	//     Status: domain.RepoStatus(req.Status),
	// }
	// list, total, err := h.svc.List(c.Request.Context(), filter)
	// if err != nil { httpx.Fail(c, err); return }
	// httpx.OK(c, pagination.PageResult[*domain.ChartRepo]{Total: total, List: list})

	_ = req
	httpx.Fail(c, errs.InvalidArg("not implemented"))
}

// Update POST /api/v1/helm/repo/:id/update
//
// 更新仓库信息 (URL / 认证 / 描述)。
//
// 实现步骤:
//  1. c.Param("id") + ShouldBindJSON(&req)
//  2. svc.Update(ctx, id, UpdateRepoInput{...})
//  3. httpx.OK(c, chartRepo)
func (h *RepoHandler) Update(c *gin.Context) {
	// TODO: 实现
	_ = c
	httpx.Fail(c, errs.InvalidArg("not implemented"))
}

// Delete POST /api/v1/helm/repo/:id/delete
//
// 删除仓库 (软删除)。
//
// 实现步骤:
//  1. c.Param("id")
//  2. svc.Delete(ctx, id)
//  3. httpx.OK[any](c, nil)
func (h *RepoHandler) Delete(c *gin.Context) {
	// TODO: 实现
	_ = c
	httpx.Fail(c, errs.InvalidArg("not implemented"))
}

// ListCharts POST /api/v1/helm/repo/:id/charts
//
// 列出仓库中的所有 Chart (实时拉取 index.yaml)。
//
// 实现步骤:
//  1. c.Param("id") 获取仓库 ID
//  2. svc.ListCharts(ctx, repoID)
//  3. httpx.OK(c, []*domain.ChartMeta)
func (h *RepoHandler) ListCharts(c *gin.Context) {
	// TODO: 实现
	_ = c
	httpx.Fail(c, errs.InvalidArg("not implemented"))
}

// ListChartVersions POST /api/v1/helm/repo/:id/charts/:chart_name/versions
//
// 列出某个 Chart 的所有可用版本。
//
// 实现步骤:
//  1. c.Param("id") + c.Param("chart_name")
//  2. svc.ListChartVersions(ctx, repoID, chartName)
//  3. httpx.OK(c, []*domain.ChartVersion)
func (h *RepoHandler) ListChartVersions(c *gin.Context) {
	// TODO: 实现
	_ = c
	httpx.Fail(c, errs.InvalidArg("not implemented"))
}

// GetChartDefaultValues POST /api/v1/helm/repo/:id/charts/:chart_name/values
//
// 获取某个 Chart 的默认 values.yaml (安装前预览可配置参数)。
//
// 实现步骤:
//  1. c.Param("id") + c.Param("chart_name") + query chart_ver
//  2. svc.GetChartDefaultValues(ctx, repoID, chartName, chartVer)
//  3. httpx.OK(c, map[string]any)
func (h *RepoHandler) GetChartDefaultValues(c *gin.Context) {
	// TODO: 实现
	_ = c
	httpx.Fail(c, errs.InvalidArg("not implemented"))
}

// 编译期防止 import 未用。
var _ = pagination.FromGin

package handler

import (
	"errors"
	"olixops/internal/config"
	"olixops/internal/modules/cluster/domain"
	"olixops/internal/modules/cluster/service"
	"olixops/pkg/cryptox"
	"olixops/pkg/errs"
	"olixops/pkg/httpx"
	"olixops/pkg/pagination"

	"github.com/gin-gonic/gin"
)

// ClusterHandler 集群 HTTP 入口。
type ClusterHandler struct {
	svc       *service.ClusterService
	k8sConfig *config.K8sConfig
}

// NewClusterHandler 构造 handler
func NewClusterHandler(svc *service.ClusterService, k8sConfig *config.K8sConfig) *ClusterHandler {
	return &ClusterHandler{
		svc:       svc,
		k8sConfig: k8sConfig,
	}
}

type ClusterCreateReq struct {
	ID          string `json:"id" binding:"required"`
	TenantID    string `json:"tenantId" binding:"required"`
	Env         string `json:"env" binding:"required"`
	Description string `json:"description" binding:"required"`
	KubeConfig  string `json:"kubeConfig" binding:"required"`
}

// Create POST /api/v1/k8s/cluster/create
func (h *ClusterHandler) Create(c *gin.Context) {
	var req ClusterCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Fail(c, errs.InvalidArg("invalid request: %v", err))
		return
	}

	aesgcm, err := cryptox.EncryptAESGCM([]byte(h.k8sConfig.SecretKey), []byte(req.KubeConfig))
	if err != nil {
		httpx.Fail(c, errors.New("save k8s info error!"))
		return
	}

	cluster, err := h.svc.Create(c.Request.Context(), service.CreateInput{
		ID:          req.ID,
		TenantID:    req.TenantID,
		Environment: req.Env,
		Description: req.Description,
		Kubeconfig:  aesgcm,
	})
	if err != nil {
		httpx.Fail(c, err)
	}
	httpx.OK[*domain.Cluster](c, cluster)
	return
}

type ClusterListReq struct {
	pagination.Query
	TenantID string `form:"tenantId" `
	Env      string `form:"env" `
	Status   string `form:"status" `
}

// List POST /api/v1/clusters
func (h *ClusterHandler) List(c *gin.Context) {
	var req ClusterListReq
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Fail(c, errs.InvalidArg("invalid request: %v", err))
	}

	req.Normalize()

	filter := &domain.ClusterListFilter{
		Query: pagination.Query{
			PageNum:  req.PageNum,
			PageSize: req.PageSize,
		},
		TenantID: req.TenantID,
		Env:      req.Env,
		Status:   req.Status,
	}
	list, total, err := h.svc.List(c.Request.Context(), filter)
	if err != nil {
		httpx.Fail(c, err)
		return
	}

	httpx.OK(c, pagination.PageResult[*domain.Cluster]{
		Total: total,
		List:  list,
	})
}

// Get GET /api/v1/clusters/:id
func (h *ClusterHandler) Get(c *gin.Context) {
	// TODO: h.svc.Get(ctx, c.Param("id")) → httpx.OK
	_ = c
}

// update POST /api/v1/clusters/:id/update
func (h *ClusterHandler) update(c *gin.Context) {
	// TODO
	_ = c
}

// refresh POST /api/v1/clusters/:id/refresh
func (h *ClusterHandler) refresh(c *gin.Context) {
	// TODO
	_ = c
}

// delete POST /api/v1/clusters/:id/delete
func (h *ClusterHandler) delete(c *gin.Context) {
	// TODO
	_ = c
}

// 编译期防止 pagination import 未用
var _ = pagination.FromGin

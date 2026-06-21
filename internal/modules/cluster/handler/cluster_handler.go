package handler

import (
	"modules/cluster/service"
	"olixops/pkg/httpx"
	"olixops/pkg/pagination"

	"github.com/gin-gonic/gin"
)

// ClusterHandler 集群 HTTP 入口。
type ClusterHandler struct {
	svc *service.ClusterService
}

// NewClusterHandler 构造 handler
func NewClusterHandler(svc *service.ClusterService) *ClusterHandler {
	return &ClusterHandler{
		svc: svc,
	}
}

type ClusterCreateReq struct {
	ID             string `json:"id" binding:"required"`
	TenantID       string `json:"tenantId" binding:"required"`
	Env            string `json:"env" binding:"required"`
	Description    string `json:"description" binding:"required"`
	KubeConfigPath string `json:"kubeConfigPath" binding:"required"`
}

// create POST /api/v1/clusters
func (h *ClusterHandler) create(c *gin.Context) {
	var req ClusterCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Fail()
	}
	h.svc
}

// list GET /api/v1/clusters
func (h *ClusterHandler) list(c *gin.Context) {
	// TODO:
	//   1. q := pagination.FromGin(c)
	//   2. filter := repository.ListFilter{
	//        Status: domain.ClusterStatus(c.Query("status")),
	//        Environment: c.Query("environment"),
	//      }
	//   3. items, total, err := h.svc.List(c.Request.Context(), q, filter)
	//   4. httpx.Paged(c, items, total, q.Page, q.PageSize)
	_ = c
}

// get GET /api/v1/clusters/:id
func (h *ClusterHandler) get(c *gin.Context) {
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
var _ = httpx.OK

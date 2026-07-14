package handler

import (
	"olixops/internal/platform/logger"
	"olixops/pkg/errs"
	"olixops/pkg/httpx"
	"time"

	"olixops/internal/modules/cluster/service"

	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NamespaceHandler namespace HTTP 入口。
type NamespaceHandler struct {
	svc *service.NamespaceService
}

// NewNamespaceHandler 构造 handler。
func NewNamespaceHandler(svc *service.NamespaceService) *NamespaceHandler {
	return &NamespaceHandler{
		svc: svc,
	}
}

type NsListRequest struct {
	ClusterID     string `json:"clusterId" binding:"required"`
	LabelSelector string `json:"labelSelector"`
	FieldSelector string `json:"fieldSelector"`
}

// List POST /api/v1/k8s/ns/list
func (h *NamespaceHandler) List(c *gin.Context) {
	var req NsListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.L().Error(err.Error())
		httpx.Fail(c, errs.InvalidArg("invalid params"))
		return
	}

	list, err := h.svc.List(c.Request.Context(), service.NsListParams{
		ListOptions: metav1.ListOptions{
			LabelSelector: req.LabelSelector,
			FieldSelector: req.FieldSelector,
		},
		ClusterID: req.ClusterID,
	})
	if err != nil {
		logger.L().Error(err.Error())
		httpx.Fail(c, err)
		return
	}

	httpx.OK(c, list)
}

func (h *NamespaceHandler) Create(c *gin.Context) {
	// TODO: bind CreateNamespaceInput, h.svc.Create, httpx.Created
	_ = c
}

// get GET /api/v1/clusters/:id/namespaces/:ns
func (h *NamespaceHandler) Get(c *gin.Context) {
	// TODO: h.svc.Get(ctx, clusterID, name) → httpx.OK
	_ = c
}

// delete POST /api/v1/clusters/:id/namespaces/:ns/delete
func (h *NamespaceHandler) Delete(c *gin.Context) {
	// TODO: h.svc.Delete(ctx, clusterID, name, 30*time.Second) → httpx.NoContent
	_ = c
}

var _ = time.Second

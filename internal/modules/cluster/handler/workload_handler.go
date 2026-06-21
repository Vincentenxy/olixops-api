package handler

import (
	"github.com/gin-gonic/gin"

	"olixops/internal/modules/cluster/service"
	"olixops/pkg/httpx"
)

// WorkloadHandler deployment / service / ingress HTTP 入口 (只读)。
type WorkloadHandler struct {
	svc *service.WorkloadService
}

// NewWorkloadHandler 构造 handler。
func NewWorkloadHandler(svc *service.WorkloadService) *WorkloadHandler {
	return &WorkloadHandler{svc: svc}
}

// deployments
func (h *WorkloadHandler) listDeployments(c *gin.Context) {
	// TODO: h.svc.ListDeployments(ctx, c.Param("id"), c.Query("namespace")) → httpx.OK
	_ = c
}

func (h *WorkloadHandler) getDeployment(c *gin.Context) {
	// TODO
	_ = c
}

// services
func (h *WorkloadHandler) listServices(c *gin.Context) {
	// TODO
	_ = c
}

func (h *WorkloadHandler) getService(c *gin.Context) {
	// TODO
	_ = c
}

// ingresses
func (h *WorkloadHandler) listIngresses(c *gin.Context) {
	// TODO
	_ = c
}

func (h *WorkloadHandler) getIngress(c *gin.Context) {
	// TODO
	_ = c
}

var _ = httpx.OK

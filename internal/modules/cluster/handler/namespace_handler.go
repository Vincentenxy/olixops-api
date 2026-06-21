package handler

import (
	"time"

	"github.com/gin-gonic/gin"

	"olixops/internal/modules/cluster/service"
)

// NamespaceHandler namespace HTTP 入口。
type NamespaceHandler struct {
	svc *service.NamespaceService
}

// NewNamespaceHandler 构造 handler。
func NewNamespaceHandler(svc *service.NamespaceService) *NamespaceHandler {
	return &NamespaceHandler{svc: svc}
}

// list GET /api/v1/clusters/:id/namespaces
func (h *NamespaceHandler) list(c *gin.Context) {
	// TODO: h.svc.List(ctx, c.Param("id")) → httpx.OK
	_ = c
}

// create POST /api/v1/clusters/:id/namespaces
func (h *NamespaceHandler) create(c *gin.Context) {
	// TODO: bind CreateNamespaceInput, h.svc.Create, httpx.Created
	_ = c
}

// get GET /api/v1/clusters/:id/namespaces/:ns
func (h *NamespaceHandler) get(c *gin.Context) {
	// TODO: h.svc.Get(ctx, clusterID, name) → httpx.OK
	_ = c
}

// delete POST /api/v1/clusters/:id/namespaces/:ns/delete
func (h *NamespaceHandler) delete(c *gin.Context) {
	// TODO: h.svc.Delete(ctx, clusterID, name, 30*time.Second) → httpx.NoContent
	_ = c
}

var _ = time.Second

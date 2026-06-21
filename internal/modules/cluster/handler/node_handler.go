package handler

import (
	"github.com/gin-gonic/gin"

	"olixops/internal/modules/cluster/service"
)

// NodeHandler 节点 HTTP 入口 (只读)。
type NodeHandler struct {
	svc *service.NodeService
}

// NewNodeHandler 构造 handler。
func NewNodeHandler(svc *service.NodeService) *NodeHandler {
	return &NodeHandler{svc: svc}
}

// list GET /api/v1/clusters/:id/nodes
func (h *NodeHandler) list(c *gin.Context) {
	// TODO:
	//   nodes, err := h.svc.List(c.Request.Context(), c.Param("id"))
	//   httpx.OK(c, nodes)
	_ = c
}

// get GET /api/v1/clusters/:id/nodes/:name
func (h *NodeHandler) get(c *gin.Context) {
	// TODO: h.svc.Get(ctx, clusterID, name) → httpx.OK
	_ = c
}

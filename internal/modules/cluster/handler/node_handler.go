package handler

import (
	"errors"
	"olixops/internal/platform/logger"
	"olixops/pkg/errs"
	"olixops/pkg/httpx"

	"olixops/internal/modules/cluster/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// NodeHandler
type NodeHandler struct {
	svc *service.NodeService
}

// NewNodeHandler 构造 handler
func NewNodeHandler(svc *service.NodeService) *NodeHandler {
	return &NodeHandler{
		svc: svc,
	}
}

type ClusterListRequest struct {
	ClusterID string `json:"clusterId" binding:"required"`
}

func (h *NodeHandler) List(c *gin.Context) {
	var req ClusterListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.L().Error(err.Error())
		httpx.Fail(c, errs.InvalidArg("invalid request"))
		return
	}

	list, err := h.svc.List(c, req.ClusterID)
	if err != nil {
		logger.L().Error("list error", zap.Error(err))
		httpx.Fail(c, errors.New(err.Error()))
		return
	}

	httpx.OK(c, list)
	return
}

func (h *NodeHandler) Get(c *gin.Context) {
	// TODO: h.svc.Get(ctx, clusterID, name) → httpx.OK
	_ = c
}

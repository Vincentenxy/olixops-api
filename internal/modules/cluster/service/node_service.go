package service

import (
	"context"
	"olixops/internal/platform/logger"

	"olixops/internal/modules/cluster/adapter"
	"olixops/internal/modules/cluster/domain"
	"olixops/internal/modules/cluster/repository"
	"olixops/pkg/errs"

	"go.uber.org/zap"
)

// NodeService 节点只读服务。
type NodeService struct {
	repo    repository.ClusterRepo
	factory adapter.Factory
}

// NewNodeService 构造服务
func NewNodeService(repo repository.ClusterRepo, factory adapter.Factory) *NodeService {
	return &NodeService{
		repo:    repo,
		factory: factory,
	}
}

// List 列集群节点
func (s *NodeService) List(ctx context.Context, clusterID string) ([]domain.Node, error) {
	client, err := s.factory.GetK8sClient(ctx, clusterID)
	if err != nil {
		logger.L().Error("GetK8sClient failed", zap.Error(err))
		return nil, err
	}

	nodes, err := client.ListNodes(ctx)
	if err != nil {
		logger.L().Error("ListNodes failed", zap.Error(err))
		return nil, err
	}

	return nodes, nil
}

// Get 取单个节点。
func (s *NodeService) Get(ctx context.Context, clusterID, name string) (*domain.Node, error) {
	// TODO: clientFor → cli.GetNode(ctx, name)
	_ = name
	return nil, errs.NotFound("TODO")
}

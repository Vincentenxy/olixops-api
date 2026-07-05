package service

import (
	"context"

	"olixops/internal/modules/cluster/adapter"
	"olixops/internal/modules/cluster/domain"
	"olixops/internal/modules/cluster/repository"
	"olixops/pkg/errs"
)

// NodeService 节点只读服务。
type NodeService struct {
	repo    repository.ClusterRepo
	factory adapter.Factory
}

// NewNodeService 构造服务。
func NewNodeService(repo repository.ClusterRepo, factory adapter.Factory) *NodeService {
	return &NodeService{repo: repo, factory: factory}
}

// List 列集群节点。
func (s *NodeService) List(ctx context.Context, clusterID string) ([]domain.Node, error) {
	// TODO: clientFor → cli.ListNodes(ctx)
	_ = clusterID
	return nil, errs.NotFound("TODO")
}

// Get 取单个节点。
func (s *NodeService) Get(ctx context.Context, clusterID, name string) (*domain.Node, error) {
	// TODO: clientFor → cli.GetNode(ctx, name)
	_ = name
	return nil, errs.NotFound("TODO")
}

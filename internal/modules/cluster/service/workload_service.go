package service

import (
	"context"

	"olixops/internal/modules/cluster/adapter"
	"olixops/internal/modules/cluster/domain"
	"olixops/internal/modules/cluster/repository"
	"olixops/pkg/errs"
)

// WorkloadService deployment / service / ingress 只读查询。
//
// 第一阶段只读, 写入操作归 application 或 helm 模块。
type WorkloadService struct {
	repo    repository.ClusterRepo
	factory adapter.Factory
}

// NewWorkloadService 构造服务。
func NewWorkloadService(repo repository.ClusterRepo, factory adapter.Factory) *WorkloadService {
	return &WorkloadService{repo: repo, factory: factory}
}

func (s *WorkloadService) clientFor(ctx context.Context, clusterID string) (adapter.K8sClient, error) {
	// TODO: 复用 namespaceService.clientFor 的逻辑, 提取成共享 helper
	_ = ctx
	return nil, errs.NotFound("TODO")
}

// ListDeployments 列 deployment, namespace 为空时列所有 namespace 的。
func (s *WorkloadService) ListDeployments(ctx context.Context, clusterID, namespace string) ([]domain.Deployment, error) {
	// TODO: clientFor → cli.ListDeployments(ctx, namespace)
	_ = namespace
	return nil, errs.NotFound("TODO")
}

// GetDeployment 取单个 deployment。
func (s *WorkloadService) GetDeployment(ctx context.Context, clusterID, namespace, name string) (*domain.Deployment, error) {
	// TODO
	_ = name
	return nil, errs.NotFound("TODO")
}

// ListServices 列 service。
func (s *WorkloadService) ListServices(ctx context.Context, clusterID, namespace string) ([]domain.Service, error) {
	// TODO
	return nil, errs.NotFound("TODO")
}

// GetService 取单个 service。
func (s *WorkloadService) GetService(ctx context.Context, clusterID, namespace, name string) (*domain.Service, error) {
	// TODO
	_ = name
	return nil, errs.NotFound("TODO")
}

// ListIngresses 列 ingress。
func (s *WorkloadService) ListIngresses(ctx context.Context, clusterID, namespace string) ([]domain.Ingress, error) {
	// TODO
	return nil, errs.NotFound("TODO")
}

// GetIngress 取单个 ingress。
func (s *WorkloadService) GetIngress(ctx context.Context, clusterID, namespace, name string) (*domain.Ingress, error) {
	// TODO
	_ = name
	return nil, errs.NotFound("TODO")
}

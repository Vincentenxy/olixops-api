package service

import (
	"context"
	"olixops/internal/platform/logger"
	"time"

	"olixops/internal/modules/cluster/adapter"
	"olixops/internal/modules/cluster/domain"
	"olixops/internal/modules/cluster/repository"
	"olixops/pkg/errs"

	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NamespaceService namespace 管理服务。
//
// 关键: 不存 ns 数据, 每次操作都实时查 K8s。
type NamespaceService struct {
	repo    repository.ClusterRepo // 用来根据 ID 查集群拿 kubeconfig
	factory adapter.Factory
}

// NewNamespaceService 构造服务
func NewNamespaceService(repo repository.ClusterRepo, factory adapter.Factory) *NamespaceService {
	return &NamespaceService{
		repo:    repo,
		factory: factory,
	}
}

// clientFor 根据 clusterID 拿到 K8sClient, 内部 helper, 三个 service 共用。
func (s *NamespaceService) clientFor(ctx context.Context, clusterID string) (adapter.K8sClient, error) {
	// TODO:
	//   1. s.repo.FindByID(ctx, clusterID) → 拿 cluster.Kubeconfig
	//   2. s.factory.Build(ctx, cluster.Kubeconfig) → K8sClient
	//   3. 检查 cluster.Status == active, 不是就返回 errs.Unavailable
	_ = ctx
	return nil, errs.NotFound("TODO")
}

type NsListParams struct {
	metav1.ListOptions
	ClusterID string
}

// List 列出集群的所有 namespace
func (s *NamespaceService) List(ctx context.Context, nsListParams NsListParams) ([]domain.Namespace, error) {
	client, err := s.factory.GetK8sClient(ctx, nsListParams.ClusterID)
	if err != nil {
		logger.L().Error("could not get k8s client", zap.Error(err))
		return nil, err
	}

	// add label selector
	var listOptions metav1.ListOptions
	listOptions.LabelSelector = nsListParams.LabelSelector
	listOptions.FieldSelector = nsListParams.FieldSelector

	namespaces, err := client.ListNamespaces(ctx, listOptions)
	if err != nil {
		logger.L().Error("could not list namespaces", zap.Error(err))
		return nil, err
	}
	return namespaces, err
}

// Get 取单个 namespace。
func (s *NamespaceService) Get(ctx context.Context, clusterID, name string) (*domain.Namespace, error) {
	// TODO: clientFor → cli.GetNamespace(ctx, name)
	_ = name
	return nil, errs.NotFound("TODO")
}

// CreateInput 创建 namespace 入参。
type CreateNamespaceInput struct {
	Name   string            `json:"name" binding:"required,min=1,max=63"`
	Labels map[string]string `json:"labels,omitempty"`
}

// Create 在 K8s 上创建 namespace。
func (s *NamespaceService) Create(ctx context.Context, clusterID string, in CreateNamespaceInput) (*domain.Namespace, error) {
	// TODO: clientFor → cli.CreateNamespace(ctx, in.Name, in.Labels)
	_ = in
	return nil, errs.NotFound("TODO")
}

// Delete 从 K8s 删除 namespace, 等待 Terminating 完成 (最多 timeout)。
func (s *NamespaceService) Delete(ctx context.Context, clusterID, name string, timeout time.Duration) error {
	// TODO: clientFor → cli.DeleteNamespace(ctx, name, timeout)
	_ = name
	_ = timeout
	return errs.NotFound("TODO")
}

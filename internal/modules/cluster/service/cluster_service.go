// Package service 编排 cluster 模块的业务逻辑。
//
// 业务代码只调:
//   - repository.ClusterRepository (持久化)
//   - adapter.K8sClient + adapter.Factory (K8s 通信)
//
// 不直接 import k8s.io/* 和 gorm.io/*。
package service

import (
	"context"
	"modules/cluster/adapter"
	"modules/cluster/domain"
	"modules/cluster/repository"
	"olixops/internal/platform/logger"
	"olixops/pkg/errs"
	"olixops/pkg/pagination"
	"time"

	"github.com/google/uuid"
)

// ClusterService 集群管理服务。
type ClusterService struct {
	repo    repository.ClusterRepo
	factory adapter.Factory
}

// NewClusterService 构造服务。
func NewClusterService(repo repository.ClusterRepo, factory adapter.Factory) *ClusterService {
	return &ClusterService{
		repo:    repo,
		factory: factory,
	}
}

// CreateInput 注册集群的入参
type CreateInput struct {
	ID             string `json:"id"`
	Name           string `json:"name" binding:"required,min=1,max=128"`
	TenantID       string `json:"tenantId" binding:"required,min=1,max=128"`
	Environment    string `json:"environment" binding:"required,max=32"`
	Description    string `json:"description" binding:"max=512"`
	KubeconfigPath string `json:"kubeconfigPath" binding:"required"`
}

// UpdateInput 更新集群入参 (名称/描述/环境, 不含 kubeconfig 和探测字段)。
type UpdateInput struct {
	Name        *string `json:"name,omitempty" binding:"omitempty,min=1,max=128"`
	Environment *string `json:"environment,omitempty" binding:"omitempty,max=32"`
	Description *string `json:"description,omitempty" binding:"max=512"`
}

func (s *ClusterService) Create(ctx context.Context, input CreateInput) error {

	log := logger.FromContext()

	cluster := &domain.Cluster{
		ID:       input.ID,
		TenantID: iput.TenantID,
	}

	return s.repo.Create(ctx, cluster)
}

// Register 注册新集群:
//  1. 校验名称不重复
//  2. 构造 K8sClient 立即探测 (Probe)
//  3. 探测成功 → 写库, status=active; 失败 → 写库, status=failed, 仍保留 (让用户看到错误)
func (s *ClusterService) Register(ctx context.Context, in CreateInput) (*domain.Cluster, error) {
	// TODO 实现步骤:
	//   1. 检查名称唯一: _, err := s.repo.FindByName(ctx, in.Name)
	//      err == nil → return nil, errs.AlreadyExists("cluster")
	//   2. 构造 client: k8sCli, err := s.factory.Build(ctx, in.Kubeconfig)
	//      err != nil → return nil, errs.InvalidArg("kubeconfig invalid: %v", err)
	//   3. 探测: version, nodeCount, err := k8sCli.Probe(ctx)
	//      err != nil → 仍然创建集群, 但 status=failed
	//   4. 提取 ServerURL (从 kubeconfig 解析)
	//   5. 构造 Cluster 实体, ID = uuid.NewString()
	//   6. s.repo.Create(ctx, cluster)
	//   7. return cluster, nil
	_ = ctx
	_ = in
	return nil, errs.NotFound("TODO")
}

// Get 按 ID 查询, 含实时探测 (强制刷新)。
func (s *ClusterService) Get(ctx context.Context, id string) (*domain.Cluster, error) {
	// TODO: 先 s.repo.FindByID, 再可选地重新探测 (Get 不强制刷新, 详情看 status)
	c, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return c, nil
}

// List 列表 (按 env / status 过滤, 不实时探测)。
func (s *ClusterService) List(ctx context.Context, q pagination.Query, filter repository.ListFilter) ([]*domain.Cluster, int64, error) {
	// TODO: 透传给 repo
	return s.repo.List(ctx, q, filter)
}

// Update 修改元数据 (name/description/environment), 不动 kubeconfig。
func (s *ClusterService) Update(ctx context.Context, id string, in UpdateInput) (*domain.Cluster, error) {
	// TODO:
	//   1. FindByID
	//   2. 应用 in 中的非 nil 字段
	//   3. s.repo.Update
	_ = in
	return nil, errs.NotFound("TODO")
}

// Refresh 主动重新探测集群, 更新 status / version / nodeCount。
func (s *ClusterService) Refresh(ctx context.Context, id string) (*domain.Cluster, error) {
	// TODO:
	//   1. FindByID
	//   2. factory.Build(kubeconfig)
	//   3. k8sCli.Probe
	//   4. 成功 → status=active, 失败 → status=unreachable
	//   5. repo.UpdateProbe (partial update)
	//   6. FindByID 返回最新
	_ = ctx
	return nil, errs.NotFound("TODO")
}

// Delete 注销集群 (软删除, 不连 K8s)。
func (s *ClusterService) Delete(ctx context.Context, id string) error {
	// TODO: s.repo.Delete(ctx, id)
	_ = ctx
	return errs.NotFound("TODO")
}

// 编译期防止忘记 import
var _ = time.Now
var _ = uuid.NewString

package repository

import (
	"context"
	"olixops/internal/modules/cluster/domain"
	"time"
)

// ClusterRepository 是 cluster 元数据的持久化接口
type ClusterRepo interface {

	// Create 新增集群
	Create(ctx context.Context, cluster *domain.Cluster) error

	// GetByID 根据ID查询单条（自动过滤软删除）
	GetByID(ctx context.Context, id string) (*domain.Cluster, error)

	// List 分页查询，按租户/环境/状态筛选
	List(ctx context.Context, filter *domain.ClusterListFilter) ([]*domain.Cluster, int64, error)

	// Update 更新集群基础信息
	Update(ctx context.Context, cluster *domain.Cluster) error

	// UpdateStatus 单独更新集群状态、探测时间（高频探针场景，单独优化）
	UpdateStatus(ctx context.Context, id string, status domain.ClusterStatus, lastProbeAt time.Time) error

	// Delete 软删除
	Delete(ctx context.Context, id string) error
}

package repository

import (
	"context"
	"olixops/internal/modules/cluster/domain"
	"time"

	"gorm.io/gorm"
)

type clusterRepo struct {
	db *gorm.DB
}

func (cr *clusterRepo) Create(ctx context.Context, cluster *domain.Cluster) error {
	return cr.db.WithContext(ctx).Create(cluster).Error
}

func (cr *clusterRepo) List(ctx context.Context, filter *domain.ClusterListFilter) ([]*domain.Cluster, int64, error) {
	tx := cr.db.WithContext(ctx)
	if filter.TenantID != "" {
		tx = tx.Where("tenant_id = ?", filter.TenantID)
	}
	if filter.Env != "" {
		tx = tx.Where("env = ?", filter.Env)
	}
	if filter.Status != "" {
		tx = tx.Where("status = ?", filter.Status)
	}

	var total int64
	err := tx.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// page
	var list []*domain.Cluster
	err = tx.Order("created_at desc").Offset(filter.Offset()).Limit(filter.Limit()).Find(&list).Error
	return list, total, err
}

func (cr *clusterRepo) Update(ctx context.Context, cluster *domain.Cluster) error {
	return cr.db.WithContext(ctx).Where("id = ?", cluster.ID).Updates(cluster).Error
}

func (cr *clusterRepo) UpdateStatus(ctx context.Context, id string, status domain.ClusterStatus, lastProbeAt time.Time) error {
	updates := map[string]any{
		"status":        status,
		"last_probe_at": lastProbeAt,
	}
	return cr.db.WithContext(ctx).Where("id = ?", id).Updates(updates).Error
}

func (cr *clusterRepo) Delete(ctx context.Context, id string) error {
	return cr.db.WithContext(ctx).Where("id = ?", id).Delete(&domain.Cluster{}).Error
}

func (cr *clusterRepo) GetByID(ctx context.Context, id string) (*domain.Cluster, error) {
	var c domain.Cluster
	err := cr.db.WithContext(ctx).Where("id = ?", id).First(&c).Error
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func NewClusterRepo(db *gorm.DB) ClusterRepo {
	return &clusterRepo{
		db: db,
	}
}

package repository

import (
	"context"
	"fmt"
	"olixops/internal/modules/cluster/domain"
	"olixops/internal/platform/logger"
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
	if cr.db == nil {
		logger.L().Error("clusterRepo.List db is nil")
		return nil, 0, fmt.Errorf("db is nil")
	}

	buildQuery := func() *gorm.DB {
		query := cr.db.WithContext(ctx).Model(&domain.Cluster{})
		if filter != nil {
			if filter.TenantID != "" {
				query = query.Where("tenant_id = ?", filter.TenantID)
			}
			if filter.Env != "" {
				query = query.Where("env = ?", filter.Env)
			}
			if filter.Status != "" {
				query = query.Where("status = ?", filter.Status)
			}
		}
		return query
	}

	// 统计总数
	var total int64
	if err := buildQuery().Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 查询数据
	var list []*domain.Cluster
	query := buildQuery().Order("created_at desc")
	if filter != nil {
		query = query.Offset(filter.Offset()).Limit(filter.Limit())
	}
	err := query.Find(&list).Error
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

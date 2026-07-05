// Package domain 定义 cluster 模块的实体。
//
// 实体是纯 struct, 不依赖任何外部库 (GORM tags 是唯一外部依赖, 可选)。
package domain

import (
	"olixops/pkg/pagination"
	"time"
)

// ClusterStatus 集群状态枚举。
//
// 状态机:
//
//	unknown -> connecting -> active
//	                |           |
//	                v           v
//	            failed      unreachable (探测失败后)
type ClusterStatus string

const (
	ClusterStatusUnknown     ClusterStatus = "unknown"
	ClusterStatusOnline      ClusterStatus = "online"
	ClusterStatusOffline     ClusterStatus = "offline"
	ClusterStatusUnreachable ClusterStatus = "unreachable"
)

type Cluster struct {
	ID          string        `json:"id" gorm:"primaryKey;size:64"`
	TenantID    string        `json:"tenantId" gorm:"size:64;index:idx_tenant_env;not null"`
	Name        string        `json:"name" gorm:"size:128;not null;uniqueIndex:idx_name"`
	Environment string        `json:"environment" gorm:"size:32;index:idx_tenant_env"`
	CreatorID   string        `json:"creatorId" gorm:"size:64;index"`
	Description string        `json:"description" gorm:"type:text"`
	KubeConfig  string        `json:"-" gorm:"type:text;not null"`
	Status      ClusterStatus `json:"status" gorm:"size:32;index:idx_status;default:unknown"`
	LastProbeAt time.Time     `json:"lastProbeAt"`
	LastSyncAt  *time.Time    `json:"lastSyncAt"`
	NodeCount   int           `json:"nodeCount" gorm:"default:0"`
	K8sVersion  string        `json:"k8sVersion" gorm:"size:64"`
	CreatedAt   time.Time     `json:"createdAt"`
	UpdatedAt   time.Time     `json:"updatedAt"`
	DeletedAt   *time.Time    `json:"-" gorm:"index"`
}

// TableName 指定表名
func (Cluster) TableName() string { return "clusters" }

// IsActive 判断集群是否处于可用状态
func (c *Cluster) IsActive() bool {
	return c.Status == ClusterStatusOnline
}

type ClusterListFilter struct {
	pagination.Query
	TenantID string `form:"tenantId" `
	Env      string `form:"env" `
	Status   string `form:"status" `
}

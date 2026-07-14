package domain

import (
	"time"

	"gorm.io/gorm"
)

// ReleaseStatus 是 Release 状态。
//
// 状态机:
//
//	pending -> deploying -> deployed
//	           |               |
//	           v               v
//	         failed      superseded (被新版本替代)
//	                          |
//	                          v
//	                     uninstalled
type ReleaseStatus string

const (
	ReleaseStatusPending      ReleaseStatus = "pending"      // 等待部署
	ReleaseStatusDeploying    ReleaseStatus = "deploying"    // 部署中
	ReleaseStatusDeployed     ReleaseStatus = "deployed"     // 部署成功
	ReleaseStatusFailed       ReleaseStatus = "failed"       // 部署失败
	ReleaseStatusSuperseded   ReleaseStatus = "superseded"   // 被新版本替代
	ReleaseStatusUninstalling ReleaseStatus = "uninstalling" // 卸载中
	ReleaseStatusUninstalled  ReleaseStatus = "uninstalled"  // 已卸载
)

// Release 是一次 Helm Release 的部署记录。
//
// 对应 Helm SDK 中的 release.Release, 但只持久化平台关心的元数据:
//   - 归属: 哪个应用 + 哪个环境 + 哪个集群/namespace
//   - Chart: 使用的 Chart 名称、版本、仓库
//   - Values: 用户自定义的 values (JSON 格式)
//   - 状态: 部署结果 (deployed / failed / ...)
//   - 历史: 每次 install/upgrade/rollback 都生成新版本记录
type Release struct {
	ID         string         `json:"id" gorm:"type:uuid;primaryKey"`
	Name       string         `json:"name" gorm:"size:128;not null;index:idx_release_ns,unique"` // Helm release name, 在 namespace 内唯一
	Namespace  string         `json:"namespace" gorm:"size:128;not null;index:idx_release_ns,unique"`
	ClusterID  string         `json:"cluster_id" gorm:"type:uuid;not null;index"`
	AppID      string         `json:"app_id" gorm:"type:uuid;index"`       // 关联应用 (可选)
	EnvID      string         `json:"env_id" gorm:"type:uuid;index"`       // 关联环境 (可选)
	RepoID     string         `json:"repo_id" gorm:"type:uuid;not null"`   // Chart 仓库
	ChartName  string         `json:"chart_name" gorm:"size:256;not null"` // Chart 名称
	ChartVer   string         `json:"chart_ver" gorm:"size:64;not null"`   // Chart 版本
	AppVersion string         `json:"app_version" gorm:"size:128"`         // Chart 中的 appVersion
	Version    int            `json:"version" gorm:"not null;default:1"`   // Helm revision 号, 每次 upgrade +1
	Values     string         `json:"values" gorm:"type:text"`             // 用户自定义 values (JSON)
	Status     ReleaseStatus  `json:"status" gorm:"size:16;not null;default:'pending';index"`
	Message    string         `json:"message" gorm:"type:text"` // 状态信息 (成功/失败原因)
	DeployedAt *time.Time     `json:"deployed_at"`              // 部署完成时间
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName 指定表名。
func (Release) TableName() string { return "olix_helm_releases" }

// IsDeployed 判断 release 是否处于已部署状态。
func (r *Release) IsDeployed() bool {
	return r.Status == ReleaseStatusDeployed
}

// ReleaseListFilter 是 release 列表查询的过滤条件。
type ReleaseListFilter struct {
	ClusterID string        `form:"cluster_id"`
	Namespace string        `form:"namespace"`
	AppID     string        `form:"app_id"`
	Status    ReleaseStatus `form:"status"`
	ChartName string        `form:"chart_name"`
}

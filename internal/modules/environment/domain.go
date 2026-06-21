// Package environment 实现环境领域模型。
//
// 环境是项目下的逻辑隔离单元(如 dev / test / sim / prod),挂载具体的集群、命名空间、变量和审批策略。
package environment

import (
	"time"

	"gorm.io/gorm"
)

// Type 是环境类型。
type Type string

const (
	TypeDev  Type = "dev"
	TypeTest Type = "test"
	TypeSim  Type = "sim"
	TypeProd Type = "prod"
)

// Status 是环境状态。
type Status string

const (
	StatusActive   Status = "active"
	StatusInactive Status = "inactive"
)

// Environment 是环境聚合根。
type Environment struct {
	ID            string         `gorm:"type:uuid;primaryKey" json:"id"`
	ProjectID     string         `gorm:"type:uuid;index:uk_project_code,unique" json:"project_id"`
	Code          string         `gorm:"type:varchar(64);index:uk_project_code,unique" json:"code"`
	Name          string         `gorm:"type:varchar(128);not null" json:"name"`
	Type          Type           `gorm:"type:varchar(16);not null;index" json:"type"`
	Description   string         `gorm:"type:text" json:"description,omitempty"`
	ClusterID     string         `gorm:"type:uuid;index" json:"cluster_id,omitempty"`
	Namespace     string         `gorm:"type:varchar(128)" json:"namespace,omitempty"`
	ApprovalLevel int            `gorm:"not null;default:0" json:"approval_level"`
	Order         int            `gorm:"not null;default:0" json:"order"`
	Status        Status         `gorm:"type:varchar(16);not null;default:'active';index" json:"status"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定数据库表名。
func (Environment) TableName() string { return "olix_environments" }

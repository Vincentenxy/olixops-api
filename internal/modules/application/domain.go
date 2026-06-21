// Package application 实现应用领域模型。
//
// 应用是平台运行单元的抽象,可以是普通服务、前端应用、任务应用或智能体应用。
// 每个应用归属一个项目,挂载多个 ServiceComponent。
package application

import (
	"time"

	"gorm.io/gorm"
)

// Kind 是应用类型。
type Kind string

const (
	KindService  Kind = "service"  // 普通业务服务
	KindFrontend Kind = "frontend" // 前端应用
	KindJob      Kind = "job"      // 后台任务
	KindAgent    Kind = "agent"    // 智能体应用
	KindSandbox  Kind = "sandbox"  // 沙箱应用
)

// Runtime 是应用运行时类型。
type Runtime string

const (
	RuntimeGo     Runtime = "go"
	RuntimeJava   Runtime = "java"
	RuntimeNode   Runtime = "node"
	RuntimePython Runtime = "python"
	RuntimeRust   Runtime = "rust"
	RuntimeOther  Runtime = "other"
)

// Status 是应用状态。
type Status string

const (
	StatusActive   Status = "active"
	StatusArchived Status = "archived"
)

// Application 是应用聚合根。
type Application struct {
	ID          string         `gorm:"type:uuid;primaryKey" json:"id"`
	ProjectID   string         `gorm:"type:uuid;index:uk_proj_key,unique;not null" json:"project_id"`
	Key         string         `gorm:"type:varchar(64);index:uk_proj_key,unique;not null" json:"key"`
	Name        string         `gorm:"type:varchar(128);not null" json:"name"`
	Description string         `gorm:"type:text" json:"description,omitempty"`
	Kind        Kind           `gorm:"type:varchar(32);not null;index" json:"kind"`
	Runtime     Runtime        `gorm:"type:varchar(32)" json:"runtime,omitempty"`
	RepoURL     string         `gorm:"type:varchar(512)" json:"repo_url,omitempty"`
	OwnerID     string         `gorm:"type:uuid;index" json:"owner_id,omitempty"`
	Status      Status         `gorm:"type:varchar(16);not null;default:'active';index" json:"status"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定数据库表名。
func (Application) TableName() string { return "olix_applications" }

// ServiceComponent 是应用下的部署组件。
type ServiceComponent struct {
	ID            string         `gorm:"type:uuid;primaryKey" json:"id"`
	ApplicationID string         `gorm:"type:uuid;index;not null" json:"application_id"`
	Name          string         `gorm:"type:varchar(128);not null" json:"name"`
	Workload      string         `gorm:"type:varchar(32);not null" json:"workload"` // Deployment / StatefulSet / Job 等
	Image         string         `gorm:"type:varchar(512)" json:"image,omitempty"`
	Port          int            `json:"port,omitempty"`
	HealthCheck   string         `gorm:"type:varchar(256)" json:"health_check,omitempty"`
	Replicas      int            `json:"replicas"`
	CPURequest    string         `gorm:"type:varchar(32)" json:"cpu_request,omitempty"`
	CPULimit      string         `gorm:"type:varchar(32)" json:"cpu_limit,omitempty"`
	MemoryRequest string         `gorm:"type:varchar(32)" json:"memory_request,omitempty"`
	MemoryLimit   string         `gorm:"type:varchar(32)" json:"memory_limit,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名。
func (ServiceComponent) TableName() string { return "olix_service_components" }

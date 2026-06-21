// Package project 实现项目领域模型。
//
// 项目是平台的核心组织单元,挂载应用、环境、权限、流水线等。
package project

import (
	"time"

	"gorm.io/gorm"
)

// Visibility 是项目可见性。
type Visibility string

const (
	VisibilityPrivate  Visibility = "private"
	VisibilityInternal Visibility = "internal"
	VisibilityPublic   Visibility = "public"
)

// Status 是项目状态。
type Status string

const (
	StatusActive   Status = "active"
	StatusArchived Status = "archived"
	StatusDeleted  Status = "deleted"
)

// Project 是项目聚合根。
type Project struct {
	ID           string         `gorm:"type:uuid;primaryKey" json:"id"`
	Key          string         `gorm:"type:varchar(64);uniqueIndex;not null" json:"key"`
	Name         string         `gorm:"type:varchar(128);not null" json:"name"`
	Description  string         `gorm:"type:text" json:"description,omitempty"`
	OwnerID      string         `gorm:"type:uuid;index;not null" json:"owner_id"`
	BusinessUnit string         `gorm:"type:varchar(128)" json:"business_unit,omitempty"`
	RepoURL      string         `gorm:"type:varchar(512)" json:"repo_url,omitempty"`
	Visibility   Visibility     `gorm:"type:varchar(16);not null;default:'private'" json:"visibility"`
	Status       Status         `gorm:"type:varchar(16);not null;default:'active';index" json:"status"`
	Labels       Labels         `gorm:"type:jsonb" json:"labels,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定数据库表名。
func (Project) TableName() string { return "olix_projects" }

// Member 是项目成员关联表。
type Member struct {
	ID        string    `gorm:"type:uuid;primaryKey" json:"id"`
	ProjectID string    `gorm:"type:uuid;index:uk_project_user,unique" json:"project_id"`
	UserID    string    `gorm:"type:uuid;index:uk_project_user,unique" json:"user_id"`
	Role      string    `gorm:"type:varchar(64);not null" json:"role"`
	InvitedBy string    `gorm:"type:uuid" json:"invited_by,omitempty"`
	JoinedAt  time.Time `json:"joined_at"`
}

// TableName 指定表名。
func (Member) TableName() string { return "olix_project_members" }

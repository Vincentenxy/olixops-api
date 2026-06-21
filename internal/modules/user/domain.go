// Package user 实现用户领域模型。
//
// 模块边界:用户的注册、登录、信息变更、状态管理。
// 不直接依赖任何外部 SDK,所有协议适配走 platform/auth。
package user

import (
	"time"

	"gorm.io/gorm"
)

// Status 是用户状态。
type Status string

const (
	StatusActive   Status = "active"
	StatusInactive Status = "inactive"
	StatusLocked   Status = "locked"
	StatusDeleted  Status = "deleted"
)

// User 是用户聚合根。
type User struct {
	ID           string         `gorm:"type:uuid;primaryKey" json:"id"`
	Username     string         `gorm:"type:varchar(64);uniqueIndex;not null" json:"username"`
	Email        string         `gorm:"type:varchar(255);uniqueIndex" json:"email"`
	DisplayName  string         `gorm:"type:varchar(128)" json:"display_name"`
	PasswordHash string         `gorm:"type:varchar(255)" json:"-"`
	PhoneNumber  string         `gorm:"type:varchar(32)" json:"phone_number,omitempty"`
	AvatarURL    string         `gorm:"type:varchar(512)" json:"avatar_url,omitempty"`
	Status       Status         `gorm:"type:varchar(16);not null;default:'active';index" json:"status"`
	Source       string         `gorm:"type:varchar(32);default:'local'" json:"source"` // local / oauth2 / ldap
	ExternalID   string         `gorm:"type:varchar(255);index" json:"external_id,omitempty"`
	LastLoginAt  *time.Time     `json:"last_login_at,omitempty"`
	LastLoginIP  string         `gorm:"type:varchar(64)" json:"last_login_ip,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定数据库表名。
func (User) TableName() string { return "olix_users" }

// IsActive 仅活跃状态可登录。
func (u *User) IsActive() bool { return u.Status == StatusActive }

// Package domain 定义 helm 模块的实体。
//
// 实体是纯 struct, 不依赖任何外部库 (GORM tags 是唯一外部依赖, 可选)。
package domain

import (
	"olixops/pkg/pagination"
	"time"

	"gorm.io/gorm"
)

// RepoType 是 Chart 仓库类型。
//
// 初期对接公共仓库, 主要支持 HTTP 和 OCI 两种协议:
//   - HTTP: 传统 Chart Repository (index.yaml + tgz)
//   - OCI:  OCI Registry (oras / docker pull 协议)
//   - Git:  本地/Git 仓库中的 Chart 目录 (后续扩展)
type RepoType string

const (
	RepoTypeHTTP RepoType = "http" // 传统 HTTP Chart Repository
	RepoTypeOCI  RepoType = "oci"  // OCI Registry
	RepoTypeGit  RepoType = "git"  // Git 仓库 (后续扩展)
)

// RepoStatus 是 Chart 仓库状态。
//
// 状态机:
//
//	unknown -> syncing -> active
//	            |           |
//	            v           v
//	        failed      unreachable (同步失败后)
type RepoStatus string

const (
	RepoStatusUnknown     RepoStatus = "unknown"
	RepoStatusActive      RepoStatus = "active"
	RepoStatusSyncing     RepoStatus = "syncing"
	RepoStatusFailed      RepoStatus = "failed"
	RepoStatusUnreachable RepoStatus = "unreachable"
)

// ChartRepo 是 Chart 仓库聚合根。
//
// 一条记录代表一个已注册的 Chart 仓库, 平台通过它拉取 Chart 索引和下载 Chart 包。
// 初期对接公共仓库 (如 bitnami, stable 等), 无需认证; 后续可扩展私有仓库 + 认证。
type ChartRepo struct {
	ID          string         `json:"id" gorm:"type:uuid;primaryKey"`
	Name        string         `json:"name" gorm:"size:128;not null;uniqueIndex:idx_repo_name"`
	URL         string         `json:"url" gorm:"size:512;not null"`
	Type        RepoType       `json:"type" gorm:"size:16;not null;default:'http'"`
	Username    string         `json:"-" gorm:"size:128"`            // 认证用户名 (可选, json:"-" 不对外暴露)
	Password    string         `json:"-" gorm:"size:256"`            // 认证密码 (可选, 后续可加密存储)
	Description string         `json:"description" gorm:"type:text"` // 仓库描述
	Status      RepoStatus     `json:"status" gorm:"size:16;not null;default:'unknown';index"`
	LastSyncAt  *time.Time     `json:"last_sync_at"`                 // 最近一次索引同步时间
	ChartCount  int            `json:"chart_count" gorm:"default:0"` // 仓库中 Chart 数量 (同步后更新)
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName 指定表名。
func (ChartRepo) TableName() string { return "olix_helm_repos" }

// IsActive 判断仓库是否处于可用状态。
func (r *ChartRepo) IsActive() bool {
	return r.Status == RepoStatusActive
}

// ChartMeta 是 Chart 的元数据 (从仓库 index.yaml 解析后缓存)。
//
// TODO: 初期可以先不持久化, 每次从仓库实时拉取 index.yaml。
//
//	后续可以建表缓存, 减少网络请求。
type ChartMeta struct {
	Name        string    `json:"name"`
	Version     string    `json:"version"`
	AppVersion  string    `json:"app_version"`
	Description string    `json:"description"`
	Icon        string    `json:"icon"`
	Home        string    `json:"home"`
	Sources     []string  `json:"sources"`
	Keywords    []string  `json:"keywords"`
	CreatedAt   time.Time `json:"created_at"`
}

// ChartVersion 是 Chart 的某个具体版本信息。
//
// 用于列表展示某个 Chart 有哪些可用版本。
type ChartVersion struct {
	Version     string    `json:"version"`
	AppVersion  string    `json:"app_version"`
	Description string    `json:"description"`
	URLs        []string  `json:"urls"` // tgz 下载地址
	Created     time.Time `json:"created"`
	Deprecated  bool      `json:"deprecated"`
}

// RepoListFilter 是仓库列表查询的过滤条件。
type RepoListFilter struct {
	pagination.Query
	Name   string     `form:"name"`
	Type   RepoType   `form:"type"`
	Status RepoStatus `form:"status"`
}

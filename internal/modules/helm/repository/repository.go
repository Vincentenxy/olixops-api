package repository

import (
	"context"
	"olixops/internal/modules/helm/domain"
)

// HelmRepo 是 Helm 模块的持久化接口, 聚合 ChartRepo 和 Release 两类实体的存储操作。
//
// 业务代码 (service) 只依赖本接口, 不直接 import gorm。
// 实现: helmRepo (GORM)
// 测试: 可用内存实现或 mock。
type HelmRepo interface {

	// ─────────────────────────────────────────────────────────────────────
	// ChartRepo 仓库管理
	// ─────────────────────────────────────────────────────────────────────

	// CreateRepo 新增 Chart 仓库。
	CreateRepo(ctx context.Context, repo *domain.ChartRepo) error

	// GetRepoByID 根据 ID 查询仓库。
	GetRepoByID(ctx context.Context, id string) (*domain.ChartRepo, error)

	// GetRepoByName 根据名称查询仓库 (唯一性校验用)。
	GetRepoByName(ctx context.Context, name string) (*domain.ChartRepo, error)

	// ListRepos 分页列表查询仓库。
	ListRepos(ctx context.Context, filter *domain.RepoListFilter) ([]*domain.ChartRepo, int64, error)

	// UpdateRepo 更新仓库信息 (URL/认证/描述)。
	UpdateRepo(ctx context.Context, repo *domain.ChartRepo) error

	// UpdateRepoStatus 单独更新仓库状态 + 同步时间 (高频同步场景, 单独优化)。
	UpdateRepoStatus(ctx context.Context, id string, status domain.RepoStatus) error

	// DeleteRepo 软删除仓库。
	DeleteRepo(ctx context.Context, id string) error

	// ─────────────────────────────────────────────────────────────────────
	// Release 管理
	// ─────────────────────────────────────────────────────────────────────

	// CreateRelease 新增 Release 记录 (install 时创建)。
	CreateRelease(ctx context.Context, rel *domain.Release) error

	// GetReleaseByID 根据 ID 查询 Release。
	GetReleaseByID(ctx context.Context, id string) (*domain.Release, error)

	// GetReleaseByName 根据 release name + namespace 查询 (Helm 唯一键)。
	GetReleaseByName(ctx context.Context, namespace, name string) (*domain.Release, error)

	// ListReleases 列表查询 Release。
	ListReleases(ctx context.Context, filter *domain.ReleaseListFilter) ([]*domain.Release, int64, error)

	// UpdateRelease 更新 Release (upgrade/rollback 后更新 chart_ver/values/status 等)。
	UpdateRelease(ctx context.Context, rel *domain.Release) error

	// UpdateReleaseStatus 单独更新状态 + message (部署过程中状态流转)。
	UpdateReleaseStatus(ctx context.Context, id string, status domain.ReleaseStatus, message string) error

	// DeleteRelease 软删除 Release 记录。
	DeleteRelease(ctx context.Context, id string) error
}

package repository

import (
	"context"
	"olixops/internal/modules/helm/domain"

	"gorm.io/gorm"
)

// helmRepo 是 HelmRepo 的 GORM 实现。
//
// TODO: 你需要实现下面的所有方法, 参考 cluster/repository/cluster_repo.go 的写法。
//
//	每个方法的核心逻辑:
//	  - CreateXxx:    db.WithContext(ctx).Create(entity).Error
//	  - GetXxxByID:   db.WithContext(ctx).Where("id = ?", id).First(&entity).Error
//	  - GetXxxByName: db.WithContext(ctx).Where("name = ?", name).First(&entity).Error
//	  - ListXxx:      构建 query → Count(&total) → Offset/Limit → Find(&list)
//	  - UpdateXxx:    db.WithContext(ctx).Where("id = ?", id).Updates(entity).Error
//	  - DeleteXxx:    db.WithContext(ctx).Where("id = ?", id).Delete(&entity).Error (GORM 软删除)
type helmRepo struct {
	db *gorm.DB
}

// NewHelmRepo 构造 HelmRepo。
func NewHelmRepo(db *gorm.DB) HelmRepo {
	return &helmRepo{db: db}
}

// ─────────────────────────────────────────────────────────────────────
// ChartRepo 仓库管理
// ─────────────────────────────────────────────────────────────────────

func (r *helmRepo) CreateRepo(ctx context.Context, repo *domain.ChartRepo) error {
	// TODO: 实现
	// return r.db.WithContext(ctx).Create(repo).Error
	_ = ctx
	_ = repo
	return nil
}

func (r *helmRepo) GetRepoByID(ctx context.Context, id string) (*domain.ChartRepo, error) {
	// TODO: 实现
	// var repo domain.ChartRepo
	// err := r.db.WithContext(ctx).Where("id = ?", id).First(&repo).Error
	// if err != nil { return nil, err }
	// return &repo, nil
	_ = ctx
	_ = id
	return nil, nil
}

func (r *helmRepo) GetRepoByName(ctx context.Context, name string) (*domain.ChartRepo, error) {
	// TODO: 实现
	// var repo domain.ChartRepo
	// err := r.db.WithContext(ctx).Where("name = ?", name).First(&repo).Error
	// if err != nil { return nil, err }
	// return &repo, nil
	_ = ctx
	_ = name
	return nil, nil
}

func (r *helmRepo) ListRepos(ctx context.Context, filter *domain.RepoListFilter) ([]*domain.ChartRepo, int64, error) {
	// TODO: 实现, 参考 clusterRepo.List 的分页查询写法
	//   1. 构建 query: r.db.WithContext(ctx).Model(&domain.ChartRepo{})
	//   2. 按 filter 条件 Where 过滤 (Name/Type/Status)
	//   3. Count(&total)
	//   4. Offset/Limit + Order("created_at desc") + Find(&list)
	_ = ctx
	_ = filter
	return nil, 0, nil
}

func (r *helmRepo) UpdateRepo(ctx context.Context, repo *domain.ChartRepo) error {
	// TODO: 实现
	// return r.db.WithContext(ctx).Where("id = ?", repo.ID).Updates(repo).Error
	_ = ctx
	_ = repo
	return nil
}

func (r *helmRepo) UpdateRepoStatus(ctx context.Context, id string, status domain.RepoStatus) error {
	// TODO: 实现, 单独更新 status + last_sync_at
	// updates := map[string]any{
	//     "status":       status,
	//     "last_sync_at": time.Now(),
	// }
	// return r.db.WithContext(ctx).Where("id = ?", id).Updates(updates).Error
	_ = ctx
	_ = id
	_ = status
	return nil
}

func (r *helmRepo) DeleteRepo(ctx context.Context, id string) error {
	// TODO: 实现
	// return r.db.WithContext(ctx).Where("id = ?", id).Delete(&domain.ChartRepo{}).Error
	_ = ctx
	_ = id
	return nil
}

// ─────────────────────────────────────────────────────────────────────
// Release 管理
// ─────────────────────────────────────────────────────────────────────

func (r *helmRepo) CreateRelease(ctx context.Context, rel *domain.Release) error {
	// TODO: 实现
	// return r.db.WithContext(ctx).Create(rel).Error
	_ = ctx
	_ = rel
	return nil
}

func (r *helmRepo) GetReleaseByID(ctx context.Context, id string) (*domain.Release, error) {
	// TODO: 实现
	// var rel domain.Release
	// err := r.db.WithContext(ctx).Where("id = ?", id).First(&rel).Error
	// if err != nil { return nil, err }
	// return &rel, nil
	_ = ctx
	_ = id
	return nil, nil
}

func (r *helmRepo) GetReleaseByName(ctx context.Context, namespace, name string) (*domain.Release, error) {
	// TODO: 实现, namespace + name 联合唯一
	// var rel domain.Release
	// err := r.db.WithContext(ctx).
	//     Where("namespace = ? AND name = ?", namespace, name).
	//     First(&rel).Error
	// if err != nil { return nil, err }
	// return &rel, nil
	_ = ctx
	_ = namespace
	_ = name
	return nil, nil
}

func (r *helmRepo) ListReleases(ctx context.Context, filter *domain.ReleaseListFilter) ([]*domain.Release, int64, error) {
	// TODO: 实现, 参考 ListRepos 的分页写法
	//   按 filter.ClusterID / Namespace / AppID / Status / ChartName 过滤
	_ = ctx
	_ = filter
	return nil, 0, nil
}

func (r *helmRepo) UpdateRelease(ctx context.Context, rel *domain.Release) error {
	// TODO: 实现
	// return r.db.WithContext(ctx).Where("id = ?", rel.ID).Updates(rel).Error
	_ = ctx
	_ = rel
	return nil
}

func (r *helmRepo) UpdateReleaseStatus(ctx context.Context, id string, status domain.ReleaseStatus, message string) error {
	// TODO: 实现
	// updates := map[string]any{
	//     "status":  status,
	//     "message": message,
	// }
	// return r.db.WithContext(ctx).Where("id = ?", id).Updates(updates).Error
	_ = ctx
	_ = id
	_ = status
	_ = message
	return nil
}

func (r *helmRepo) DeleteRelease(ctx context.Context, id string) error {
	// TODO: 实现
	// return r.db.WithContext(ctx).Where("id = ?", id).Delete(&domain.Release{}).Error
	_ = ctx
	_ = id
	return nil
}

// 编译期断言: helmRepo 必须实现 HelmRepo 接口。
var _ HelmRepo = (*helmRepo)(nil)

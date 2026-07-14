package service

import (
	"context"
	"olixops/internal/modules/helm/adapter"
	"olixops/internal/modules/helm/domain"
	"olixops/internal/modules/helm/repository"
	"olixops/internal/platform/logger"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ─────────────────────────────────────────────────────────────────────
// RepoService 是 Chart 仓库管理的应用服务。
// ─────────────────────────────────────────────────────────────────────

// RepoService 仓库管理服务。
type RepoService struct {
	repo    repository.HelmRepo
	factory adapter.Factory
}

// NewRepoService 构造服务。
func NewRepoService(repo repository.HelmRepo, factory adapter.Factory) *RepoService {
	return &RepoService{repo: repo, factory: factory}
}

// ─────────────────────────────────────────────────────────────────────
// 入参结构体
// ─────────────────────────────────────────────────────────────────────

// AddRepoInput 添加仓库入参。
type AddRepoInput struct {
	Name        string          `json:"name" binding:"required,min=2,max=128"`
	URL         string          `json:"url" binding:"required,url"`
	Type        domain.RepoType `json:"type" binding:"required,oneof=http oci"`
	Username    string          `json:"username"` // 可选
	Password    string          `json:"password"` // 可选
	Description string          `json:"description"`
}

// UpdateRepoInput 更新仓库入参。
type UpdateRepoInput struct {
	URL         *string `json:"url,omitempty" binding:"omitempty,url"`
	Username    *string `json:"username,omitempty"`
	Password    *string `json:"password,omitempty"`
	Description *string `json:"description,omitempty"`
}

// ─────────────────────────────────────────────────────────────────────
// 方法实现
// ─────────────────────────────────────────────────────────────────────

// Add 添加一个新的 Chart 仓库。
//
// 实现步骤:
//   1. 校验名称唯一性: repo.GetRepoByName → 已存在则报错
//   2. 构造 ChartRepo 实体 (ID = uuid.NewString(), Status = unknown)
//   3. repo.CreateRepo 持久化
//   4. 可选: 异步触发一次 Sync (更新 status → active/failed)
func (s *RepoService) Add(ctx context.Context, in AddRepoInput) (*domain.ChartRepo, error) {
	log := logger.FromContext(ctx)
	log.Info("adding helm repo", zap.String("name", in.Name), zap.String("url", in.URL))

	// TODO 步骤 1: 名称唯一性校验
	// existing, err := s.repo.GetRepoByName(ctx, in.Name)
	// if err == nil && existing != nil {
	//     return nil, &domain.ErrRepoAlreadyExists{Name: in.Name}
	// }

	// TODO 步骤 2: 构造实体
	// chartRepo := &domain.ChartRepo{
	//     ID:          uuid.NewString(),
	//     Name:        in.Name,
	//     URL:         in.URL,
	//     Type:        in.Type,
	//     Username:    in.Username,
	//     Password:    in.Password,
	//     Description: in.Description,
	//     Status:      domain.RepoStatusUnknown,
	// }

	// TODO 步骤 3: 持久化
	// if err := s.repo.CreateRepo(ctx, chartRepo); err != nil {
	//     return nil, err
	// }

	// TODO 步骤 4 (可选): 异步同步索引
	// go s.Sync(context.Background(), chartRepo.ID)

	_ = ctx
	_ = in
	_ = uuid.NewString
	return nil, nil
}

// Sync 同步仓库索引 (拉取 index.yaml, 更新 ChartCount + LastSyncAt + Status)。
//
// 实现步骤:
//   1. GetRepoByID 获取仓库信息
//   2. 更新状态为 syncing: repo.UpdateRepoStatus(id, syncing)
//   3. 调用 adapter.FetchIndex(URL, Type, Username, Password)
//   4. 成功: 更新 ChartCount = len(charts), Status = active, LastSyncAt = now
//   5. 失败: 更新 Status = failed, Message = err.Error()
func (s *RepoService) Sync(ctx context.Context, repoID string) error {
	log := logger.FromContext(ctx)
	log.Info("syncing helm repo", zap.String("repo_id", repoID))

	// TODO 步骤 1: 查询仓库
	// chartRepo, err := s.repo.GetRepoByID(ctx, repoID)
	// if err != nil { return err }

	// TODO 步骤 2: 更新为同步中
	// s.repo.UpdateRepoStatus(ctx, repoID, domain.RepoStatusSyncing)

	// TODO 步骤 3: 构造 HelmClient 并拉取索引
	// client, err := s.factory.Build(ctx, "") // Sync 不需要 kubeconfig
	// charts, err := client.FetchIndex(ctx, chartRepo.URL, chartRepo.Type, chartRepo.Username, chartRepo.Password)

	// TODO 步骤 4/5: 根据结果更新状态
	// 成功: UpdateRepoStatus(active), 更新 ChartCount
	// 失败: UpdateRepoStatus(failed)

	_ = ctx
	_ = repoID
	return nil
}

// Get 根据 ID 查询仓库。
func (s *RepoService) Get(ctx context.Context, id string) (*domain.ChartRepo, error) {
	// TODO: return s.repo.GetRepoByID(ctx, id)
	_ = ctx
	_ = id
	return nil, nil
}

// List 列表查询仓库 (分页 + 过滤)。
func (s *RepoService) List(ctx context.Context, filter *domain.RepoListFilter) ([]*domain.ChartRepo, int64, error) {
	// TODO: return s.repo.ListRepos(ctx, filter)
	_ = ctx
	_ = filter
	return nil, 0, nil
}

// Update 更新仓库信息。
//
// 实现步骤:
//   1. GetRepoByID 获取
//   2. 逐字段更新 (URL/Username/Password/Description)
//   3. repo.UpdateRepo 持久化
func (s *RepoService) Update(ctx context.Context, id string, in UpdateRepoInput) (*domain.ChartRepo, error) {
	// TODO: 实现
	_ = ctx
	_ = id
	_ = in
	return nil, nil
}

// Delete 删除仓库。
//
// 实现步骤:
//   1. GetRepoByID 确认存在
//   2. (可选) 检查该仓库下是否有 Release 引用
//   3. repo.DeleteRepo 软删除
func (s *RepoService) Delete(ctx context.Context, id string) error {
	// TODO: 实现
	_ = ctx
	_ = id
	return nil
}

// ListCharts 列出仓库中的所有 Chart (实时拉取 index.yaml, 不持久化)。
//
// 实现步骤:
//   1. GetRepoByID 获取仓库 URL/认证
//   2. factory.Build → FetchIndex
//   3. 返回 []*domain.ChartMeta
func (s *RepoService) ListCharts(ctx context.Context, repoID string) ([]*domain.ChartMeta, error) {
	// TODO: 实现
	_ = ctx
	_ = repoID
	return nil, nil
}

// ListChartVersions 列出某个 Chart 的所有版本。
//
// 实现步骤:
//   1. GetRepoByID 获取仓库 URL/认证
//   2. factory.Build → FetchChartVersions
//   3. 返回 []*domain.ChartVersion
func (s *RepoService) ListChartVersions(ctx context.Context, repoID, chartName string) ([]*domain.ChartVersion, error) {
	// TODO: 实现
	_ = ctx
	_ = repoID
	_ = chartName
	return nil, nil
}

// GetChartDefaultValues 获取某个 Chart 的默认 values.yaml。
//
// 实现步骤:
//   1. GetRepoByID 获取仓库 URL/认证
//   2. factory.Build → RenderValues
//   3. 返回 map[string]any
func (s *RepoService) GetChartDefaultValues(ctx context.Context, repoID, chartName, chartVer string) (map[string]any, error) {
	// TODO: 实现
	_ = ctx
	_ = repoID
	_ = chartName
	_ = chartVer
	return nil, nil
}

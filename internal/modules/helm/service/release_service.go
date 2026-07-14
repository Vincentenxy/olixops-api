package service

import (
	"context"
	"encoding/json"
	"olixops/internal/modules/helm/adapter"
	"olixops/internal/modules/helm/domain"
	"olixops/internal/modules/helm/repository"
	"olixops/internal/platform/logger"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"k8s.io/client-go/rest"
)

// ─────────────────────────────────────────────────────────────────────
// ReleaseService 是 Helm Release 生命周期的应用服务。
// ─────────────────────────────────────────────────────────────────────

// ReleaseService Release 管理服务。
type ReleaseService struct {
	repo        repository.HelmRepo
	factory     adapter.Factory
	clusterRepo ClusterKubeconfigProvider // 从 cluster 模块获取 kubeconfig
}

// ClusterKubeconfigProvider 从 cluster 模块获取 rest.Config 的接口。
//
// helm 模块不直接依赖 cluster 模块的 repository, 而是通过这个接口解耦。
// 在 module.go 的 Assemble 中注入 cluster 模块的实现。
//
// 返回 *rest.Config 而非 kubeconfig 字符串, 这样可以复用 cluster 模块已有的
// kubeconfig 解析 + 解密逻辑, Helm 模块不需要重复处理。
//
// TODO: 在 cluster 模块中实现这个接口, 核心逻辑:
//   1. 根据 clusterID 从 DB 查出 Cluster 记录
//   2. 解密 KubeConfig 字段 (AES-GCM)
//   3. 用 clientcmd.RESTConfigFromKubeConfig 构造 *rest.Config
//   4. 返回
type ClusterKubeconfigProvider interface {
	// GetRESTConfig 根据 clusterID 获取 *rest.Config。
	GetRESTConfig(ctx context.Context, clusterID string) (*rest.Config, error)
}

// NewReleaseService 构造服务。
func NewReleaseService(
	repo repository.HelmRepo,
	factory adapter.Factory,
	clusterProvider ClusterKubeconfigProvider,
) *ReleaseService {
	return &ReleaseService{
		repo:        repo,
		factory:     factory,
		clusterRepo: clusterProvider,
	}
}

// ─────────────────────────────────────────────────────────────────────
// 入参结构体
// ─────────────────────────────────────────────────────────────────────

// InstallInput 安装 Release 入参。
type InstallInput struct {
	Name      string         `json:"name" binding:"required,min=2,max=128"`       // Helm release name
	Namespace string         `json:"namespace" binding:"required,min=2,max=128"`  // 目标 namespace
	ClusterID string         `json:"cluster_id" binding:"required,uuid"`          // 目标集群
	AppID     string         `json:"app_id" binding:"omitempty,uuid"`             // 关联应用 (可选)
	EnvID     string         `json:"env_id" binding:"omitempty,uuid"`             // 关联环境 (可选)
	RepoID    string         `json:"repo_id" binding:"required,uuid"`             // Chart 仓库
	ChartName string         `json:"chart_name" binding:"required,min=2,max=256"` // Chart 名称
	ChartVer  string         `json:"chart_ver" binding:"required,min=1,max=64"`   // Chart 版本
	Values    map[string]any `json:"values"`                                      // 用户自定义 values (可选)
	Wait      bool           `json:"wait"`                                        // 是否等待资源 Ready
	Timeout   int            `json:"timeout"`                                     // 超时秒数
}

// UpgradeInput 升级 Release 入参。
type UpgradeInput struct {
	ChartVer string         `json:"chart_ver" binding:"required,min=1,max=64"` // 新的 Chart 版本
	Values   map[string]any `json:"values"`                                    // 新的 values (全量替换)
	Wait     bool           `json:"wait"`
	Timeout  int            `json:"timeout"`
}

// RollbackInput 回滚 Release 入参。
type RollbackInput struct {
	Version int  `json:"version"` // 回滚到哪个 revision, 0 = 上一个
	Wait    bool `json:"wait"`
	Timeout int  `json:"timeout"`
}

// ─────────────────────────────────────────────────────────────────────
// 方法实现
// ─────────────────────────────────────────────────────────────────────

// Install 安装一个新的 Helm Release。
//
// 实现步骤:
//   1. 校验: namespace+name 在该集群下唯一 (GetReleaseByName)
//   2. 查询 ChartRepo 获取 URL/Type/认证信息
//   3. 获取集群 kubeconfig: clusterProvider.GetKubeconfig(clusterID)
//   4. 构造 adapter.InstallInput, 调用 factory.Build(kubeconfig) → client.Install()
//   5. 成功: 构造 Release 实体 (status=deployed), repo.CreateRelease
//   6. 失败: 构造 Release 实体 (status=failed, message=err), repo.CreateRelease (保留记录便于排查)
//   7. 返回 Release 实体
func (s *ReleaseService) Install(ctx context.Context, in InstallInput) (*domain.Release, error) {
	log := logger.FromContext(ctx)
	log.Info("installing helm release",
		zap.String("name", in.Name),
		zap.String("namespace", in.Namespace),
		zap.String("chart", in.ChartName),
		zap.String("chart_ver", in.ChartVer),
	)

	// TODO 步骤 1: 唯一性校验
	// existing, err := s.repo.GetReleaseByName(ctx, in.Namespace, in.Name)
	// if err == nil && existing != nil {
	//     return nil, &domain.ErrReleaseAlreadyExists{Name: in.Name, Namespace: in.Namespace}
	// }

	// TODO 步骤 2: 查询 ChartRepo
	// chartRepo, err := s.repo.GetRepoByID(ctx, in.RepoID)
	// if err != nil { return nil, err }

	// TODO 步骤 3: 获取 rest.Config (复用 cluster 模块的解析逻辑)
	// restConfig, err := s.clusterRepo.GetRESTConfig(ctx, in.ClusterID)
	// if err != nil { return nil, err }

	// TODO 步骤 4: 调用 Helm SDK (Factory 接收 rest.Config, 不再传 kubeconfig)
	// client, err := s.factory.Build(ctx, restConfig, in.Namespace)
	// if err != nil { return nil, err }
	//
	// valuesJSON, _ := json.Marshal(in.Values)
	//
	// info, err := client.Install(ctx, adapter.InstallInput{
	//     ReleaseName: in.Name,
	//     Namespace:   in.Namespace,
	//     ChartName:   in.ChartName,
	//     ChartVer:    in.ChartVer,
	//     RepoURL:     chartRepo.URL,
	//     RepoType:    chartRepo.Type,
	//     Username:    chartRepo.Username,
	//     Password:    chartRepo.Password,
	//     Values:      in.Values,
	//     Timeout:     in.Timeout,
	//     Wait:        in.Wait,
	// })

	// TODO 步骤 5/6: 持久化 Release 记录
	// rel := &domain.Release{
	//     ID:         uuid.NewString(),
	//     Name:       in.Name,
	//     Namespace:  in.Namespace,
	//     ClusterID:  in.ClusterID,
	//     AppID:      in.AppID,
	//     EnvID:      in.EnvID,
	//     RepoID:     in.RepoID,
	//     ChartName:  in.ChartName,
	//     ChartVer:   in.ChartVer,
	//     AppVersion: info.AppVersion,
	//     Version:    1,
	//     Values:     string(valuesJSON),
	//     Status:     domain.ReleaseStatusDeployed,
	//     DeployedAt: &now,
	// }
	// if err != nil {
	//     rel.Status = domain.ReleaseStatusFailed
	//     rel.Message = err.Error()
	// }
	// s.repo.CreateRelease(ctx, rel)

	_ = ctx
	_ = in
	_ = json.Marshal
	_ = uuid.NewString
	return nil, nil
}

// Upgrade 升级一个已安装的 Release (Chart 版本或 values 变更)。
//
// 实现步骤:
//   1. GetReleaseByID 获取当前 Release
//   2. 校验当前 status 允许 upgrade (deployed / failed / superseded)
//   3. 查询 ChartRepo 获取 URL/Type/认证
//   4. 获取集群 kubeconfig
//   5. 调用 client.Upgrade()
//   6. 成功: 更新 Release (chart_ver/values/version+1/status=deployed)
//   7. 失败: 更新 status=failed, message=err
func (s *ReleaseService) Upgrade(ctx context.Context, id string, in UpgradeInput) (*domain.Release, error) {
	log := logger.FromContext(ctx)
	log.Info("upgrading helm release", zap.String("id", id), zap.String("chart_ver", in.ChartVer))

	// TODO: 实现
	_ = ctx
	_ = id
	_ = in
	return nil, nil
}

// Rollback 回滚到指定 revision。
//
// 实现步骤:
//   1. GetReleaseByID 获取当前 Release
//   2. 校验 status (只有 deployed / failed 允许 rollback)
//   3. 获取集群 kubeconfig
//   4. 调用 client.Rollback()
//   5. 成功: 更新 Release (version+1, status=deployed, message="rollback to revision X")
//   6. 失败: 更新 status=failed
//
// 注意: Rollback 不改变 chart_ver, 只是把 K8s 资源恢复到历史版本。
//
//	如果需要获取历史版本的 chart_ver, 需要查 Helm storage (secret/configmap)。
func (s *ReleaseService) Rollback(ctx context.Context, id string, in RollbackInput) (*domain.Release, error) {
	log := logger.FromContext(ctx)
	log.Info("rolling back helm release", zap.String("id", id), zap.Int("to_version", in.Version))

	// TODO: 实现
	_ = ctx
	_ = id
	_ = in
	return nil, nil
}

// Uninstall 卸载一个 Release。
//
// 实现步骤:
//   1. GetReleaseByID 获取当前 Release
//   2. 校验 status (已 uninstalled 的不允许再次卸载)
//   3. 获取集群 kubeconfig
//   4. 调用 client.Uninstall()
//   5. 更新 status=uninstalled
func (s *ReleaseService) Uninstall(ctx context.Context, id string, keepHistory bool) error {
	log := logger.FromContext(ctx)
	log.Info("uninstalling helm release", zap.String("id", id))

	// TODO: 实现
	_ = ctx
	_ = id
	_ = keepHistory
	return nil
}

// Get 根据 ID 查询 Release。
func (s *ReleaseService) Get(ctx context.Context, id string) (*domain.Release, error) {
	// TODO: return s.repo.GetReleaseByID(ctx, id)
	_ = ctx
	_ = id
	return nil, nil
}

// List 列表查询 Release。
func (s *ReleaseService) List(ctx context.Context, filter *domain.ReleaseListFilter) ([]*domain.Release, int64, error) {
	// TODO: return s.repo.ListReleases(ctx, filter)
	_ = ctx
	_ = filter
	return nil, 0, nil
}

// History 获取 Release 的版本历史 (从 Helm storage 拉取, 非 DB)。
//
// 实现步骤:
//   1. GetReleaseByID 获取 namespace + name + clusterID
//   2. 获取集群 kubeconfig
//   3. 调用 client.History()
//   4. 返回 []*adapter.ReleaseInfo
func (s *ReleaseService) History(ctx context.Context, id string) ([]*adapter.ReleaseInfo, error) {
	// TODO: 实现
	_ = ctx
	_ = id
	return nil, nil
}

// GetResources 获取 Release 部署的 K8s 资源列表。
//
// 实现步骤:
//   1. GetReleaseByID 获取 namespace + name + clusterID
//   2. 获取集群 kubeconfig
//   3. 调用 client.Get() → ReleaseInfo.Resources
func (s *ReleaseService) GetResources(ctx context.Context, id string) ([]adapter.ResourceInfo, error) {
	// TODO: 实现
	_ = ctx
	_ = id
	return nil, nil
}

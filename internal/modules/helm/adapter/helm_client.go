// Package adapter 把 Helm SDK 封装成业务接口, 业务代码 (service) 不直接 import helm.sh/helm/v3。
//
// 设计要点:
//
//   - HelmClient 接口暴露业务概念 (install/upgrade/rollback/uninstall),
//     不暴露 Helm SDK 类型 (避免业务代码传 *release.Release 之类的 SDK 对象)
//   - 接口方法都接收 context, 内部用 Helm action 包操作 K8s
//   - 真实实现在 helm/ 子包, 测试用 noop 在 noop/ 子包
//
// 依赖:
//
//	go get helm.sh/helm/v3@latest
//
// 核心 SDK 包:
//
//	helm.sh/helm/v3/pkg/action   — Install/Upgrade/Rollback/Uninstall/List 等操作
//	helm.sh/helm/v3/pkg/chart    — Chart 结构
//	helm.sh/helm/v3/pkg/release  — Release 结构
//	helm.sh/helm/v3/pkg/repo     — Repository 索引 (index.yaml)
//	helm.sh/helm/v3/pkg/cli      — EnvSettings (kubeconfig 等)
//	helm.sh/helm/v3/pkg/getter   — 下载 Chart (http/oci)
package adapter

import (
	"context"
	"olixops/internal/modules/helm/domain"

	"k8s.io/client-go/rest"
)

// ─────────────────────────────────────────────────────────────────────
// Helm 操作的入参/出参结构体 (业务层友好, 不暴露 SDK 类型)
// ─────────────────────────────────────────────────────────────────────

// InstallInput 是 Helm install 的入参。
//
// 注意: 不需要 Kubeconfig 字段, 由 Factory.Build(restConfig, namespace) 统一注入。
type InstallInput struct {
	ReleaseName string          // Helm release 名称
	Namespace   string          // 目标 namespace
	ChartName   string          // Chart 名称 (如 "bitnami/nginx")
	ChartVer    string          // Chart 版本 (如 "15.0.0"), 空则取最新
	RepoURL     string          // Chart 仓库 URL
	RepoType    domain.RepoType // 仓库类型 (http/oci)
	Username    string          // 仓库认证 (可选)
	Password    string          // 仓库认证 (可选)
	Values      map[string]any  // 用户自定义 values (会被 merge 到 Chart 默认 values 之上)
	Timeout     int             // 超时秒数, 0 则使用默认值 (300s)
	Wait        bool            // 是否等待所有资源 Ready
}

// UpgradeInput 是 Helm upgrade 的入参。
type UpgradeInput struct {
	ReleaseName string
	Namespace   string
	ChartName   string
	ChartVer    string
	RepoURL     string
	RepoType    domain.RepoType
	Username    string
	Password    string
	Values      map[string]any // 新的 values, 全量替换
	Timeout     int
	Wait        bool
	Install     bool // --install: release 不存在时自动 install
}

// RollbackInput 是 Helm rollback 的入参。
type RollbackInput struct {
	ReleaseName string
	Namespace   string
	Version     int // 回滚到哪个 revision, 0 = 回滚到上一个
	Timeout     int
	Wait        bool
}

// UninstallInput 是 Helm uninstall 的入参。
type UninstallInput struct {
	ReleaseName string
	Namespace   string
	Timeout     int
	KeepHistory bool // --keep-history: 保留 release 历史 (不彻底删除)
}

// ReleaseInfo 是 Helm release 的摘要信息 (从 SDK 转换而来)。
type ReleaseInfo struct {
	Name       string         `json:"name"`
	Namespace  string         `json:"namespace"`
	Revision   int            `json:"revision"`
	Status     string         `json:"status"`
	Chart      string         `json:"chart"`       // 如 "nginx-15.0.0"
	AppVersion string         `json:"app_version"` // 如 "1.25.3"
	Values     map[string]any `json:"values"`      // 合并后的最终 values
	Resources  []ResourceInfo `json:"resources"`   // 部署的 K8s 资源列表
	Notes      string         `json:"notes"`       // Chart NOTES.txt 渲染结果
	DeployedAt string         `json:"deployed_at"`
}

// ResourceInfo 是 Helm 部署后创建的 K8s 资源摘要。
type ResourceInfo struct {
	Kind      string `json:"kind"` // Deployment / Service / Ingress 等
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Status    string `json:"status"` // Ready / Pending / Failed
}

// ─────────────────────────────────────────────────────────────────────
// HelmClient 接口
// ─────────────────────────────────────────────────────────────────────

// HelmClient 是 Helm 模块与 Helm SDK 通信的抽象接口。
//
// 实现可以是真实的 (helm/ 子包, 调 Helm SDK) 或 noop (noop/ 子包, 测试用)。
// 业务代码永远不直接 import helm.sh/helm/v3, 只调本接口。
//
// 初始化时需要:
//   - kubeconfig (或 rest.Config) 用于连接目标集群
//   - Helm action.Configuration (内部持有 K8s client + storage)
//
// 每个方法对应一个 Helm CLI 命令:
//
//	Install   → helm install
//	Upgrade   → helm upgrade
//	Rollback  → helm rollback
//	Uninstall → helm uninstall
//	List      → helm list
//	Get       → helm get (values / manifest / notes)
//	History   → helm history
type HelmClient interface {

	// ─────────────────────────────────────────────────────────────────
	// Release 生命周期操作
	// ─────────────────────────────────────────────────────────────────

	// Install 安装一个 Helm Chart。
	//
	// 实现步骤:
	//   1. 构造 action.Configuration (用 kubeconfig + namespace)
	//   2. 构造 action.NewInstall(cfg)
	//   3. 设置 ReleaseName / Namespace / Version / Wait / Timeout
	//   4. 调用 LocateChart (从 RepoURL 下载 Chart)
	//   5. 调用 LoadChart (加载 Chart)
	//   6. merge values
	//   7. install.Run(chart, values)
	Install(ctx context.Context, in InstallInput) (*ReleaseInfo, error)

	// Upgrade 升级一个已安装的 Release。
	//
	// 实现步骤:
	//   1. action.NewUpgrade(cfg)
	//   2. 设置 Version / Wait / Timeout / Install (如果 --install)
	//   3. LocateChart + LoadChart
	//   4. upgrade.Run(releaseName, chart, values)
	Upgrade(ctx context.Context, in UpgradeInput) (*ReleaseInfo, error)

	// Rollback 回滚到指定 revision。
	//
	// 实现步骤:
	//   1. action.NewRollback(cfg)
	//   2. 设置 Version / Wait / Timeout
	//   3. rollback.Run(releaseName)
	Rollback(ctx context.Context, in RollbackInput) error

	// Uninstall 卸载一个 Release。
	//
	// 实现步骤:
	//   1. action.NewUninstall(cfg)
	//   2. 设置 KeepHistory / Timeout
	//   3. uninstall.Run(releaseName)
	Uninstall(ctx context.Context, in UninstallInput) error

	// ─────────────────────────────────────────────────────────────────
	// 查询操作
	// ─────────────────────────────────────────────────────────────────

	// List 列出当前集群指定 namespace 下的所有 Release。
	//
	// 实现步骤:
	//   1. action.NewList(cfg)
	//   2. 设置 Namespace / All (是否包含已卸载的)
	//   3. list.Run()
	//
	// 注意: 不需要传 kubeconfig, Factory.Build 时已注入 rest.Config。
	List(ctx context.Context, namespace string) ([]*ReleaseInfo, error)

	// Get 获取单个 Release 的详细信息。
	//
	// 实现步骤:
	//   1. action.NewGet(cfg)
	//   2. get.Run(releaseName)
	Get(ctx context.Context, releaseName, namespace string) (*ReleaseInfo, error)

	// History 获取 Release 的版本历史 (所有 revision)。
	//
	// 实现步骤:
	//   1. action.NewHistory(cfg)
	//   2. 设置 Max (最大返回条数, 默认 256)
	//   3. history.Run(releaseName)
	History(ctx context.Context, releaseName, namespace string, max int) ([]*ReleaseInfo, error)

	// ─────────────────────────────────────────────────────────────────
	// Chart 仓库操作
	// ─────────────────────────────────────────────────────────────────

	// FetchIndex 从仓库 URL 拉取并解析 index.yaml, 返回所有 Chart 元数据。
	//
	// 实现步骤 (HTTP 仓库):
	//   1. 构造 repo.Entry{Name, URL, Username, Password}
	//   2. 用 getter.All(settings) 下载 URL/index.yaml
	//   3. repo.LoadIndexFile 解析
	//   4. 转换为 []domain.ChartMeta
	//
	// OCI 仓库: 需要用 oras 协议, 暂不支持 List, 只能按 Chart 名 pull。
	FetchIndex(ctx context.Context, repoURL string, repoType domain.RepoType, username, password string) ([]*domain.ChartMeta, error)

	// FetchChartVersions 获取某个 Chart 的所有可用版本。
	//
	// 实现步骤:
	//   1. 先 FetchIndex
	//   2. 在结果中按 chartName 过滤
	//   3. 转换为 []domain.ChartVersion
	FetchChartVersions(ctx context.Context, repoURL string, repoType domain.RepoType, username, password, chartName string) ([]*domain.ChartVersion, error)

	// RenderValues 获取某个 Chart 的默认 values.yaml (渲染前)。
	//
	// 实现步骤:
	//   1. 下载并加载 Chart (LocateChart + LoadChart)
	//   2. 返回 chart.Values (map[string]any)
	//
	// 用途: 让用户在安装前看到 Chart 支持哪些可配置参数。
	RenderValues(ctx context.Context, repoURL string, repoType domain.RepoType, username, password, chartName, chartVer string) (map[string]any, error)
}

// ─────────────────────────────────────────────────────────────────────
// Factory 接口
// ─────────────────────────────────────────────────────────────────────

// Factory 构造 HelmClient。
//
// 每次操作可能面向不同集群, 所以 Factory 负责按需构造。
// 入参使用 *rest.Config, 与 cluster 模块的 K8sClient 共享同一份配置,
// 避免 Helm 模块重复解析 kubeconfig。
type Factory interface {
	// Build 根据已有的 rest.Config 构造一个 HelmClient。
	//
	// 内部步骤:
	//   1. 用 rest.Config 构造 RESTClientGetter
	//   2. 构造 action.Configuration (RESTClientGetter + storage)
	//   3. 返回包装好的 HelmClient
	//
	// rest.Config 由 cluster 模块的 Factory 统一解析 kubeconfig 后提供,
	// Helm 模块不需要关心 kubeconfig 的来源和加解密。
	Build(ctx context.Context, restConfig *rest.Config, namespace string) (HelmClient, error)
}

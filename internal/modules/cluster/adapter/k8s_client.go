// Package adapter 把 K8s SDK 封装成业务接口, 业务代码 (service) 不直接 import k8s.io/...。
//
// 设计要点:
//
//   - K8sClient 接口暴露业务概念 (namespace / node / workload),
//     不暴露 k8s.io 类型 (避免业务代码传 *corev1.Namespace 之类的 SDK 对象)
//   - 接口方法都接收 context, 内部用 client-go 的 rest.Config 调 apiserver
//   - 真实实现在 k8s/ 子包, 测试用 noop 在 noop/ 子包
package adapter

import (
	"context"
	"time"

	"olixops/internal/modules/cluster/domain"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

// K8sClient 是 cluster 模块与 K8s apiserver 通信的抽象接口。
//
// 实现可以是真实的 (k8s/ 子包, 调 client-go) 或 noop (noop/ 子包, 测试用)。
// 业务代码永远不直接 import k8s.io/client-go, 只调本接口。
type K8sClient interface {
	GetRESTConfig() *rest.Config

	// Probe 探测集群健康, 返回 (apiserver 版本, 节点数, 错误)。
	// 用于 Cluster.Register / Cluster.Refresh。
	Probe(ctx context.Context) (version string, nodeCount int, err error)

	// ListNamespaces 列出 namespace。
	ListNamespaces(ctx context.Context, listOptions metav1.ListOptions) ([]domain.Namespace, error)

	// GetNamespace 取单个 namespace。
	GetNamespace(ctx context.Context, name string) (*domain.Namespace, error)

	// CreateNamespace 创建 namespace。
	CreateNamespace(ctx context.Context, name string, labels map[string]string) (*domain.Namespace, error)

	// DeleteNamespace 删除 namespace, timeout 是等待 Terminating 完成的最大时长。
	DeleteNamespace(ctx context.Context, name string, timeout time.Duration) error

	// ListNodes 列出节点。
	ListNodes(ctx context.Context) ([]domain.Node, error)

	// GetNode 取单个节点。
	GetNode(ctx context.Context, name string) (*domain.Node, error)

	// ListDeployments 列出 deployment, 可选按 namespace 过滤 (空字符串 = 全部)。
	ListDeployments(ctx context.Context, namespace string) ([]domain.Deployment, error)

	// GetDeployment 取单个 deployment。
	GetDeployment(ctx context.Context, namespace, name string) (*domain.Deployment, error)

	// ListServices 列出 service。
	ListServices(ctx context.Context, namespace string) ([]domain.Service, error)

	// GetService 取单个 service。
	GetService(ctx context.Context, namespace, name string) (*domain.Service, error)

	// ListIngresses 列出 ingress。
	ListIngresses(ctx context.Context, namespace string) ([]domain.Ingress, error)

	// GetIngress 取单个 ingress。
	GetIngress(ctx context.Context, namespace, name string) (*domain.Ingress, error)
}

// Factory 构造 K8sClient (按需实现, 避免业务层直接 new)。
type Factory interface {
	// Build 根据 base64 编码的 kubeconfig 构造 K8sClient。
	// 每次调用都返回新实例, 因为 clientset 内部状态不共享。
	Build(ctx context.Context, cluster *domain.Cluster) (K8sClient, error)

	GetK8sClient(ctx context.Context, clientId string) (K8sClient, error)
}

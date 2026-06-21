// Package noop 提供 K8sClient 的 noop 实现, 用于单元测试和本地 dev 环境。
//
// 返回固定假数据, 不调任何 K8s SDK。
// 业务代码用这个 noop 跑测试时不需要真集群。
package noop

import (
	"context"
	"fmt"
	"sync"
	"time"

	"olixops/internal/modules/cluster/adapter"
	"olixops/internal/modules/cluster/domain"
)

// NoopK8sClient 是 adapter.K8sClient 的 noop 实现。
//
// 特性:
//   - List* 方法返回预置数据 (可改)
//   - Create/Delete 维护内部 map 模拟 K8s 状态
//   - Probe 永远返回 v1.28.0 / 3 个节点
//   - 所有方法线程安全 (sync.Mutex)
type NoopK8sClient struct {
	mu          sync.Mutex
	namespaces  map[string]*domain.Namespace
	nodes       map[string]*domain.Node
	deployments map[string]*domain.Deployment // key = ns/name
	services    map[string]*domain.Service
	ingresses   map[string]*domain.Ingress
}

// New 构造 noop client, 预置一些假数据方便测试。
func New() *NoopK8sClient {
	c := &NoopK8sClient{
		namespaces:  make(map[string]*domain.Namespace),
		nodes:       make(map[string]*domain.Node),
		deployments: make(map[string]*domain.Deployment),
		services:    make(map[string]*domain.Service),
		ingresses:   make(map[string]*domain.Ingress),
	}
	// 预置一些测试数据
	c.namespaces["default"] = &domain.Namespace{
		Name: "default", Status: domain.NamespaceStatusActive,
		CreatedAt: time.Now(),
	}
	c.namespaces["kube-system"] = &domain.Namespace{
		Name: "kube-system", Status: domain.NamespaceStatusActive,
		CreatedAt: time.Now(),
	}
	c.nodes["node-1"] = &domain.Node{
		Name: "node-1", Role: domain.NodeRoleMaster, Status: domain.NodeStatusReady,
		InternalIP: "10.0.0.1", OSImage: "Ubuntu 22.04", KubeletVersion: "v1.28.0",
		CapacityCPU: "4", CapacityMemory: "8Gi",
		CreatedAt: time.Now(),
	}
	return c
}

// 编译期断言: NoopK8sClient 必须实现 adapter.K8sClient。
var _ adapter.K8sClient = (*NoopK8sClient)(nil)

// TODO: 实现 K8sClient 接口的所有 14 个方法。
// 参考 adapter/k8s_client.go 的接口定义, 用 c.mu Lock/Unlock 保护 c.namespaces 等 map。
// 例:
//
//	func (c *NoopK8sClient) Probe(ctx context.Context) (string, int, error) {
//	    return "v1.28.0-noop", 3, nil
//	}

func (c *NoopK8sClient) Probe(ctx context.Context) (version string, nodeCount int, err error) {
	// TODO
	_ = ctx
	return "v1.28.0-noop", 0, fmt.Errorf("not implemented")
}

func (c *NoopK8sClient) ListNamespaces(ctx context.Context) ([]domain.Namespace, error) {
	// TODO
	_ = ctx
	return nil, fmt.Errorf("not implemented")
}

func (c *NoopK8sClient) GetNamespace(ctx context.Context, name string) (*domain.Namespace, error) {
	// TODO
	_ = ctx
	_ = name
	return nil, fmt.Errorf("not implemented")
}

func (c *NoopK8sClient) CreateNamespace(ctx context.Context, name string, labels map[string]string) (*domain.Namespace, error) {
	// TODO
	_ = ctx
	_ = name
	_ = labels
	return nil, fmt.Errorf("not implemented")
}

func (c *NoopK8sClient) DeleteNamespace(ctx context.Context, name string, timeout time.Duration) error {
	// TODO
	_ = ctx
	_ = name
	_ = timeout
	return fmt.Errorf("not implemented")
}

func (c *NoopK8sClient) ListNodes(ctx context.Context) ([]domain.Node, error) {
	// TODO
	_ = ctx
	return nil, fmt.Errorf("not implemented")
}

func (c *NoopK8sClient) GetNode(ctx context.Context, name string) (*domain.Node, error) {
	// TODO
	_ = ctx
	_ = name
	return nil, fmt.Errorf("not implemented")
}

func (c *NoopK8sClient) ListDeployments(ctx context.Context, namespace string) ([]domain.Deployment, error) {
	// TODO
	_ = ctx
	_ = namespace
	return nil, fmt.Errorf("not implemented")
}

func (c *NoopK8sClient) GetDeployment(ctx context.Context, namespace, name string) (*domain.Deployment, error) {
	// TODO
	_ = ctx
	_ = namespace
	_ = name
	return nil, fmt.Errorf("not implemented")
}

func (c *NoopK8sClient) ListServices(ctx context.Context, namespace string) ([]domain.Service, error) {
	// TODO
	_ = ctx
	_ = namespace
	return nil, fmt.Errorf("not implemented")
}

func (c *NoopK8sClient) GetService(ctx context.Context, namespace, name string) (*domain.Service, error) {
	// TODO
	_ = ctx
	_ = namespace
	_ = name
	return nil, fmt.Errorf("not implemented")
}

func (c *NoopK8sClient) ListIngresses(ctx context.Context, namespace string) ([]domain.Ingress, error) {
	// TODO
	_ = ctx
	_ = namespace
	return nil, fmt.Errorf("not implemented")
}

func (c *NoopK8sClient) GetIngress(ctx context.Context, namespace, name string) (*domain.Ingress, error) {
	// TODO
	_ = ctx
	_ = namespace
	_ = name
	return nil, fmt.Errorf("not implemented")
}

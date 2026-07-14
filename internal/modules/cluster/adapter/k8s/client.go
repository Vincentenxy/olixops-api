package k8s

import (
	"context"
	"fmt"
	"olixops/internal/config"
	"olixops/internal/modules/cluster/repository"
	"olixops/internal/platform/logger"
	"olixops/pkg/cryptox"
	"sync"
	"time"

	"olixops/internal/modules/cluster/adapter"
	"olixops/internal/modules/cluster/domain"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// K8sClient 基于 client-go 实现
type K8sClientInstance struct {
	clientSet *kubernetes.Clientset
	restCfg   *rest.Config
}

// 编译期断言: K8sClientInstance 必须实现 adapter.K8sClient。
var _ adapter.K8sClient = (*K8sClientInstance)(nil)

// Factory 是 K8sClient 的工厂
type factory struct {
	repo        repository.ClusterRepo
	clientMap   map[string]*K8sClientInstance
	once        sync.Once
	initErr     error
	mu          sync.RWMutex
	logger      *zap.Logger
	k8sConfig   *config.K8sConfig
	initialized bool
}

// 编译期断言
var _ adapter.Factory = (*factory)(nil)

func (f *factory) Build(ctx context.Context, cluster *domain.Cluster) (adapter.K8sClient, error) {
	f.logger.Info("building k8s client")

	aesgcm, err := cryptox.DecryptAESGCM([]byte(f.k8sConfig.SecretKey), cluster.KubeConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt AESGCM: %w", err)
	}

	tcc, err := clientcmd.NewClientConfigFromBytes(aesgcm)
	if err != nil {
		return nil, fmt.Errorf("load kubeconfig failed: %w", err)
	}

	restConfig, err := tcc.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("build rest config failed: %w", err)
	}

	// create clientset
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("create clientset failed: %w", err)
	}

	instance := K8sClientInstance{
		clientSet: clientset,
		restCfg:   restConfig,
	}
	f.clientMap[cluster.ID] = &instance

	return &instance, nil
}

func (f *factory) GetK8sClient(ctx context.Context, clientId string) (adapter.K8sClient, error) {

	// get from amp
	f.mu.RLock()
	client, ok := f.clientMap[clientId]
	f.mu.RUnlock()

	if ok {
		return client, nil
	}

	// create
	f.mu.Lock()
	defer f.mu.Unlock()

	// double check, do not multiple creation
	if client, ok = f.clientMap[clientId]; ok {
		return client, nil
	}

	// get k8s config from db
	cluster, err2 := f.repo.GetByID(ctx, clientId)
	if err2 != nil {
		return nil, fmt.Errorf("client %s not found in db: %w", clientId, err2)
	}

	// create client
	curClient, err2 := f.Build(ctx, cluster)
	if err2 != nil {
		f.logger.Warn("build k8s client failed",
			zap.String("clientId", clientId),
			zap.Error(err2),
		)
		return nil, err2
	}

	return curClient, nil
}

func NewFactory(ctx context.Context, k8sConfig *config.K8sConfig, repo repository.ClusterRepo) adapter.Factory {
	l := logger.L()

	clusterMap := make(map[string]*K8sClientInstance)
	factoryIns := factory{
		repo:      repo,
		clientMap: clusterMap,
		logger:    l,
		k8sConfig: k8sConfig,
	}
	factoryIns.asyncLoadAllClients(ctx)
	return &factoryIns
}

// asyncLoadAllClients 异步加载所有客户端
func (f *factory) asyncLoadAllClients(ctx context.Context) {
	f.once.Do(func() {
		bgCtx := context.Background()

		list, total, err := f.repo.List(bgCtx, nil)

		f.mu.Lock()
		defer f.mu.Unlock()

		if err != nil {
			f.logger.Error("load cluster list failed", zap.Error(err))
			f.initialized = true
			return
		}

		if total == 0 {
			f.logger.Warn("cluster list is empty, no clients to init")
			f.initialized = true
			return
		}

		// create all clients
		for _, ci := range list {
			_, err := f.Build(bgCtx, ci)
			if err != nil {
				f.logger.Error("build k8s client failed",
					zap.String("clusterID", ci.ID),
					zap.Error(err))
				continue
			}
		}

		f.logger.Info("all k8s clients loaded",
			zap.Int("total", len(f.clientMap)),
			zap.Int("expected", len(list)))
		f.initialized = true
	})
}

func (f *factory) GetRESTConfig(ctx context.Context, clusterID string) (*rest.Config, error) {
	client, err := f.GetK8sClient(ctx, clusterID)
	if err != nil {
		logger.L().Error("get k8s client failed",
			zap.Error(err),
			zap.String("clusterID", clusterID),
		)
		return nil, err
	}

	return client.GetRESTConfig(), nil
}

// 获取config配置
func (c *K8sClientInstance) GetRESTConfig() *rest.Config {
	return c.restCfg
}

func (c *K8sClientInstance) Probe(ctx context.Context) (version string, nodeCount int, err error) {
	// TODO: 调 c.clientset.Discovery().ServerVersion() + c.clientset.CoreV1().Nodes().List
	_ = ctx
	return "", 0, fmt.Errorf("not implemented")
}

func (c *K8sClientInstance) ListNamespaces(ctx context.Context, listOptions metav1.ListOptions) ([]domain.Namespace, error) {
	// 添加上对应的选项 如：labelSelector
	list, err := c.clientSet.CoreV1().Namespaces().List(ctx, listOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %w", err)
	}

	namespaces := make([]domain.Namespace, 0, len(list.Items))
	for _, ns := range list.Items {
		namespaces = append(namespaces, domain.Namespace{
			Name:        ns.Name,
			UID:         string(ns.UID),
			Status:      domain.NamespaceStatus(ns.Status.Phase),
			Labels:      ns.Labels,
			Annotations: ns.Annotations,
			// CreatedAt:   ns.CreationTimestamp.Time,
		})
	}

	return namespaces, nil
}

func (c *K8sClientInstance) GetNamespace(ctx context.Context, name string) (*domain.Namespace, error) {
	// TODO
	_ = ctx
	_ = name
	return nil, fmt.Errorf("not implemented")
}

func (c *K8sClientInstance) CreateNamespace(ctx context.Context, name string, labels map[string]string) (*domain.Namespace, error) {
	// TODO: clientset.CoreV1().Namespaces().Create(ctx, &corev1.Namespace{...})
	_ = ctx
	_ = name
	_ = labels
	return nil, fmt.Errorf("not implemented")
}

func (c *K8sClientInstance) DeleteNamespace(ctx context.Context, name string, timeout time.Duration) error {
	// TODO: 调 Delete, 然后用 PollUntilContextTimeout 等待 Terminating 结束
	_ = ctx
	_ = name
	_ = timeout
	return fmt.Errorf("not implemented")
}

func (c *K8sClientInstance) ListNodes(ctx context.Context) ([]domain.Node, error) {
	if c.clientSet == nil {
		return nil, fmt.Errorf("clientSet is nil")
	}

	nodes, err := c.clientSet.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("list nodes failed: %w", err)
	}

	result := make([]domain.Node, 0, len(nodes.Items))
	for _, node := range nodes.Items {
		result = append(result, c.convertToDomainNode(node))
	}

	return result, nil
}

func (c *K8sClientInstance) convertToDomainNode(node corev1.Node) domain.Node {
	return domain.Node{
		Name:      node.Name,
		Status:    getNodeStatus(node),
		Roles:     getNodeRoles(node),
		Version:   node.Status.NodeInfo.KubeletVersion,
		OSImage:   node.Status.NodeInfo.OSImage,
		Kernel:    node.Status.NodeInfo.KernelVersion,
		Container: node.Status.NodeInfo.ContainerRuntimeVersion,
		CPUCores:  getNodeCPU(node),
		Memory:    getNodeMemory(node),
		Pods:      getNodePods(node),
		CreatedAt: node.CreationTimestamp.Time,
		Labels:    node.Labels,
		Taints:    getNodeTaints(node),
	}
}

// getNodeStatus 获取节点状态
func getNodeStatus(node corev1.Node) string {
	for _, condition := range node.Status.Conditions {
		if condition.Type == corev1.NodeReady {
			if condition.Status == corev1.ConditionTrue {
				return "Ready"
			}
			return "NotReady"
		}
	}
	return "Unknown"
}

// getNodeRoles 获取节点角色
func getNodeRoles(node corev1.Node) []string {
	var roles []string
	for label := range node.Labels {
		if label == "node-role.kubernetes.io/control-plane" ||
			label == "node-role.kubernetes.io/master" {
			roles = append(roles, "control-plane")
		}
		if label == "node-role.kubernetes.io/worker" {
			roles = append(roles, "worker")
		}
	}
	// 如果没有任何角色标签，默认为 worker
	if len(roles) == 0 {
		roles = append(roles, "worker")
	}
	return roles
}

// getNodeCPU 获取节点 CPU 核数
func getNodeCPU(node corev1.Node) int64 {
	if cpu, ok := node.Status.Allocatable[corev1.ResourceCPU]; ok {
		return cpu.MilliValue() / 1000 // 从毫核转换为核
	}
	return 0
}

// getNodeMemory 获取节点内存（GB）
func getNodeMemory(node corev1.Node) int64 {
	if mem, ok := node.Status.Allocatable[corev1.ResourceMemory]; ok {
		return mem.Value() / (1024 * 1024 * 1024) // 转换为 GB
	}
	return 0
}

// getNodePods 获取节点 Pod 容量
func getNodePods(node corev1.Node) int64 {
	if pods, ok := node.Status.Allocatable[corev1.ResourcePods]; ok {
		return pods.Value()
	}
	return 0
}

// getNodeTaints 获取节点污点
func getNodeTaints(node corev1.Node) []string {
	var taints []string
	for _, taint := range node.Spec.Taints {
		taints = append(taints, fmt.Sprintf("%s=%s:%s", taint.Key, taint.Value, taint.Effect))
	}
	return taints
}

// getNodeIP 获取节点 IP
func getNodeIP(node corev1.Node) string {
	for _, addr := range node.Status.Addresses {
		if addr.Type == corev1.NodeInternalIP {
			return addr.Address
		}
	}
	for _, addr := range node.Status.Addresses {
		if addr.Type == corev1.NodeExternalIP {
			return addr.Address
		}
	}
	return ""
}

// getNodeHostname 获取节点主机名
func getNodeHostname(node corev1.Node) string {
	for _, addr := range node.Status.Addresses {
		if addr.Type == corev1.NodeHostName {
			return addr.Address
		}
	}
	return ""
}

func (c *K8sClientInstance) GetNode(ctx context.Context, name string) (*domain.Node, error) {
	// TODO
	_ = ctx
	_ = name
	return nil, fmt.Errorf("not implemented")
}

func (c *K8sClientInstance) ListDeployments(ctx context.Context, namespace string) ([]domain.Deployment, error) {
	// TODO: clientset.AppsV1().Deployments(namespace).List
	_ = ctx
	_ = namespace
	return nil, fmt.Errorf("not implemented")
}

func (c *K8sClientInstance) GetDeployment(ctx context.Context, namespace, name string) (*domain.Deployment, error) {
	// TODO
	_ = ctx
	_ = namespace
	_ = name
	return nil, fmt.Errorf("not implemented")
}

func (c *K8sClientInstance) ListServices(ctx context.Context, namespace string) ([]domain.Service, error) {
	// TODO
	_ = ctx
	_ = namespace
	return nil, fmt.Errorf("not implemented")
}

func (c *K8sClientInstance) GetService(ctx context.Context, namespace, name string) (*domain.Service, error) {
	// TODO
	_ = ctx
	_ = namespace
	_ = name
	return nil, fmt.Errorf("not implemented")
}

func (c *K8sClientInstance) ListIngresses(ctx context.Context, namespace string) ([]domain.Ingress, error) {
	// TODO
	_ = ctx
	_ = namespace
	return nil, fmt.Errorf("not implemented")
}

func (c *K8sClientInstance) GetIngress(ctx context.Context, namespace, name string) (*domain.Ingress, error) {
	// TODO
	_ = ctx
	_ = namespace
	_ = name
	return nil, fmt.Errorf("not implemented")
}

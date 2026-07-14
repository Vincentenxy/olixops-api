package domain

import "time"

// NodeRole 节点角色。
type NodeRole string

const (
	NodeRoleMaster NodeRole = "master" // control-plane
	NodeRoleWorker NodeRole = "worker"
)

// NodeStatus 节点状态。
type NodeStatus string

const (
	NodeStatusReady    NodeStatus = "Ready"
	NodeStatusNotReady NodeStatus = "NotReady"
)

// Node K8s 节点的本地视图 (只读, 不入库)。
type Node struct {
	Name         string            `json:"name"`
	Status       string            `json:"status"`    // Ready/NotReady/Unknown
	Roles        []string          `json:"roles"`     // control-plane, worker
	Version      string            `json:"version"`   // v1.28.0
	OSImage      string            `json:"osImage"`   // Ubuntu 22.04 LTS
	Kernel       string            `json:"kernel"`    // 5.15.0-91-generic
	Container    string            `json:"container"` // containerd://1.7.6
	CPUCores     int64             `json:"cpuCores"`  // 8
	Memory       int64             `json:"memory"`    // 16 (GB)
	Pods         int64             `json:"pods"`      // 110
	CreatedAt    time.Time         `json:"createdAt"`
	Labels       map[string]string `json:"labels"`
	Taints       []string          `json:"taints"`
	IPAddress    string            `json:"ipAddress"`
	Hostname     string            `json:"hostname"`
	Architecture string            `json:"architecture"` // amd64, arm64
}

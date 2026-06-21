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
	Name              string            `json:"name"`
	Role              NodeRole          `json:"role"`
	Status            NodeStatus        `json:"status"`
	InternalIP        string            `json:"internalIP"`
	OSImage           string            `json:"osImage"`
	KubeletVersion    string            `json:"kubeletVersion"`
	CapacityCPU       string            `json:"capacityCPU"`    // 例如 "4"
	CapacityMemory    string            `json:"capacityMemory"` // 例如 "8Gi"
	AllocatableCPU    string            `json:"allocatableCPU"`
	AllocatableMemory string            `json:"allocatableMemory"`
	Labels            map[string]string `json:"labels,omitempty"`
	CreatedAt         time.Time         `json:"createdAt"`
}

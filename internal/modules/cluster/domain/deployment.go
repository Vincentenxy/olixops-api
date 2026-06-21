package domain

import "time"

// Deployment K8s Deployment 的本地视图 (只读)。
type Deployment struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Replicas  int32             `json:"replicas"`  // 期望副本
	Ready     int32             `json:"ready"`     // 就绪副本
	Available int32             `json:"available"` // 可用副本
	Images    []string          `json:"images"`    // 使用的镜像
	Labels    map[string]string `json:"labels,omitempty"`
	CreatedAt time.Time         `json:"createdAt"`
}

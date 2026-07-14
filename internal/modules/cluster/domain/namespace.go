package domain

import "time"

// NamespaceStatus namespace 状态。
type NamespaceStatus string

const (
	NamespaceStatusActive      NamespaceStatus = "Active"
	NamespaceStatusTerminating NamespaceStatus = "Terminating"
)

// Namespace K8s namespace 的本地视图。
//
// 注意: 本实体不是从 K8s 同步过来的缓存, 而是从 adapter 实时查询的结果。
// 集群删除时不需要清理本表的 namespace 记录 (因为没有 namespace 表)。
type Namespace struct {
	Name          string            `json:"name"`
	Status        NamespaceStatus   `json:"status"`
	UID           string            `json:"uid"`
	Labels        map[string]string `json:"labels,omitempty"`
	Annotations   map[string]string `json:"annotations,omitempty"`
	CreatedAt     time.Time         `json:"createdAt"`
	ResourceQuota *ResourceQuota    `json:"resourceQuota,omitempty"`
}

// ResourceQuota namespace 的资源配额 (来自 K8s ResourceQuota 对象)。
type ResourceQuota struct {
	Hard map[string]string `json:"hard,omitempty"` // 例如 {"cpu": "4", "memory": "8Gi"}
	Used map[string]string `json:"used,omitempty"` // 实际使用
}

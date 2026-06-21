package domain

// ServicePort 服务端口。
type ServicePort struct {
	Name     string `json:"name"`
	Port     int32  `json:"port"`
	Protocol string `json:"protocol"` // TCP / UDP
	NodePort int32  `json:"nodePort,omitempty"`
}

// Service K8s Service 的本地视图 (只读)。
type Service struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Type      string            `json:"type"` // ClusterIP / NodePort / LoadBalancer
	ClusterIP string            `json:"clusterIP"`
	Ports     []ServicePort     `json:"ports"`
	Selector  map[string]string `json:"selector,omitempty"`
}

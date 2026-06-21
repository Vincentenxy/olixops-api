package domain

// IngressTLS Ingress 的 TLS 配置。
type IngressTLS struct {
	Hosts      []string `json:"hosts"`
	SecretName string   `json:"secretName"`
}

// IngressPath Ingress 的一条路径规则。
type IngressPath struct {
	Path     string `json:"path"`
	PathType string `json:"pathType"` // Prefix / Exact / ImplementationSpecific
	Backend  string `json:"backend"`  // service.name:port 形式
}

// Ingress K8s Ingress 的本地视图 (只读)。
type Ingress struct {
	Name      string        `json:"name"`
	Namespace string        `json:"namespace"`
	ClassName string        `json:"className,omitempty"` // nginx / traefik 等
	Hosts     []string      `json:"hosts"`
	Paths     []IngressPath `json:"paths"`
	TLS       []IngressTLS  `json:"tls,omitempty"`
}

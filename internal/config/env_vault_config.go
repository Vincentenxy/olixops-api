package config

import "time"

// Config 描述 envVault 连接参数.
//
// 字段:
//
//	APIURL      envVault 服务地址, 例如 https://vault.internal
//	APIToken    静态 API token (从环境变量 / K8s Secret 注入)
//	SecretPath  secret 路径, 例如 "olixops/db-encryption-key"
//	Timeout     HTTP 超时, 默认 10s
type EnvVaultConfig struct {
	URL     string        `mapstructure:"url"`
	Timeout time.Duration `mapstructure:"timeout"`
}

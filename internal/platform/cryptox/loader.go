package cryptox

import (
	"context"
	"encoding/hex"
	"fmt"
	"os"
)

// KeyLoader 负责加载 DEK (Data Encryption Key, 32 字节).
//
// 不同环境用不同实现:
//   - dev: EnvLoader (从 OLIXOPS_DB_ENCRYPTION_KEY 环境变量读)
//   - prod: envVault 子包 (HTTP / SDK 调 envVault)
type KeyLoader interface {
	// LoadKey 返回 32 字节 DEK.
	LoadKey(ctx context.Context) ([]byte, error)
}

// EnvLoader 从环境变量读 key, 用于 dev / 单机部署.
type EnvLoader struct{}

// LoadKey 实现 KeyLoader.
func (EnvLoader) LoadKey(_ context.Context) ([]byte, error) {
	raw := os.Getenv("OLIXOPS_DB_ENCRYPTION_KEY")
	if raw == "" {
		return nil, fmt.Errorf("OLIXOPS_DB_ENCRYPTION_KEY is empty")
	}
	// TODO: 支持两种格式: hex (64 字符) / 原始字符串 (32 字节)
	// 推荐 hex, 容易在 shell / k8s secret 里转义
	if len(raw) == 64 {
		key, err := hex.DecodeString(raw)
		if err != nil {
			return nil, fmt.Errorf("OLIXOPS_DB_ENCRYPTION_KEY is not valid hex: %w", err)
		}
		return key, nil
	}
	// 兼容: 直接当 32 字节原始字符串
	if len(raw) != 32 {
		return nil, fmt.Errorf("OLIXOPS_DB_ENCRYPTION_KEY must be 32 bytes (or 64 hex chars), got %d", len(raw))
	}
	return []byte(raw), nil
}

// NewServiceFromEnv 是 dev 场景的便利构造函数, 等价于 New(EnvLoader{}.LoadKey()).
//
// TODO: 把这个函数整合进 app.go 启动流程, 优先用 envVault loader, 降级到 EnvLoader.
func NewServiceFromEnv(ctx context.Context) (*Service, error) {
	key, err := EnvLoader{}.LoadKey(ctx)
	if err != nil {
		return nil, err
	}
	return New(key)
}

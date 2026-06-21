// Package envvault 集成自建 envVault, 用于生产环境加载 DEK (Data Encryption Key).
//
// 本包只走 HTTP REST API 方式调 envVault (不依赖 SDK, 跨语言通用).
//
// 集成方式:
//
//	cfg := envvault.Config{
//	    APIURL:     os.Getenv("ENVVAULT_URL"),
//	    APIToken:   os.Getenv("ENVVAULT_TOKEN"),       // 静态 token
//	    SecretPath: "olixops/db-encryption-key",     // 路径
//	    Timeout:    10 * time.Second,                  // 可选, 默认 10s
//	}
//	loader, err := envvault.NewLoader(ctx, cfg)
//	key, err := loader.LoadKey(ctx)
//	svc, err := cryptox.New(key)
//
// envVault API 约定 (以自建服务文档为准, 可能要微调):
//
//	GET {APIURL}/v1/secret/data/{SecretPath}
//	Header X-Vault-Token: {APIToken}
//	Response: { "data": { "data": { "value": "<hex-key>" } } }
//
// 未来要换 SDK, 在 httpClient 旁加个 sdkClient 即可, Loader 不动.
package envvault

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"olixops/internal/config"
	ctx2 "olixops/pkg/ctx"
	"olixops/pkg/httpx"
)

const (
	apiSecretsList = "/api/v1/secrets/list"
)

// Loader 是 envVault 的 KeyLoader 实现.
type Loader struct {
	cfg        config.EnvVaultConfig
	httpClient *httpClient // 当前只支持 HTTP; 未来加 SDK 时换 interface
}

// NewLoader 构造 envVault loader.
//
// 流程:
//
//  1. 校验 cfg (APIURL / APIToken / SecretPath 必填)
//  2. cfg.Timeout < 1s 时设默认 10s
//  3. 构造 httpClient
func NewLoader(_ context.Context, cfg config.EnvVaultConfig) (*Loader, error) {
	if cfg.URL == "" {
		return nil, fmt.Errorf("url is required")
	}

	httpCli, err := newHTTPClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("new http client: %w", err)
	}
	loader := Loader{
		cfg:        cfg,
		httpClient: httpCli,
	}

	return &loader, nil
}

// LoadKey 实现 cryptox.KeyLoader.
//
// 流程:
//
//  1. 调 httpClient.GetSecret(ctx, cfg.SecretPath)
//  2. 找 "value" 字段 (约定 key 名)
//  3. hex decode (envVault 存 hex 编码的 32 字节 DEK)
//  4. 校验长度 = 32, 返回 []byte
func (l *Loader) LoadKey(ctx context.Context) ([]byte, error) {
	// TODO 实现: 见 http_client.go 的 GetSecret 返回 map[string]string
	//   secret, err := l.client.GetSecret(ctx, l.cfg.SecretPath)
	//   if err != nil { return nil, fmt.Errorf("envvault GetSecret: %w", err) }
	//   hexKey, ok := secret["value"]
	//   if !ok { return nil, fmt.Errorf("envvault: 'value' field not found in %s", l.cfg.SecretPath) }
	//   key, err := hex.DecodeString(hexKey)
	//   if err != nil { return nil, fmt.Errorf("envvault value not hex: %w", err) }
	//   if len(key) != 32 { return nil, fmt.Errorf("envvault key len %d, want 32", len(key)) }
	//   return key, nil

	_ = ctx
	return nil, fmt.Errorf("not implemented")
}

func (l *Loader) LoadProjectSecrets(ctx context.Context, body *SecretsListRequest) (*[]SecretsListResponse, error) {

	if body == nil {
		return nil, fmt.Errorf("body is required")
	}

	buf, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("json marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		l.cfg.URL+apiSecretsList,
		bytes.NewBuffer(buf),
	)
	if err != nil {
		return nil, fmt.Errorf("create request error: %w", err)
	}

	userMeta := ctx2.GetUserMeta(ctx)
	extraHeader := make(http.Header)
	extraHeader.Add("Content-Type", "application/json")
	extraHeader.Add("Authorization", userMeta.JWT)

	respBytes, err := l.httpClient.DoWithExtraHeader(req, extraHeader)
	if err != nil {
		return nil, err
	}
	var resp httpx.Response[[]SecretsListResponse]
	if err = json.Unmarshal(respBytes, &resp); err != nil {
		return nil, fmt.Errorf("json unmarshal: %w", err)
	}

	if resp.IsFailed() {
		return nil, fmt.Errorf("request failed %s", resp.Msg)
	}

	return &resp.Data, nil
}

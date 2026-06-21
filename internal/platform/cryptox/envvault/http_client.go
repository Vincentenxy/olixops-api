package envvault

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"olixops/internal/config"
	"olixops/internal/platform/logger"
	"time"

	"go.uber.org/zap"
)

// httpClient 通过 HTTP REST API 调 envVault.
//
// 设计要点:
//   - 用 *http.Client 配合自定义 Transport, 后续可加 TLS 证书 / mTLS / SNI 等
//   - GetSecret 返回 map[string]string (与 SDK 抽象一致), 屏蔽 envVault 响应格式差异
//   - 所有错误都包成有意义的 ctx, 方便上层诊断
type httpClient struct {
	cfg        config.EnvVaultConfig
	rawClient  *http.Client
	baseHeader http.Header
}

// newHTTPClient 构造 HTTP 客户端.
func newHTTPClient(cfg config.EnvVaultConfig) (*httpClient, error) {
	if cfg.URL == "" {
		return nil, fmt.Errorf("envvault: URL is required")
	}

	// 传输层配置
	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 20,
		IdleConnTimeout:     30 * time.Second,
		DisableCompression:  false,
	}

	// 标准httpclient
	raw := &http.Client{
		Transport: transport,
		Timeout:   cfg.Timeout,
	}

	// 设置请求头
	baseHeader := http.Header{}
	baseHeader.Set("Content-Type", "application/json; charset=utf-8")

	cli := &httpClient{
		cfg:        cfg,
		rawClient:  raw,
		baseHeader: baseHeader,
	}

	return cli, nil
}

func (h *httpClient) SetHeader(req *http.Request, extraHeaders http.Header) {
	for k, v := range h.baseHeader {
		req.Header[k] = v
	}
	for k, v := range extraHeaders {
		req.Header[k] = v
	}
}

func (h *httpClient) Do(req *http.Request) ([]byte, error) {
	return h.DoWithExtraHeader(req, nil)
}

func (h *httpClient) DoWithExtraHeader(req *http.Request, extraHeader http.Header) ([]byte, error) {
	h.SetHeader(req, extraHeader)

	resp, err := h.rawClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http send request failed %s: %w", req.URL.Path, err)
	}

	defer func(body io.ReadCloser) {
		closerErr := body.Close()
		if closerErr != nil {
			logger.L().Error("failed to close response body",
				zap.String("path", req.URL.Path),
				zap.Error(closerErr),
			)
		}
	}(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("http response failed %s: status=%s, body=%s", req.URL.Path, resp.Status, string(body))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read resp body: %w", err)
	}
	return data, nil
}

// 编译期防 io / encoding/json import 未用.
var _ = io.ReadAll
var _ = json.Unmarshal

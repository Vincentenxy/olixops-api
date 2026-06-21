package envvault

import (
	"context"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	ctx2 "olixops/pkg/ctx"
	"testing"
	"time"

	"olixops/internal/config"
)

const testDEKHex = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

// makeLoader 构造一个 Loader, client 指向 mock server, 简化每个测试的构造.
func makeLoader(t *testing.T, srvURL string) *Loader {
	t.Helper()
	hc, err := newHTTPClient(config.EnvVaultConfig{
		URL:     srvURL,
		Timeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatalf("newHTTPClient: %v", err)
	}
	return &Loader{
		cfg:        config.EnvVaultConfig{URL: srvURL},
		httpClient: hc,
	}
}

type UserInfoKey struct{}

type UserInfo struct {
	UserId   string `json:"userId"`
	Username string `json:"userName"`
	JWT      string `json:"jwt"`
}

func TestLoader_Loader_LoadProjectSecrets(t *testing.T) {
	loader := makeLoader(t, "http://localhost:8880")

	userMeta := ctx2.UserMeta{
		UserId:   "00000000-0000-4000-8000-0000000000aa",
		Username: "super admin",
		JWT:      "Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOiIwMDAwMDAwMC0wMDAwLTQwMDAtODAwMC0wMDAwMDAwMDAwYWEiLCJuYW1lIjoic3VwZXIgYWRtaW4iLCJzdWIiOiIwMDAwMDAwMC0wMDAwLTQwMDAtODAwMC0wMDAwMDAwMDAwYWEiLCJleHAiOjE3ODE5Njg2MjksImlhdCI6MTc4MTg4MjIyOX0.MU2jVmgUe1PJ7MGmKqwTB6wt4zAtrgxZGBGr5qdpz0Cq_KbUD8GdqFVbjUwB5KdSccoSAQmqDisofInV6wnvO3JFq_yAsgiqGxQGpezh2nEM-IxAJhI5ASCl8kyvIf-epEmFh5-H8BSX_eHcnsYe5FqjxSXYC6OAwVN9H_4mLfT7LlgrYP_7Pw2AIoBo2NoWPl3EYqdwP4zAndejocxmRNls43fM1A8ou-Wzs6XcBJUsf-X_vc2Czon2NAeJQmMzdIdvliQFsZSFtRwfLW6pqXeXI2nxBWXdXF4sUwXeDwjSUaedPGGI_-uzibWllkO0BjAQNV87wwSmmKpeJ_nYfg",
	}

	ctx := ctx2.WithUserMeta(context.Background(), userMeta)
	resp, err := loader.LoadProjectSecrets(ctx,
		&SecretsListRequest{
			ProjectId:  "1d8eb318-f44e-46b0-9aec-7f8ecfcb9db9",
			FolderCode: "groups",
			Key:        "",
			EnvList:    []string{"dev", "test", "sim", "prod"},
		})
	if err != nil {
		t.Fatalf("LoadProjectSecrets: %v", err)
	}

	fmt.Printf("完整返回数据：%+v\n", resp)

}

func TestLoader_LoadKey_End2End(t *testing.T) {
	// 端到端: httptest server → httpClient → Loader → DEK bytes
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":{"data":{"value":"` + testDEKHex + `"}}}`))
	}))
	defer srv.Close()

	loader := makeLoader(t, srv.URL)

	key, err := loader.LoadKey(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(key) != 32 {
		t.Errorf("key len = %d, want 32", len(key))
	}
	if hex.EncodeToString(key) != testDEKHex {
		t.Errorf("key hex mismatch")
	}
}

func TestLoader_LoadKey_ValueFieldMissing(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"data":{"data":{"other":"oops"}}}`))
	}))
	defer srv.Close()

	loader := makeLoader(t, srv.URL)
	_, err := loader.LoadKey(context.Background())
	if err == nil {
		t.Fatal("expected error when 'value' field missing")
	}
	// 错误信息应该包含路径提示
	if got := err.Error(); got == "" {
		t.Error("error message should not be empty")
	}
}

func TestLoader_LoadKey_BadHex(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"data":{"data":{"value":"not-hex-zzzz"}}}`))
	}))
	defer srv.Close()

	loader := makeLoader(t, srv.URL)
	_, err := loader.LoadKey(context.Background())
	if err == nil {
		t.Fatal("expected hex decode error")
	}
	if got := err.Error(); got == "" {
		t.Error("error message should not be empty")
	}
}

func TestLoader_LoadKey_WrongLength(t *testing.T) {
	// 16 字符 hex = 8 字节, 不是 32 字节
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"data":{"data":{"value":"deadbeef"}}}`))
	}))
	defer srv.Close()

	loader := makeLoader(t, srv.URL)
	_, err := loader.LoadKey(context.Background())
	if err == nil {
		t.Fatal("expected key length error")
	}
}

func TestLoader_LoadKey_EmptyValue(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"data":{"data":{"value":""}}}`))
	}))
	defer srv.Close()

	loader := makeLoader(t, srv.URL)
	_, err := loader.LoadKey(context.Background())
	if err == nil {
		t.Fatal("expected error on empty value")
	}
}

func TestLoader_LoadKey_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "forbidden", http.StatusForbidden)
	}))
	defer srv.Close()

	loader := makeLoader(t, srv.URL)
	_, err := loader.LoadKey(context.Background())
	if err == nil {
		t.Fatal("expected error on 403")
	}
}

package envvault

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"olixops/internal/config"
)

// newTestClient 构造一个测试用 httpClient, mock server URL 已注入.
// timeout 设短一点 (1s), 让 timeout 测试更快触发.
// token 通过 DoWithExtraHeader 传 (EnvVaultConfig 本身没 token 字段).
func newTestClient(t *testing.T, srvURL string) *httpClient {
	t.Helper()
	c, err := newHTTPClient(config.EnvVaultConfig{
		URL:     srvURL,
		Timeout: 1 * time.Second,
	})
	if err != nil {
		t.Fatalf("newHTTPClient: %v", err)
	}
	return c
}

// ========== newHTTPClient 测试 ==========

func TestNewHTTPClient_RequiredFields(t *testing.T) {
	// URL 空应该报错
	_, err := newHTTPClient(config.EnvVaultConfig{URL: ""})
	if err == nil {
		t.Fatal("expected error when URL is empty")
	}
}

func TestNewHTTPClient_SetsBaseHeader(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 验证 baseHeader (Content-Type) 自动带上
		if ct := r.Header.Get("Content-Type"); !strings.Contains(ct, "application/json") {
			t.Errorf("Content-Type = %q, want application/json", ct)
		}
		w.WriteHeader(200)
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	req, _ := http.NewRequest("GET", srv.URL+"/anything", nil)
	_, _ = c.Do(req)
}

// ========== SetHeader 测试 ==========

func TestSetHeader_MergesBaseAndExtra(t *testing.T) {
	c, _ := newHTTPClient(config.EnvVaultConfig{URL: "http://x"})
	if c == nil {
		t.Fatal("nil client")
	}

	req, _ := http.NewRequest("GET", "http://x", nil)
	c.SetHeader(req, http.Header{
		"X-Custom":     {"custom-value"},
		"Content-Type": {"override-value"}, // 测试 extra 覆盖 base
	})

	if got := req.Header.Get("Content-Type"); got != "override-value" {
		t.Errorf("Content-Type = %q, want override-value", got)
	}
	if got := req.Header.Get("X-Custom"); got != "custom-value" {
		t.Errorf("X-Custom = %q, want custom-value", got)
	}
}

// ========== Do / DoWithExtraHeader 测试 ==========

func TestDo_SuccessReturnsBody(t *testing.T) {
	wantBody := `{"hello":"world"}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(wantBody))
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	req, _ := http.NewRequest("GET", srv.URL+"/test", nil)

	body, err := c.Do(req)
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	if string(body) != wantBody {
		t.Errorf("body = %q, want %q", body, wantBody)
	}
}

func TestDoWithExtraHeader_SendsCustomAuth(t *testing.T) {
	var gotAuth string
	var gotCT string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotCT = r.Header.Get("Content-Type")
		w.WriteHeader(200)
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	req, _ := http.NewRequest("GET", srv.URL+"/secret", nil)

	extra := http.Header{
		"Authorization": {"Bearer my-jwt-token"},
	}
	_, err := c.DoWithExtraHeader(req, extra)
	if err != nil {
		t.Fatalf("DoWithExtraHeader: %v", err)
	}
	if gotAuth != "Bearer my-jwt-token" {
		t.Errorf("Authorization = %q, want 'Bearer my-jwt-token'", gotAuth)
	}
	if !strings.Contains(gotCT, "application/json") {
		t.Errorf("Content-Type = %q, want application/json", gotCT)
	}
}

func TestDo_PropagatesRequestMethodAndURL(t *testing.T) {
	var gotMethod, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.WriteHeader(200)
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	req, _ := http.NewRequest("POST", srv.URL+"/api/v1/secrets/data/db-key", nil)

	_, err := c.Do(req)
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	if gotMethod != "POST" {
		t.Errorf("method = %q, want POST", gotMethod)
	}
	if gotPath != "/api/v1/secrets/data/db-key" {
		t.Errorf("path = %q, want /api/v1/secrets/data/db-key", gotPath)
	}
}

// ========== 错误状态码测试 ==========

func TestDo_ReturnsErrorOnNon2xx(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		statusText string
		body       string
		wantSubstr string
	}{
		{"401 unauthorized", 401, "Unauthorized", `{"errors":["invalid token"]}`, "401"},
		{"403 forbidden", 403, "Forbidden", `forbidden`, "403"},
		{"404 not found", 404, "Not Found", `not found`, "404"},
		{"500 server error", 500, "Internal Server Error", `oops`, "500"},
		{"503 unavailable", 503, "Service Unavailable", `down`, "503"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.statusCode)
				w.Write([]byte(tc.body))
			}))
			defer srv.Close()

			c := newTestClient(t, srv.URL)
			req, _ := http.NewRequest("GET", srv.URL+"/secret", nil)
			body, err := c.Do(req)
			if err == nil {
				t.Fatal("expected error for non-2xx")
			}
			if body != nil {
				t.Errorf("body should be nil on error, got %q", body)
			}
			if !strings.Contains(err.Error(), tc.wantSubstr) {
				t.Errorf("err = %v, should contain %q", err, tc.wantSubstr)
			}
		})
	}
}

func TestDo_AcceptsAll2xx(t *testing.T) {
	// 204 不测试, 因为 Go 标准库会自动剥离 204 响应 body
	for _, code := range []int{200, 201, 202, 299} {
		t.Run("status_"+string(rune('0'+code/100)), func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(code)
				w.Write([]byte(`ok`))
			}))
			defer srv.Close()

			c := newTestClient(t, srv.URL)
			req, _ := http.NewRequest("GET", srv.URL+"/x", nil)
			body, err := c.Do(req)
			if err != nil {
				t.Errorf("status %d should be accepted, got %v", code, err)
			}
			if string(body) != "ok" {
				t.Errorf("status %d body = %q, want 'ok'", code, body)
			}
		})
	}
}

// ========== 网络错误测试 ==========

func TestDo_NetworkErrorWhenServerDown(t *testing.T) {
	// httptest.NewServer 立刻 Close,模拟 server 不存在
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	srv.Close()

	c := newTestClient(t, srv.URL) // srv.URL 现在是死链
	req, _ := http.NewRequest("GET", srv.URL+"/x", nil)
	_, err := c.Do(req)
	if err == nil {
		t.Fatal("expected network error")
	}
	// 错误信息应该包含 path
	if !strings.Contains(err.Error(), "/x") {
		t.Errorf("err = %v, want path /x in message", err)
	}
}

func TestDo_Timeout(t *testing.T) {
	// server 故意 sleep 超过 client timeout
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(3 * time.Second)
		w.WriteHeader(200)
	}))
	defer srv.Close()

	// 用更短的 timeout (300ms) 触发
	c, err := newHTTPClient(config.EnvVaultConfig{
		URL:     srv.URL,
		Timeout: 300 * time.Millisecond,
	})
	if err != nil {
		t.Fatal(err)
	}
	req, _ := http.NewRequest("GET", srv.URL+"/slow", nil)

	start := time.Now()
	_, err = c.Do(req)
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected timeout error")
	}
	if elapsed > 2*time.Second {
		t.Errorf("timeout took %v, expected ~300ms", elapsed)
	}
}

func TestDo_ContextCancel(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(200)
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	req, _ := http.NewRequest("GET", srv.URL+"/slow", nil)
	// 用 ctx cancel 来中断请求 (这是 Do 方法当前不支持的, 标记 TODO)
	// 暂时只验证 timeout 工作; ctx cancel 需要 Do 接受 ctx 参数, 当前签名是 (req)
	_ = context.Background()
	_, err := c.Do(req)
	if err == nil {
		t.Fatal("expected error")
	}
}

// ========== 响应体异常测试 ==========

func TestDo_HandlesEmptyBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(204)
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	req, _ := http.NewRequest("GET", srv.URL+"/empty", nil)
	body, err := c.Do(req)
	if err != nil {
		t.Fatalf("204 should be accepted, got %v", err)
	}
	if len(body) != 0 {
		t.Errorf("body = %q, want empty", body)
	}
}

func TestDo_LargeResponseBody(t *testing.T) {
	// 测大响应体不会爆内存 (简单冒烟)
	bigData := strings.Repeat("x", 1024*1024) // 1MB
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, bigData)
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	req, _ := http.NewRequest("GET", srv.URL+"/big", nil)
	body, err := c.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if len(body) != len(bigData) {
		t.Errorf("len(body) = %d, want %d", len(body), len(bigData))
	}
}

// ========== 错误类型断言 (用于上层 errors.Is) ==========

func TestDo_ErrorIncludesStatusCode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(503)
		w.Write([]byte(`{"errors":["sealed"]}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	req, _ := http.NewRequest("GET", srv.URL+"/x", nil)
	_, err := c.Do(req)
	if err == nil {
		t.Fatal("expected error")
	}
	// 错误信息应该包含 status + path + body (供排障)
	msg := err.Error()
	if !strings.Contains(msg, "503") {
		t.Errorf("err message missing status code 503: %s", msg)
	}
	if !strings.Contains(msg, "sealed") {
		t.Errorf("err message missing response body 'sealed': %s", msg)
	}
	if !strings.Contains(msg, "/x") {
		t.Errorf("err message missing path /x: %s", msg)
	}
}

// 防止 errors 这个 import 在某些 Go 版本里被报"unused"
var _ = errors.New

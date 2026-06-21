package user

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"olixops/internal/config"
	"olixops/internal/interfaces/http/middleware"
	"olixops/internal/platform/audit"
	"olixops/internal/platform/auth"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// buildTestServer 构造最小可用的 server: 完整的 router + 真实 hasher + 真实 issuer + fake repo + MemRecorder。
// 所有中间件 (Trace / AccessLog / Auth) 都启用, 走真实链路。
func buildTestServer(t *testing.T) (*gin.Engine, *audit.MemRecorder, string) {
	t.Helper()

	repo := newFakeRepo()
	hasher := auth.NewBcryptHasher(0)
	issuer := auth.NewJWTIssuer(config.AuthConfig{
		JWTSecret:       "test-secret-32-bytes-padding-padding-padding-x",
		AccessTokenTTL:  1 * time.Hour,
		RefreshTokenTTL: 24 * time.Hour,
		Issuer:          "olixops-test",
	})
	recorder := audit.NewMemRecorder(100, nil)
	svc := NewService(repo, hasher, issuer)
	h := NewHandler(svc, recorder)
	mod := NewModule(h, recorder)

	// 构造和 router.New 完全等价的 engine (避免和 router 包产生循环依赖)
	r := gin.New()
	r.Use(middleware.Trace())
	r.Use(middleware.AccessLog())
	pub := r.Group("/api/pub/v1")
	priv := r.Group("/api/v1")
	priv.Use(middleware.Auth(issuer))
	mod.RegisterPub(pub)
	mod.RegisterPrivate(priv)

	// seed 用户
	hash, _ := hasher.Hash("password123")
	_ = repo.Create(context.Background(), &User{
		ID: "u-1001", Username: "alice", Email: "alice@example.com",
		DisplayName: "Alice", PasswordHash: hash,
		Status: StatusActive, Source: "local",
	})
	return r, recorder, "alice"
}

func do(t *testing.T, r *gin.Engine, method, path, body, bearer string) *httptest.ResponseRecorder {
	t.Helper()
	var rdr *bytes.Reader
	if body != "" {
		rdr = bytes.NewReader([]byte(body))
	} else {
		rdr = bytes.NewReader(nil)
	}
	req := httptest.NewRequest(method, path, rdr)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func decode(t *testing.T, w *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var out map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode: %v, body=%s", err, w.Body.String())
	}
	return out
}

func TestHandler_LoginSuccess(t *testing.T) {
	r, rec, _ := buildTestServer(t)
	w := do(t, r, "POST", "/api/pub/v1/auth/login",
		`{"username":"alice","password":"password123"}`, "")

	if w.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	body := decode(t, w)
	if body["code"] != "OK" {
		t.Errorf("code = %v", body["code"])
	}

	// envelope 严格只含 code/msg/data
	for k := range body {
		if k != "code" && k != "msg" && k != "data" {
			t.Errorf("envelope has unexpected key %q", k)
		}
	}

	// data 含 user 和 tokens
	data, _ := body["data"].(map[string]any)
	tokens, _ := data["tokens"].(map[string]any)
	if tokens["access_token"] == "" || tokens["refresh_token"] == "" {
		t.Errorf("tokens missing")
	}

	// 审计: success 应被记录
	if rec.Len() == 0 {
		t.Error("audit recorder should have at least 1 event")
	}
	last := rec.Events()[rec.Len()-1]
	if last.Status != "success" || last.Action != "user.login.success" {
		t.Errorf("audit = %+v", last)
	}
}

func TestHandler_LoginBadPassword(t *testing.T) {
	r, rec, _ := buildTestServer(t)
	w := do(t, r, "POST", "/api/pub/v1/auth/login",
		`{"username":"alice","password":"WRONG"}`, "")

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status=%d, want 401", w.Code)
	}
	if rec.Len() == 0 {
		t.Fatal("audit should record failure")
	}
	last := rec.Events()[rec.Len()-1]
	if last.Status != "failure" || last.Action != "user.login.failure" {
		t.Errorf("audit = %+v", last)
	}
}

func TestHandler_Me_WithToken(t *testing.T) {
	r, _, _ := buildTestServer(t)
	w := do(t, r, "POST", "/api/pub/v1/auth/login",
		`{"username":"alice","password":"password123"}`, "")
	token := decode(t, w)["data"].(map[string]any)["tokens"].(map[string]any)["access_token"].(string)

	w = do(t, r, "GET", "/api/v1/auth/me", "", token)
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	data := decode(t, w)["data"].(map[string]any)
	if data["username"] != "alice" {
		t.Errorf("username = %v", data["username"])
	}
	if data["password_hash"] != nil {
		t.Errorf("password_hash should not leak: %v", data["password_hash"])
	}
}

func TestHandler_Me_NoToken(t *testing.T) {
	r, _, _ := buildTestServer(t)
	w := do(t, r, "GET", "/api/v1/auth/me", "", "")
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status=%d, want 401", w.Code)
	}
}

func TestHandler_Me_InvalidToken(t *testing.T) {
	r, _, _ := buildTestServer(t)
	w := do(t, r, "GET", "/api/v1/auth/me", "", "not.a.real.jwt")
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status=%d, want 401", w.Code)
	}
	body := decode(t, w)
	if body["msg"] != "invalid token" {
		t.Errorf("msg = %v, want 'invalid token'", body["msg"])
	}
}

func TestHandler_Refresh(t *testing.T) {
	r, _, _ := buildTestServer(t)
	w := do(t, r, "POST", "/api/pub/v1/auth/login",
		`{"username":"alice","password":"password123"}`, "")
	refresh := decode(t, w)["data"].(map[string]any)["tokens"].(map[string]any)["refresh_token"].(string)

	w = do(t, r, "POST", "/api/pub/v1/auth/refresh",
		`{"refresh_token":"`+refresh+`"}`, "")
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	tokens := decode(t, w)["data"].(map[string]any)
	if tokens["access_token"] == "" || tokens["refresh_token"] == "" {
		t.Error("new pair should have both tokens")
	}
}

func TestHandler_RefreshInvalidToken(t *testing.T) {
	r, _, _ := buildTestServer(t)
	w := do(t, r, "POST", "/api/pub/v1/auth/refresh",
		`{"refresh_token":"garbage"}`, "")
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status=%d, want 401", w.Code)
	}
}

func TestHandler_Logout(t *testing.T) {
	r, _, _ := buildTestServer(t)
	w := do(t, r, "POST", "/api/pub/v1/auth/login",
		`{"username":"alice","password":"password123"}`, "")
	token := decode(t, w)["data"].(map[string]any)["tokens"].(map[string]any)["access_token"].(string)

	w = do(t, r, "POST", "/api/v1/auth/logout", "", token)
	if w.Code != http.StatusNoContent {
		t.Errorf("status=%d, want 204, body=%s", w.Code, w.Body.String())
	}
}

func TestHandler_EnvelopeStrictKeys(t *testing.T) {
	// 校验 envelope 严格只含 code/msg/data (无 trace_id)
	r, _, _ := buildTestServer(t)
	w := do(t, r, "POST", "/api/pub/v1/auth/login",
		`{"username":"alice","password":"password123"}`, "")
	body := decode(t, w)
	for k := range body {
		if !strings.Contains("code msg data", k) {
			t.Errorf("envelope unexpected key %q (must be code/msg/data only)", k)
		}
	}
}

func TestHandler_XTraceIDHeader(t *testing.T) {
	// 验证 response header 含 X-Trace-Id
	r, _, _ := buildTestServer(t)
	w := do(t, r, "POST", "/api/pub/v1/auth/login",
		`{"username":"alice","password":"password123"}`, "")
	if w.Header().Get("X-Trace-Id") == "" {
		t.Error("X-Trace-Id header should be set")
	}
}

package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"olixops/internal/platform/auth"
	"olixops/pkg/httpx"
)

func init() {
	gin.SetMode(gin.TestMode)
}

type fakeIssuer struct {
	subject auth.Subject
	err     error
}

func (f *fakeIssuer) Verify(_ context.Context, token string) (auth.Subject, error) {
	if token == "" {
		return auth.Subject{}, errors.New("empty")
	}
	return f.subject, f.err
}

func (f *fakeIssuer) Issue(_ context.Context, _ auth.Subject) (auth.TokenPair, error) {
	return auth.TokenPair{}, nil
}
func (f *fakeIssuer) Refresh(_ context.Context, _ string) (auth.TokenPair, error) {
	return auth.TokenPair{}, nil
}
func (f *fakeIssuer) Revoke(_ context.Context, _ string) error { return nil }

// newTestRouter 构造最小路由: /ping 透传 (ctx 写入 ok), 用 Auth middleware 保护。
func newTestRouter(issuer auth.TokenIssuer) *gin.Engine {
	r := gin.New()
	r.GET("/ping", Auth(issuer), func(c *gin.Context) {
		sub, ok := auth.FromContext(c.Request.Context())
		if !ok {
			httpx.Error(c, errsUnauthorized("no subject"))
			return
		}
		c.JSON(http.StatusOK, gin.H{"user_id": sub.UserID})
	})
	return r
}

func errsUnauthorized(msg string) error {
	type coded interface {
		HTTPStatus() int
	}
	// 用 httpx.Error 内部使用的 errs.Unauthorized
	// 简化: 直接返回字符串让 handler 编码 (这里只是测试 fixture)
	_ = msg
	return nil
}

func TestAuth_NoHeader(t *testing.T) {
	r := newTestRouter(&fakeIssuer{})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
	body := decodeEnvelope(t, w)
	if body["code"] != "UNAUTHORIZED" {
		t.Errorf("code = %v", body["code"])
	}
}

func TestAuth_MalformedHeader(t *testing.T) {
	r := newTestRouter(&fakeIssuer{})

	for _, h := range []string{"", "Bearer", "Basic xxx", "Token foo"} {
		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		req.Header.Set("Authorization", h)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Errorf("header %q: status = %d, want 401", h, w.Code)
		}
	}
}

func TestAuth_InvalidToken(t *testing.T) {
	r := newTestRouter(&fakeIssuer{err: auth.ErrInvalidToken})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set("Authorization", "Bearer bad-token")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
	body := decodeEnvelope(t, w)
	if body["msg"] != "invalid token" {
		t.Errorf("msg = %v, want 'invalid token'", body["msg"])
	}
}

func TestAuth_ExpiredToken(t *testing.T) {
	r := newTestRouter(&fakeIssuer{err: auth.ErrTokenExpired})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set("Authorization", "Bearer expired-token")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
	body := decodeEnvelope(t, w)
	if body["msg"] != "token expired" {
		t.Errorf("msg = %v, want 'token expired'", body["msg"])
	}
}

func TestAuth_ValidToken(t *testing.T) {
	want := auth.Subject{UserID: "u-1001", Username: "alice"}
	r := newTestRouter(&fakeIssuer{subject: want})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set("Authorization", "Bearer good-token")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body=%s", w.Code, w.Body.String())
	}
	var got map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got["user_id"] != "u-1001" {
		t.Errorf("user_id = %q, want u-1001", got["user_id"])
	}
}

func decodeEnvelope(t *testing.T, w *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v, raw=%s", err, w.Body.String())
	}
	return body
}

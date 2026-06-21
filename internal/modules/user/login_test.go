package user

import (
	"context"
	"errors"
	"testing"
	"time"

	"olixops/internal/config"
	"olixops/internal/platform/auth"
	"olixops/pkg/errs"
)

func newTestService(t *testing.T) (*Service, *fakeRepo) {
	t.Helper()
	repo := newFakeRepo()
	hasher := auth.NewBcryptHasher(0)
	issuer := auth.NewJWTIssuer(config.AuthConfig{
		JWTSecret:       "test-secret-32-bytes-padding-padding-padding-x",
		AccessTokenTTL:  1 * time.Hour,
		RefreshTokenTTL: 24 * time.Hour,
		Issuer:          "olixops-test",
	})
	// seed 一个 active 用户
	hash, _ := hasher.Hash("password123")
	_ = repo.Create(context.Background(), &User{
		ID:           "u-1001",
		Username:     "alice",
		Email:        "alice@example.com",
		DisplayName:  "Alice",
		PasswordHash: hash,
		Status:       StatusActive,
		Source:       "local",
	})
	return NewService(repo, hasher, issuer), repo
}

func TestLogin_HappyPath(t *testing.T) {
	svc, _ := newTestService(t)
	res, err := svc.Login(context.Background(), LoginInput{Username: "alice", Password: "password123"})
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if res.User.ID != "u-1001" {
		t.Errorf("User.ID = %q", res.User.ID)
	}
	if res.Tokens.AccessToken == "" || res.Tokens.RefreshToken == "" {
		t.Error("tokens should not be empty")
	}
	if res.Tokens.TokenType != "Bearer" {
		t.Errorf("TokenType = %q", res.Tokens.TokenType)
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	svc, _ := newTestService(t)
	_, err := svc.Login(context.Background(), LoginInput{Username: "alice", Password: "WRONG"})
	if err == nil {
		t.Fatal("expected error")
	}
	e := errs.As(err)
	if e == nil || e.Code != errs.CodeUnauthorized {
		t.Errorf("err = %v, want UNAUTHORIZED", err)
	}
}

func TestLogin_UserNotFound(t *testing.T) {
	svc, _ := newTestService(t)
	_, err := svc.Login(context.Background(), LoginInput{Username: "ghost", Password: "any"})
	e := errs.As(err)
	if e == nil || e.Code != errs.CodeUnauthorized {
		t.Errorf("user not found should return UNAUTHORIZED (anti-enumeration), got %v", err)
	}
}

func TestLogin_InactiveUser(t *testing.T) {
	svc, repo := newTestService(t)
	hash, _ := auth.NewBcryptHasher(0).Hash("password123")
	_ = repo.Create(context.Background(), &User{
		ID: "u-locked", Username: "bob", Email: "bob@x.com",
		PasswordHash: hash, Status: StatusLocked, Source: "local",
	})
	_, err := svc.Login(context.Background(), LoginInput{Username: "bob", Password: "password123"})
	e := errs.As(err)
	if e == nil || e.Code != errs.CodeForbidden {
		t.Errorf("inactive user: err = %v, want FORBIDDEN", err)
	}
}

func TestLogin_UpdatesLastLogin(t *testing.T) {
	svc, repo := newTestService(t)
	_, err := svc.Login(context.Background(), LoginInput{Username: "alice", Password: "password123", IP: "10.0.0.1"})
	if err != nil {
		t.Fatal(err)
	}
	u, _ := repo.FindByID(context.Background(), "u-1001")
	if u.LastLoginIP != "10.0.0.1" {
		t.Errorf("LastLoginIP = %q, want 10.0.0.1", u.LastLoginIP)
	}
	if u.LastLoginAt == nil {
		t.Error("LastLoginAt should be set")
	}
}

func TestRefresh_RetainsClaims(t *testing.T) {
	svc, _ := newTestService(t)
	res, _ := svc.Login(context.Background(), LoginInput{Username: "alice", Password: "password123"})

	pair, err := svc.Refresh(context.Background(), RefreshInput{RefreshToken: res.Tokens.RefreshToken})
	if err != nil {
		t.Fatalf("Refresh: %v", err)
	}
	if pair.AccessToken == res.Tokens.AccessToken {
		t.Error("Refresh should produce new access token")
	}

	// 解析新 access token 验证 Username/Email 没丢
	parts := parseJWTForTest(t, pair.AccessToken)
	if parts["username"] != "alice" {
		t.Errorf("username lost: %v", parts["username"])
	}
	if parts["email"] != "alice@example.com" {
		t.Errorf("email lost: %v", parts["email"])
	}
}

func TestMe_FromContext(t *testing.T) {
	svc, _ := newTestService(t)
	res, _ := svc.Login(context.Background(), LoginInput{Username: "alice", Password: "password123"})

	ctx := auth.WithSubject(context.Background(), auth.Subject{
		UserID: res.User.ID, Username: res.User.Username, Email: res.User.Email,
	})
	u, err := svc.Me(ctx)
	if err != nil {
		t.Fatalf("Me: %v", err)
	}
	if u.ID != "u-1001" || u.Username != "alice" {
		t.Errorf("Me returned %+v", u)
	}
}

func TestMe_NoSubject(t *testing.T) {
	svc, _ := newTestService(t)
	_, err := svc.Me(context.Background())
	if err == nil || !errors.Is(err, auth.ErrUserNotFound) && errs.As(err).Code != errs.CodeUnauthorized {
		t.Errorf("Me without subject: %v", err)
	}
}

// parseJWTForTest 解析 JWT payload 部分 (中间段 base64), 仅用于测试断言。
func parseJWTForTest(t *testing.T, token string) map[string]any {
	t.Helper()
	parts := splitJWT(token)
	if len(parts) != 3 {
		t.Fatalf("invalid JWT, got %d parts", len(parts))
	}
	return decodeBase64JSON(t, parts[1])
}

func splitJWT(token string) []string {
	out := []string{}
	start := 0
	for i := 0; i < len(token); i++ {
		if token[i] == '.' {
			out = append(out, token[start:i])
			start = i + 1
		}
	}
	out = append(out, token[start:])
	return out
}

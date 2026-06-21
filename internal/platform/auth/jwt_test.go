package auth

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"olixops/internal/config"
)

func testIssuer(t *testing.T) *JWTIssuer {
	t.Helper()
	return NewJWTIssuer(config.AuthConfig{
		JWTSecret:       "test-secret-at-least-32-bytes-long-padding-padding",
		AccessTokenTTL:  1 * time.Hour,
		RefreshTokenTTL: 24 * time.Hour,
		Issuer:          "olixops-test",
	})
}

func sampleSubject() Subject {
	return Subject{
		UserID:   "u-1001",
		Username: "alice",
		Email:    "alice@example.com",
		TeamIDs:  []string{"t-1", "t-2"},
		Roles:    []string{"admin"},
		Extra:    map[string]string{"dept": "engineering"},
	}
}

func TestIssueVerifyRoundTrip(t *testing.T) {
	j := testIssuer(t)
	pair, err := j.Issue(context.Background(), sampleSubject())
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	if pair.AccessToken == "" || pair.RefreshToken == "" {
		t.Fatal("tokens should not be empty")
	}
	if pair.TokenType != "Bearer" {
		t.Errorf("TokenType = %q, want Bearer", pair.TokenType)
	}

	sub, err := j.Verify(context.Background(), pair.AccessToken)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	want := sampleSubject()
	if sub.UserID != want.UserID || sub.Username != want.Username || sub.Email != want.Email {
		t.Errorf("Subject = %+v, want %+v", sub, want)
	}
	if len(sub.TeamIDs) != 2 || sub.TeamIDs[0] != "t-1" {
		t.Errorf("TeamIDs = %v", sub.TeamIDs)
	}
	if len(sub.Roles) != 1 || sub.Roles[0] != "admin" {
		t.Errorf("Roles = %v", sub.Roles)
	}
	if sub.Extra["dept"] != "engineering" {
		t.Errorf("Extra = %v", sub.Extra)
	}
}

func TestVerifyRejectsRefreshToken(t *testing.T) {
	j := testIssuer(t)
	pair, err := j.Issue(context.Background(), sampleSubject())
	if err != nil {
		t.Fatal(err)
	}
	_, err = j.Verify(context.Background(), pair.RefreshToken)
	if !errors.Is(err, ErrInvalidToken) {
		t.Errorf("Verify with refresh token = %v, want ErrInvalidToken", err)
	}
}

func TestRefreshProducesNewPairWithSameClaims(t *testing.T) {
	j := testIssuer(t)
	first, err := j.Issue(context.Background(), sampleSubject())
	if err != nil {
		t.Fatal(err)
	}

	second, err := j.Refresh(context.Background(), first.RefreshToken)
	if err != nil {
		t.Fatalf("Refresh: %v", err)
	}
	if second.AccessToken == first.AccessToken {
		t.Error("Refresh should produce a new access token")
	}

	// Refresh → Issue 链路必须保留 username/email/roles/teams/extra
	sub, err := j.Verify(context.Background(), second.AccessToken)
	if err != nil {
		t.Fatal(err)
	}
	if sub.Username != "alice" {
		t.Errorf("Username lost on refresh: got %q", sub.Username)
	}
	if sub.Email != "alice@example.com" {
		t.Errorf("Email lost on refresh: got %q", sub.Email)
	}
	if len(sub.Roles) != 1 || sub.Roles[0] != "admin" {
		t.Errorf("Roles lost on refresh: got %v", sub.Roles)
	}
	if len(sub.TeamIDs) != 2 {
		t.Errorf("TeamIDs lost on refresh: got %v", sub.TeamIDs)
	}
	if sub.Extra["dept"] != "engineering" {
		t.Errorf("Extra lost on refresh: got %v", sub.Extra)
	}
}

func TestVerifyExpiredToken(t *testing.T) {
	j := NewJWTIssuer(config.AuthConfig{
		JWTSecret:       "test-secret-at-least-32-bytes-long-padding-padding",
		AccessTokenTTL:  -1 * time.Hour, // 已过期
		RefreshTokenTTL: 24 * time.Hour,
		Issuer:          "olixops-test",
	})
	pair, err := j.Issue(context.Background(), sampleSubject())
	if err != nil {
		t.Fatal(err)
	}
	_, err = j.Verify(context.Background(), pair.AccessToken)
	if !errors.Is(err, ErrTokenExpired) {
		t.Errorf("expired token: got %v, want ErrTokenExpired", err)
	}
}

func TestVerifyInvalidSignature(t *testing.T) {
	j1 := testIssuer(t)
	j2 := NewJWTIssuer(config.AuthConfig{
		JWTSecret:       "different-secret-also-32-bytes-long-padding-x",
		AccessTokenTTL:  1 * time.Hour,
		RefreshTokenTTL: 24 * time.Hour,
		Issuer:          "olixops-test",
	})
	pair, err := j1.Issue(context.Background(), sampleSubject())
	if err != nil {
		t.Fatal(err)
	}
	_, err = j2.Verify(context.Background(), pair.AccessToken)
	if !errors.Is(err, ErrInvalidToken) {
		t.Errorf("wrong secret: got %v, want ErrInvalidToken", err)
	}
}

func TestVerifyMalformedToken(t *testing.T) {
	j := testIssuer(t)
	for _, raw := range []string{"", "not.a.token", "only-one-part", "two.parts"} {
		_, err := j.Verify(context.Background(), raw)
		if !errors.Is(err, ErrInvalidToken) {
			t.Errorf("malformed token %q: got %v, want ErrInvalidToken", raw, err)
		}
	}
}

func TestRevokeIsNoop(t *testing.T) {
	j := testIssuer(t)
	if err := j.Revoke(context.Background(), "any-token"); err != nil {
		t.Errorf("Revoke should be noop, got %v", err)
	}
}

func TestIssueRequiresSecret(t *testing.T) {
	j := NewJWTIssuer(config.AuthConfig{JWTSecret: ""})
	_, err := j.Issue(context.Background(), sampleSubject())
	if err == nil || !strings.Contains(err.Error(), "jwt_secret") {
		t.Errorf("expected jwt_secret error, got %v", err)
	}
}

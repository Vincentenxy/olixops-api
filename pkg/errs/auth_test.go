package errs

import (
	"errors"
	"fmt"
	"testing"

	"olixops/internal/platform/auth"
)

func TestFromAuthError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantCode Code
		wantMsg  string
	}{
		{"token expired", auth.ErrTokenExpired, CodeUnauthorized, "token expired"},
		{"invalid token", auth.ErrInvalidToken, CodeUnauthorized, "invalid token"},
		{"invalid password", auth.ErrInvalidPassword, CodeUnauthorized, "invalid credentials"},
		{"user not found", auth.ErrUserNotFound, CodeUnauthorized, "user not found"},
		{"provider unready", auth.ErrProviderUnready, CodeUnavailable, "oauth provider not configured"},
		{"wrapped expired", fmt.Errorf("verify: %w", auth.ErrTokenExpired), CodeUnauthorized, "token expired"},
		{"unknown error", errors.New("random"), CodeUnauthorized, "auth failure"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FromAuthError(tt.err)
			if got.Code != tt.wantCode {
				t.Errorf("code = %s, want %s", got.Code, tt.wantCode)
			}
			if got.Message != tt.wantMsg {
				t.Errorf("message = %q, want %q", got.Message, tt.wantMsg)
			}
		})
	}
}

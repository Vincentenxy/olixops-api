// Package auth 抽象统一认证能力,业务层只依赖这里的接口。
//
// 当前提供:
//   - JWT 签发与校验(内部默认实现)
//   - OAuth2/OIDC 适配接口(实现留待 adapter 层)
//   - 密码哈希工具
package auth

import (
	"context"
	"errors"
	"time"
)

var (
	ErrInvalidToken    = errors.New("auth: invalid token")
	ErrTokenExpired    = errors.New("auth: token expired")
	ErrInvalidPassword = errors.New("auth: invalid password")
	ErrUserNotFound    = errors.New("auth: user not found")
	ErrProviderUnready = errors.New("auth: oauth2 provider not configured")
)

// Subject 表示已认证主体,从 Token 中解析得到。
type Subject struct {
	UserID    string            `json:"user_id"`
	Username  string            `json:"username"`
	Email     string            `json:"email"`
	TeamIDs   []string          `json:"team_ids,omitempty"`
	Roles     []string          `json:"roles,omitempty"`
	Issuer    string            `json:"issuer,omitempty"`
	ExpiresAt time.Time         `json:"expires_at"`
	Extra     map[string]string `json:"extra,omitempty"`
}

// TokenPair 是访问令牌与刷新令牌组合。
type TokenPair struct {
	AccessToken           string    `json:"access_token"`
	RefreshToken          string    `json:"refresh_token"`
	AccessTokenExpiresAt  time.Time `json:"access_token_expires_at"`
	RefreshTokenExpiresAt time.Time `json:"refresh_token_expires_at"`
	TokenType             string    `json:"token_type"` // 通常为 Bearer
}

// TokenIssuer 负责签发与校验业务 Token。
type TokenIssuer interface {
	Issue(ctx context.Context, sub Subject) (TokenPair, error)
	Verify(ctx context.Context, accessToken string) (Subject, error)
	Refresh(ctx context.Context, refreshToken string) (TokenPair, error)
	Revoke(ctx context.Context, accessToken string) error
}

// PasswordHasher 用于本地账户密码哈希与校验。
type PasswordHasher interface {
	Hash(plain string) (string, error)
	Verify(plain, hashed string) error
}

// OAuth2Provider 抽象 OAuth2/OIDC 适配,具体实现放在 adapter 层。
type OAuth2Provider interface {
	Name() string
	AuthorizeURL(state string) (string, error)
	Exchange(ctx context.Context, code string) (TokenPair, error)
	UserInfo(ctx context.Context, accessToken string) (Subject, error)
}

// 用作 context 中存放 Subject 的 key。
type subjectCtxKey struct{}

// WithSubject 将 Subject 注入 context。
func WithSubject(ctx context.Context, sub Subject) context.Context {
	return context.WithValue(ctx, subjectCtxKey{}, sub)
}

// FromContext 取出 Subject。
func FromContext(ctx context.Context) (Subject, bool) {
	if ctx == nil {
		return Subject{}, false
	}
	sub, ok := ctx.Value(subjectCtxKey{}).(Subject)
	return sub, ok
}

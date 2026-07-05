package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"olixops/internal/config"
)

// JWTIssuer 是基于 HMAC-SHA256 的默认 TokenIssuer 实现
type JWTIssuer struct {
	cfg *config.AuthConfig
}

// 编译期断言: JWTIssuer 必须实现 TokenIssuer interface.
// 加新方法时编译器立刻报错, 不用等业务代码用到才发现.
var _ TokenIssuer = (*JWTIssuer)(nil)

// NewJWTIssuer 构造默认 JWT 签发器。
func NewJWTIssuer(cfg *config.AuthConfig) *JWTIssuer {
	return &JWTIssuer{cfg: cfg}
}

type jwtClaims struct {
	jwt.RegisteredClaims
	Username string            `json:"username,omitempty"`
	Email    string            `json:"email,omitempty"`
	TeamIDs  []string          `json:"team_ids,omitempty"`
	Roles    []string          `json:"roles,omitempty"`
	TokenUse string            `json:"token_use,omitempty"` // access / refresh
	Extra    map[string]string `json:"extra,omitempty"`
}

// Issue 签发 access + refresh 一对令牌。
func (j *JWTIssuer) Issue(_ context.Context, sub Subject) (TokenPair, error) {
	if j.cfg.JWTSecret == "" {
		return TokenPair{}, errors.New("auth.jwt_secret is empty")
	}
	now := time.Now()
	accessExp := now.Add(j.cfg.AccessTokenTTL)
	refreshExp := now.Add(j.cfg.RefreshTokenTTL)

	access, err := j.signClaims(jwtClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   sub.UserID,
			Issuer:    j.cfg.Issuer,
			ID:        uuid.NewString(),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(accessExp),
		},
		Username: sub.Username,
		Email:    sub.Email,
		TeamIDs:  sub.TeamIDs,
		Roles:    sub.Roles,
		Extra:    sub.Extra,
		TokenUse: "access",
	})
	if err != nil {
		return TokenPair{},
			err
	}

	refresh, err := j.signClaims(jwtClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   sub.UserID,
			Issuer:    j.cfg.Issuer,
			ID:        uuid.NewString(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(refreshExp),
		},
		// refresh 也带完整 Subject, 避免 refresh → Issue 链路丢失 username/email/roles/teams。
		Username: sub.Username,
		Email:    sub.Email,
		TeamIDs:  sub.TeamIDs,
		Roles:    sub.Roles,
		Extra:    sub.Extra,
		TokenUse: "refresh",
	})
	if err != nil {
		return TokenPair{}, err
	}
	return TokenPair{
		AccessToken:           access,
		RefreshToken:          refresh,
		AccessTokenExpiresAt:  accessExp,
		RefreshTokenExpiresAt: refreshExp,
		TokenType:             "Bearer",
	}, nil
}

// Verify 解析并校验 access token。
func (j *JWTIssuer) Verify(_ context.Context, accessToken string) (Subject, error) {
	claims, err := j.parse(accessToken)
	if err != nil {
		return Subject{}, err
	}
	if claims.TokenUse != "" && claims.TokenUse != "access" {
		return Subject{}, ErrInvalidToken
	}
	return j.subjectFromClaims(claims), nil
}

// Refresh 用 refresh token 换取新的 token pair
func (j *JWTIssuer) Refresh(ctx context.Context, refreshToken string) (TokenPair, error) {
	claims, err := j.parse(refreshToken)
	if err != nil {
		return TokenPair{}, err
	}
	if claims.TokenUse != "refresh" {
		return TokenPair{}, ErrInvalidToken
	}
	sub := j.subjectFromClaims(claims)
	return j.Issue(ctx, sub)
}

// Revoke 默认实现不存储黑名单,留待业务层结合 Redis 实现。
func (j *JWTIssuer) Revoke(_ context.Context, _ string) error {
	return nil
}

func (j *JWTIssuer) signClaims(c jwtClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	return token.SignedString([]byte(j.cfg.JWTSecret))
}

func (j *JWTIssuer) parse(raw string) (*jwtClaims, error) {
	if j.cfg.JWTSecret == "" {
		return nil, errors.New("auth.jwt_secret is empty")
	}
	claims := &jwtClaims{}
	tok, err := jwt.ParseWithClaims(raw, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(j.cfg.JWTSecret), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrInvalidToken
	}
	if !tok.Valid {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

func (j *JWTIssuer) subjectFromClaims(c *jwtClaims) Subject {
	sub := Subject{
		UserID:   c.Subject,
		Username: c.Username,
		Email:    c.Email,
		TeamIDs:  c.TeamIDs,
		Roles:    c.Roles,
		Issuer:   c.Issuer,
		Extra:    c.Extra,
	}
	if c.ExpiresAt != nil {
		sub.ExpiresAt = c.ExpiresAt.Time
	}
	return sub
}

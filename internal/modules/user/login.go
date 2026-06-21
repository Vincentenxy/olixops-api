package user

import (
	"context"
	"strings"
	"time"

	"olixops/internal/platform/auth"
	"olixops/pkg/errs"
)

// LoginInput 是账号密码登录入参。
type LoginInput struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	IP       string `json:"-"` // 由 handler 从 gin.ClientIP() 注入
}

// RefreshInput 是用 refresh token 换新 pair 的入参。
type RefreshInput struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// LoginResult 是登录成功返回。
type LoginResult struct {
	User   *User          `json:"user"`
	Tokens auth.TokenPair `json:"tokens"`
}

// Login 账号密码登录, 校验通过后签发 access + refresh。
//
// 失败一律返回 errs.Unauthorized("invalid credentials"),
// 不区分"用户不存在"和"密码错误", 防止账号枚举。
func (s *Service) Login(ctx context.Context, in LoginInput) (LoginResult, error) {
	in.Username = strings.TrimSpace(in.Username)
	u, err := s.repo.FindByUsername(ctx, in.Username)
	if err != nil {
		// 用户不存在也用 invalid credentials, 防枚举
		if isNotFound(err) {
			return LoginResult{}, errs.Unauthorized("invalid credentials")
		}
		return LoginResult{}, err
	}
	if !u.IsActive() {
		return LoginResult{}, errs.Forbidden("user is not active")
	}
	if err := s.hasher.Verify(in.Password, u.PasswordHash); err != nil {
		return LoginResult{}, errs.Unauthorized("invalid credentials")
	}

	// 部分字段更新最后登录信息; 失败不影响登录结果, 仅记日志
	if uerr := s.repo.UpdateLastLogin(ctx, u.ID, in.IP, time.Now()); uerr != nil {
		// 不阻塞登录流程, 业务上可降级
		_ = uerr
	}

	tokens, err := s.issuer.Issue(ctx, subjectFromUser(u))
	if err != nil {
		return LoginResult{}, errs.Internal("issue tokens: %v", err)
	}
	return LoginResult{User: u, Tokens: tokens}, nil
}

// Refresh 用 refresh token 换新 pair (rotate)。
// 当前实现不撤销旧 refresh (no-op), 依赖 token TTL 限制泄露窗口; 后续接 Redis 黑名单。
func (s *Service) Refresh(ctx context.Context, in RefreshInput) (auth.TokenPair, error) {
	pair, err := s.issuer.Refresh(ctx, in.RefreshToken)
	if err != nil {
		return auth.TokenPair{}, errs.FromAuthError(err)
	}
	return pair, nil
}

// Logout 第一阶段 noop: JWTIssuer.Revoke 是 stub。
// 后续阶段接 Redis 黑名单时, 这里把 token jti 写入黑名单。
func (s *Service) Logout(_ context.Context, _ string) error {
	return nil
}

// Me 从 ctx 取出 Subject, 返回对应用户完整信息。
func (s *Service) Me(ctx context.Context) (*User, error) {
	sub, ok := auth.FromContext(ctx)
	if !ok {
		return nil, errs.Unauthorized("no subject in context")
	}
	return s.repo.FindByID(ctx, sub.UserID)
}

// subjectFromUser 把 user 转换为 JWT Subject。
// TODO(rbac): Roles / TeamIDs 暂时留空, 等 RBAC 阶段接 role / team 表后填充。
func subjectFromUser(u *User) auth.Subject {
	return auth.Subject{
		UserID:   u.ID,
		Username: u.Username,
		Email:    u.Email,
		Issuer:   "olixops",
		Extra: map[string]string{
			"status": string(u.Status),
			"source": u.Source,
		},
	}
}

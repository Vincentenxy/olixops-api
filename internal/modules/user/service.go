package user

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"olixops/internal/platform/auth"
	"olixops/pkg/errs"
)

// CreateInput 是创建用户的入参。
type CreateInput struct {
	Username    string `json:"username" binding:"required,min=3,max=64"`
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=8"`
	DisplayName string `json:"display_name" binding:"max=128"`
	PhoneNumber string `json:"phone_number" binding:"max=32"`
}

// UpdateInput 是更新用户的入参,所有字段可选。
type UpdateInput struct {
	DisplayName *string `json:"display_name,omitempty" binding:"omitempty,max=128"`
	PhoneNumber *string `json:"phone_number,omitempty" binding:"omitempty,max=32"`
	AvatarURL   *string `json:"avatar_url,omitempty" binding:"omitempty,url"`
	Status      *Status `json:"status,omitempty"`
}

// ChangePasswordInput 是修改密码入参。
type ChangePasswordInput struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

// Service 是用户领域服务。
type Service struct {
	repo   Repository
	hasher auth.PasswordHasher
	issuer auth.TokenIssuer
}

// NewService 构造服务。
// issuer 用于 Login / Refresh 签发 token pair; 第一阶段必须有, 后续 OAuth 接入时复用。
func NewService(repo Repository, hasher auth.PasswordHasher, issuer auth.TokenIssuer) *Service {
	return &Service{
		repo:   repo,
		hasher: hasher,
		issuer: issuer,
	}
}

// Create 创建用户。
func (s *Service) Create(ctx context.Context, in CreateInput) (*User, error) {
	in.Username = strings.TrimSpace(in.Username)
	in.Email = strings.ToLower(strings.TrimSpace(in.Email))

	if existing, err := s.repo.FindByUsername(ctx, in.Username); err == nil && existing != nil {
		return nil, errs.AlreadyExists("username")
	} else if err != nil && !isNotFound(err) {
		return nil, err
	}
	if in.Email != "" {
		if existing, err := s.repo.FindByEmail(ctx, in.Email); err == nil && existing != nil {
			return nil, errs.AlreadyExists("email")
		} else if err != nil && !isNotFound(err) {
			return nil, err
		}
	}

	hash, err := s.hasher.Hash(in.Password)
	if err != nil {
		return nil, errs.Internal("hash password: %v", err)
	}
	u := &User{
		ID:           uuid.NewString(),
		Username:     in.Username,
		Email:        in.Email,
		DisplayName:  defaultIfEmpty(in.DisplayName, in.Username),
		PasswordHash: hash,
		PhoneNumber:  in.PhoneNumber,
		Status:       StatusActive,
		Source:       "local",
	}
	if err := s.repo.Create(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

// Get 按 ID 查询。
func (s *Service) Get(ctx context.Context, id string) (*User, error) {
	return s.repo.FindByID(ctx, id)
}

// List 列表查询。
func (s *Service) List(ctx context.Context, filter ListFilter) ([]*User, int64, error) {
	return s.repo.List(ctx, filter)
}

// Update 更新基础信息。
func (s *Service) Update(ctx context.Context, id string, in UpdateInput) (*User, error) {
	u, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if in.DisplayName != nil {
		u.DisplayName = *in.DisplayName
	}
	if in.PhoneNumber != nil {
		u.PhoneNumber = *in.PhoneNumber
	}
	if in.AvatarURL != nil {
		u.AvatarURL = *in.AvatarURL
	}
	if in.Status != nil {
		u.Status = *in.Status
	}
	if err := s.repo.Update(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

// ChangePassword 修改密码。
func (s *Service) ChangePassword(ctx context.Context, id string, in ChangePasswordInput) error {
	u, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if err := s.hasher.Verify(in.OldPassword, u.PasswordHash); err != nil {
		return errs.Unauthorized("old password mismatch")
	}
	hash, err := s.hasher.Hash(in.NewPassword)
	if err != nil {
		return errs.Internal("hash password: %v", err)
	}
	u.PasswordHash = hash
	return s.repo.Update(ctx, u)
}

// Authenticate 用账户密码登录,校验通过返回用户。
func (s *Service) Authenticate(ctx context.Context, username, password string) (*User, error) {
	u, err := s.repo.FindByUsername(ctx, username)
	if err != nil {
		if isNotFound(err) {
			return nil, errs.Unauthorized("invalid credentials")
		}
		return nil, err
	}
	if !u.IsActive() {
		return nil, errs.Forbidden("user is not active")
	}
	if err := s.hasher.Verify(password, u.PasswordHash); err != nil {
		return nil, errs.Unauthorized("invalid credentials")
	}
	now := time.Now()
	u.LastLoginAt = &now
	_ = s.repo.Update(ctx, u)
	return u, nil
}

// Delete 软删除用户。
func (s *Service) Delete(ctx context.Context, id string) error {
	if _, err := s.repo.FindByID(ctx, id); err != nil {
		return err
	}
	return s.repo.Delete(ctx, id)
}

func defaultIfEmpty(v, fallback string) string {
	if strings.TrimSpace(v) == "" {
		return fallback
	}
	return v
}

func isNotFound(err error) bool {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return true
	}
	e := errs.As(err)
	return e != nil && e.Code == errs.CodeNotFound
}

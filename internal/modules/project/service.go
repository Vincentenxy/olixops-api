package project

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"olixops/pkg/errs"
	"olixops/pkg/pagination"
)

var keyRegex = regexp.MustCompile(`^[a-z][a-z0-9-]{2,63}$`)

// CreateInput 是创建项目的入参。
type CreateInput struct {
	Key          string     `json:"key" binding:"required"`
	Name         string     `json:"name" binding:"required,min=2,max=128"`
	Description  string     `json:"description"`
	OwnerID      string     `json:"owner_id" binding:"required,uuid"`
	BusinessUnit string     `json:"business_unit"`
	RepoURL      string     `json:"repo_url" binding:"omitempty,url"`
	Visibility   Visibility `json:"visibility"`
	Labels       Labels     `json:"labels"`
}

// UpdateInput 是更新项目的入参。
type UpdateInput struct {
	Name         *string     `json:"name,omitempty" binding:"omitempty,min=2,max=128"`
	Description  *string     `json:"description,omitempty"`
	OwnerID      *string     `json:"owner_id,omitempty" binding:"omitempty,uuid"`
	BusinessUnit *string     `json:"business_unit,omitempty"`
	RepoURL      *string     `json:"repo_url,omitempty" binding:"omitempty,url"`
	Visibility   *Visibility `json:"visibility,omitempty"`
	Status       *Status     `json:"status,omitempty"`
	Labels       Labels      `json:"labels,omitempty"`
}

// AddMemberInput 是添加成员的入参。
type AddMemberInput struct {
	UserID    string `json:"user_id" binding:"required,uuid"`
	Role      string `json:"role" binding:"required"`
	InvitedBy string `json:"invited_by" binding:"omitempty,uuid"`
}

// Service 是项目领域服务。
type Service struct {
	repo Repository
}

// NewService 构造服务。
func NewService(repo Repository) *Service { return &Service{repo: repo} }

// Create 创建项目。
func (s *Service) Create(ctx context.Context, in CreateInput) (*Project, error) {
	in.Key = strings.ToLower(strings.TrimSpace(in.Key))
	if !keyRegex.MatchString(in.Key) {
		return nil, errs.InvalidArg("key must match %s", keyRegex.String())
	}
	if existing, err := s.repo.FindByKey(ctx, in.Key); err == nil && existing != nil {
		return nil, errs.AlreadyExists("project key")
	} else if err != nil && !isNotFound(err) {
		return nil, err
	}
	visibility := in.Visibility
	if visibility == "" {
		visibility = VisibilityPrivate
	}
	p := &Project{
		ID:           uuid.NewString(),
		Key:          in.Key,
		Name:         strings.TrimSpace(in.Name),
		Description:  in.Description,
		OwnerID:      in.OwnerID,
		BusinessUnit: in.BusinessUnit,
		RepoURL:      in.RepoURL,
		Visibility:   visibility,
		Status:       StatusActive,
		Labels:       in.Labels,
	}
	if err := s.repo.Create(ctx, p); err != nil {
		return nil, err
	}
	// owner 默认作为 project.owner。
	_ = s.repo.AddMember(ctx, &Member{
		ID:        uuid.NewString(),
		ProjectID: p.ID,
		UserID:    p.OwnerID,
		Role:      "project.owner",
		JoinedAt:  time.Now(),
	})
	return p, nil
}

// Get 按 ID 查询。
func (s *Service) Get(ctx context.Context, id string) (*Project, error) {
	return s.repo.FindByID(ctx, id)
}

// GetByKey 按 key 查询。
func (s *Service) GetByKey(ctx context.Context, key string) (*Project, error) {
	return s.repo.FindByKey(ctx, key)
}

// List 列表查询。
func (s *Service) List(ctx context.Context, q pagination.Query, filter ListFilter) ([]*Project, int64, error) {
	return s.repo.List(ctx, q, filter)
}

// Update 更新项目。
func (s *Service) Update(ctx context.Context, id string, in UpdateInput) (*Project, error) {
	p, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if in.Name != nil {
		p.Name = *in.Name
	}
	if in.Description != nil {
		p.Description = *in.Description
	}
	if in.OwnerID != nil {
		p.OwnerID = *in.OwnerID
	}
	if in.BusinessUnit != nil {
		p.BusinessUnit = *in.BusinessUnit
	}
	if in.RepoURL != nil {
		p.RepoURL = *in.RepoURL
	}
	if in.Visibility != nil {
		p.Visibility = *in.Visibility
	}
	if in.Status != nil {
		p.Status = *in.Status
	}
	if in.Labels != nil {
		p.Labels = in.Labels
	}
	if err := s.repo.Update(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

// Delete 软删除项目。
func (s *Service) Delete(ctx context.Context, id string) error {
	if _, err := s.repo.FindByID(ctx, id); err != nil {
		return err
	}
	return s.repo.Delete(ctx, id)
}

// AddMember 增加项目成员。
func (s *Service) AddMember(ctx context.Context, projectID string, in AddMemberInput) (*Member, error) {
	if _, err := s.repo.FindByID(ctx, projectID); err != nil {
		return nil, err
	}
	m := &Member{
		ID:        uuid.NewString(),
		ProjectID: projectID,
		UserID:    in.UserID,
		Role:      in.Role,
		InvitedBy: in.InvitedBy,
		JoinedAt:  time.Now(),
	}
	if err := s.repo.AddMember(ctx, m); err != nil {
		return nil, err
	}
	return m, nil
}

// RemoveMember 移除项目成员。
func (s *Service) RemoveMember(ctx context.Context, projectID, userID string) error {
	return s.repo.RemoveMember(ctx, projectID, userID)
}

// ListMembers 列出成员。
func (s *Service) ListMembers(ctx context.Context, projectID string) ([]*Member, error) {
	if _, err := s.repo.FindByID(ctx, projectID); err != nil {
		return nil, err
	}
	return s.repo.ListMembers(ctx, projectID)
}

func isNotFound(err error) bool {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return true
	}
	e := errs.As(err)
	return e != nil && e.Code == errs.CodeNotFound
}

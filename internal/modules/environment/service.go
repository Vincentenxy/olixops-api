package environment

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"olixops/pkg/errs"
	"olixops/pkg/pagination"
)

// CreateInput 创建环境入参。
type CreateInput struct {
	ProjectID     string `json:"project_id" binding:"required,uuid"`
	Code          string `json:"code" binding:"required,min=2,max=64"`
	Name          string `json:"name" binding:"required,min=2,max=128"`
	Type          Type   `json:"type" binding:"required,oneof=dev test sim prod"`
	Description   string `json:"description"`
	ClusterID     string `json:"cluster_id" binding:"omitempty,uuid"`
	Namespace     string `json:"namespace"`
	ApprovalLevel int    `json:"approval_level" binding:"min=0,max=10"`
	Order         int    `json:"order" binding:"min=0"`
}

// UpdateInput 更新环境入参。
type UpdateInput struct {
	Name          *string `json:"name,omitempty" binding:"omitempty,min=2,max=128"`
	Description   *string `json:"description,omitempty"`
	ClusterID     *string `json:"cluster_id,omitempty" binding:"omitempty,uuid"`
	Namespace     *string `json:"namespace,omitempty"`
	ApprovalLevel *int    `json:"approval_level,omitempty" binding:"omitempty,min=0,max=10"`
	Order         *int    `json:"order,omitempty" binding:"omitempty,min=0"`
	Status        *Status `json:"status,omitempty"`
}

// Service 是环境领域服务。
type Service struct {
	repo Repository
}

// NewService 构造服务。
func NewService(repo Repository) *Service { return &Service{repo: repo} }

// Create 创建环境。
func (s *Service) Create(ctx context.Context, in CreateInput) (*Environment, error) {
	in.Code = strings.ToLower(strings.TrimSpace(in.Code))
	if existing, err := s.repo.FindByCode(ctx, in.ProjectID, in.Code); err == nil && existing != nil {
		return nil, errs.AlreadyExists("environment code")
	} else if err != nil && !isNotFound(err) {
		return nil, err
	}
	e := &Environment{
		ID:            uuid.NewString(),
		ProjectID:     in.ProjectID,
		Code:          in.Code,
		Name:          strings.TrimSpace(in.Name),
		Type:          in.Type,
		Description:   in.Description,
		ClusterID:     in.ClusterID,
		Namespace:     in.Namespace,
		ApprovalLevel: in.ApprovalLevel,
		Order:         in.Order,
		Status:        StatusActive,
	}
	if err := s.repo.Create(ctx, e); err != nil {
		return nil, err
	}
	return e, nil
}

// Get 按 ID 查询。
func (s *Service) Get(ctx context.Context, id string) (*Environment, error) {
	return s.repo.FindByID(ctx, id)
}

// List 列表查询。
func (s *Service) List(ctx context.Context, q pagination.Query, filter ListFilter) ([]*Environment, int64, error) {
	return s.repo.List(ctx, q, filter)
}

// Update 更新环境。
func (s *Service) Update(ctx context.Context, id string, in UpdateInput) (*Environment, error) {
	e, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if in.Name != nil {
		e.Name = *in.Name
	}
	if in.Description != nil {
		e.Description = *in.Description
	}
	if in.ClusterID != nil {
		e.ClusterID = *in.ClusterID
	}
	if in.Namespace != nil {
		e.Namespace = *in.Namespace
	}
	if in.ApprovalLevel != nil {
		e.ApprovalLevel = *in.ApprovalLevel
	}
	if in.Order != nil {
		e.Order = *in.Order
	}
	if in.Status != nil {
		e.Status = *in.Status
	}
	if err := s.repo.Update(ctx, e); err != nil {
		return nil, err
	}
	return e, nil
}

// Delete 软删除环境。
func (s *Service) Delete(ctx context.Context, id string) error {
	if _, err := s.repo.FindByID(ctx, id); err != nil {
		return err
	}
	return s.repo.Delete(ctx, id)
}

func isNotFound(err error) bool {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return true
	}
	e := errs.As(err)
	return e != nil && e.Code == errs.CodeNotFound
}

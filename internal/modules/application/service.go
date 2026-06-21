package application

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"olixops/pkg/errs"
	"olixops/pkg/pagination"
)

// CreateInput 创建应用入参。
type CreateInput struct {
	ProjectID   string  `json:"project_id" binding:"required,uuid"`
	Key         string  `json:"key" binding:"required,min=2,max=64"`
	Name        string  `json:"name" binding:"required,min=2,max=128"`
	Description string  `json:"description"`
	Kind        Kind    `json:"kind" binding:"required,oneof=service frontend job agent sandbox"`
	Runtime     Runtime `json:"runtime"`
	RepoURL     string  `json:"repo_url" binding:"omitempty,url"`
	OwnerID     string  `json:"owner_id" binding:"omitempty,uuid"`
}

// UpdateInput 更新应用入参。
type UpdateInput struct {
	Name        *string  `json:"name,omitempty" binding:"omitempty,min=2,max=128"`
	Description *string  `json:"description,omitempty"`
	Runtime     *Runtime `json:"runtime,omitempty"`
	RepoURL     *string  `json:"repo_url,omitempty" binding:"omitempty,url"`
	OwnerID     *string  `json:"owner_id,omitempty" binding:"omitempty,uuid"`
	Status      *Status  `json:"status,omitempty"`
}

// ComponentInput 是组件入参。
type ComponentInput struct {
	Name          string `json:"name" binding:"required,min=1,max=128"`
	Workload      string `json:"workload" binding:"required"`
	Image         string `json:"image"`
	Port          int    `json:"port" binding:"omitempty,min=1,max=65535"`
	HealthCheck   string `json:"health_check"`
	Replicas      int    `json:"replicas" binding:"min=0,max=10000"`
	CPURequest    string `json:"cpu_request"`
	CPULimit      string `json:"cpu_limit"`
	MemoryRequest string `json:"memory_request"`
	MemoryLimit   string `json:"memory_limit"`
}

// Service 是应用领域服务。
type Service struct {
	repo Repository
}

// NewService 构造服务。
func NewService(repo Repository) *Service { return &Service{repo: repo} }

// Create 创建应用。
func (s *Service) Create(ctx context.Context, in CreateInput) (*Application, error) {
	in.Key = strings.ToLower(strings.TrimSpace(in.Key))
	if existing, err := s.repo.FindByKey(ctx, in.ProjectID, in.Key); err == nil && existing != nil {
		return nil, errs.AlreadyExists("application key")
	} else if err != nil && !isNotFound(err) {
		return nil, err
	}
	a := &Application{
		ID:          uuid.NewString(),
		ProjectID:   in.ProjectID,
		Key:         in.Key,
		Name:        strings.TrimSpace(in.Name),
		Description: in.Description,
		Kind:        in.Kind,
		Runtime:     in.Runtime,
		RepoURL:     in.RepoURL,
		OwnerID:     in.OwnerID,
		Status:      StatusActive,
	}
	if err := s.repo.Create(ctx, a); err != nil {
		return nil, err
	}
	return a, nil
}

// Get 按 ID 查询。
func (s *Service) Get(ctx context.Context, id string) (*Application, error) {
	return s.repo.FindByID(ctx, id)
}

// List 列表查询。
func (s *Service) List(ctx context.Context, q pagination.Query, filter ListFilter) ([]*Application, int64, error) {
	return s.repo.List(ctx, q, filter)
}

// Update 更新应用。
func (s *Service) Update(ctx context.Context, id string, in UpdateInput) (*Application, error) {
	a, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if in.Name != nil {
		a.Name = *in.Name
	}
	if in.Description != nil {
		a.Description = *in.Description
	}
	if in.Runtime != nil {
		a.Runtime = *in.Runtime
	}
	if in.RepoURL != nil {
		a.RepoURL = *in.RepoURL
	}
	if in.OwnerID != nil {
		a.OwnerID = *in.OwnerID
	}
	if in.Status != nil {
		a.Status = *in.Status
	}
	if err := s.repo.Update(ctx, a); err != nil {
		return nil, err
	}
	return a, nil
}

// Delete 软删除应用。
func (s *Service) Delete(ctx context.Context, id string) error {
	if _, err := s.repo.FindByID(ctx, id); err != nil {
		return err
	}
	return s.repo.Delete(ctx, id)
}

// AddComponent 添加组件。
func (s *Service) AddComponent(ctx context.Context, applicationID string, in ComponentInput) (*ServiceComponent, error) {
	if _, err := s.repo.FindByID(ctx, applicationID); err != nil {
		return nil, err
	}
	c := &ServiceComponent{
		ID:            uuid.NewString(),
		ApplicationID: applicationID,
		Name:          in.Name,
		Workload:      in.Workload,
		Image:         in.Image,
		Port:          in.Port,
		HealthCheck:   in.HealthCheck,
		Replicas:      in.Replicas,
		CPURequest:    in.CPURequest,
		CPULimit:      in.CPULimit,
		MemoryRequest: in.MemoryRequest,
		MemoryLimit:   in.MemoryLimit,
	}
	if err := s.repo.CreateComponent(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

// ListComponents 列出组件。
func (s *Service) ListComponents(ctx context.Context, applicationID string) ([]*ServiceComponent, error) {
	if _, err := s.repo.FindByID(ctx, applicationID); err != nil {
		return nil, err
	}
	return s.repo.ListComponents(ctx, applicationID)
}

// DeleteComponent 删除组件。
func (s *Service) DeleteComponent(ctx context.Context, id string) error {
	return s.repo.DeleteComponent(ctx, id)
}

func isNotFound(err error) bool {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return true
	}
	e := errs.As(err)
	return e != nil && e.Code == errs.CodeNotFound
}

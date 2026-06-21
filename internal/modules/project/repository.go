package project

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"olixops/internal/platform/database"
	"olixops/pkg/errs"
	"olixops/pkg/pagination"
)

// Repository 是项目持久化接口。
type Repository interface {
	Create(ctx context.Context, p *Project) error
	Update(ctx context.Context, p *Project) error
	Delete(ctx context.Context, id string) error
	FindByID(ctx context.Context, id string) (*Project, error)
	FindByKey(ctx context.Context, key string) (*Project, error)
	List(ctx context.Context, q pagination.Query, filter ListFilter) ([]*Project, int64, error)

	AddMember(ctx context.Context, m *Member) error
	RemoveMember(ctx context.Context, projectID, userID string) error
	ListMembers(ctx context.Context, projectID string) ([]*Member, error)
}

// ListFilter 是项目列表过滤条件。
type ListFilter struct {
	OwnerID    string
	Status     Status
	Visibility Visibility
}

type gormRepository struct {
	db *gorm.DB
}

// NewRepository 构造默认仓储。
func NewRepository(db *gorm.DB) Repository { return &gormRepository{db: db} }

func (r *gormRepository) tx(ctx context.Context) *gorm.DB {
	return database.FromContext(ctx, r.db)
}

func (r *gormRepository) Create(ctx context.Context, p *Project) error {
	return r.tx(ctx).Create(p).Error
}

func (r *gormRepository) Update(ctx context.Context, p *Project) error {
	return r.tx(ctx).Save(p).Error
}

func (r *gormRepository) Delete(ctx context.Context, id string) error {
	return r.tx(ctx).Delete(&Project{}, "id = ?", id).Error
}

func (r *gormRepository) FindByID(ctx context.Context, id string) (*Project, error) {
	var p Project
	if err := r.tx(ctx).First(&p, "id = ?", id).Error; err != nil {
		return nil, mapErr(err)
	}
	return &p, nil
}

func (r *gormRepository) FindByKey(ctx context.Context, key string) (*Project, error) {
	var p Project
	if err := r.tx(ctx).First(&p, "key = ?", key).Error; err != nil {
		return nil, mapErr(err)
	}
	return &p, nil
}

func (r *gormRepository) List(ctx context.Context, q pagination.Query, filter ListFilter) ([]*Project, int64, error) {
	tx := r.tx(ctx).Model(&Project{})
	if filter.OwnerID != "" {
		tx = tx.Where("owner_id = ?", filter.OwnerID)
	}
	if filter.Status != "" {
		tx = tx.Where("status = ?", filter.Status)
	}
	if filter.Visibility != "" {
		tx = tx.Where("visibility = ?", filter.Visibility)
	}
	if q.Keyword != "" {
		tx = tx.Where("name ILIKE ? OR key ILIKE ?", "%"+q.Keyword+"%", "%"+q.Keyword+"%")
	}
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	orderBy := "created_at"
	if q.OrderBy != "" {
		orderBy = q.OrderBy
	}
	var items []*Project
	if err := tx.Order(orderBy + " " + q.Order).Offset(q.Offset()).Limit(q.Limit()).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *gormRepository) AddMember(ctx context.Context, m *Member) error {
	return r.tx(ctx).Create(m).Error
}

func (r *gormRepository) RemoveMember(ctx context.Context, projectID, userID string) error {
	return r.tx(ctx).Where("project_id = ? AND user_id = ?", projectID, userID).Delete(&Member{}).Error
}

func (r *gormRepository) ListMembers(ctx context.Context, projectID string) ([]*Member, error) {
	var ms []*Member
	if err := r.tx(ctx).Where("project_id = ?", projectID).Find(&ms).Error; err != nil {
		return nil, err
	}
	return ms, nil
}

func mapErr(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errs.NotFound("project")
	}
	return err
}

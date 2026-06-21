package environment

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"olixops/internal/platform/database"
	"olixops/pkg/errs"
	"olixops/pkg/pagination"
)

// Repository 是环境持久化接口。
type Repository interface {
	Create(ctx context.Context, e *Environment) error
	Update(ctx context.Context, e *Environment) error
	Delete(ctx context.Context, id string) error
	FindByID(ctx context.Context, id string) (*Environment, error)
	FindByCode(ctx context.Context, projectID, code string) (*Environment, error)
	List(ctx context.Context, q pagination.Query, filter ListFilter) ([]*Environment, int64, error)
}

// ListFilter 列表过滤条件。
type ListFilter struct {
	ProjectID string
	Type      Type
	Status    Status
}

type gormRepository struct {
	db *gorm.DB
}

// NewRepository 构造默认仓储。
func NewRepository(db *gorm.DB) Repository { return &gormRepository{db: db} }

func (r *gormRepository) tx(ctx context.Context) *gorm.DB {
	return database.FromContext(ctx, r.db)
}

func (r *gormRepository) Create(ctx context.Context, e *Environment) error {
	return r.tx(ctx).Create(e).Error
}

func (r *gormRepository) Update(ctx context.Context, e *Environment) error {
	return r.tx(ctx).Save(e).Error
}

func (r *gormRepository) Delete(ctx context.Context, id string) error {
	return r.tx(ctx).Delete(&Environment{}, "id = ?", id).Error
}

func (r *gormRepository) FindByID(ctx context.Context, id string) (*Environment, error) {
	var e Environment
	if err := r.tx(ctx).First(&e, "id = ?", id).Error; err != nil {
		return nil, mapErr(err)
	}
	return &e, nil
}

func (r *gormRepository) FindByCode(ctx context.Context, projectID, code string) (*Environment, error) {
	var e Environment
	if err := r.tx(ctx).First(&e, "project_id = ? AND code = ?", projectID, code).Error; err != nil {
		return nil, mapErr(err)
	}
	return &e, nil
}

func (r *gormRepository) List(ctx context.Context, q pagination.Query, filter ListFilter) ([]*Environment, int64, error) {
	tx := r.tx(ctx).Model(&Environment{})
	if filter.ProjectID != "" {
		tx = tx.Where("project_id = ?", filter.ProjectID)
	}
	if filter.Type != "" {
		tx = tx.Where("type = ?", filter.Type)
	}
	if filter.Status != "" {
		tx = tx.Where("status = ?", filter.Status)
	}
	if q.Keyword != "" {
		tx = tx.Where("name ILIKE ? OR code ILIKE ?", "%"+q.Keyword+"%", "%"+q.Keyword+"%")
	}
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	orderBy := `"order"`
	if q.OrderBy != "" {
		orderBy = q.OrderBy
	}
	var items []*Environment
	if err := tx.Order(orderBy + " " + q.Order).Offset(q.Offset()).Limit(q.Limit()).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func mapErr(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errs.NotFound("environment")
	}
	return err
}

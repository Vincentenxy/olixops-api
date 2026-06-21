package application

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"olixops/internal/platform/database"
	"olixops/pkg/errs"
	"olixops/pkg/pagination"
)

// Repository 是应用持久化接口。
type Repository interface {
	Create(ctx context.Context, a *Application) error
	Update(ctx context.Context, a *Application) error
	Delete(ctx context.Context, id string) error
	FindByID(ctx context.Context, id string) (*Application, error)
	FindByKey(ctx context.Context, projectID, key string) (*Application, error)
	List(ctx context.Context, q pagination.Query, filter ListFilter) ([]*Application, int64, error)

	CreateComponent(ctx context.Context, c *ServiceComponent) error
	UpdateComponent(ctx context.Context, c *ServiceComponent) error
	DeleteComponent(ctx context.Context, id string) error
	ListComponents(ctx context.Context, applicationID string) ([]*ServiceComponent, error)
}

// ListFilter 过滤条件。
type ListFilter struct {
	ProjectID string
	Kind      Kind
	Status    Status
	OwnerID   string
}

type gormRepository struct {
	db *gorm.DB
}

// NewRepository 构造默认仓储。
func NewRepository(db *gorm.DB) Repository { return &gormRepository{db: db} }

func (r *gormRepository) tx(ctx context.Context) *gorm.DB {
	return database.FromContext(ctx, r.db)
}

func (r *gormRepository) Create(ctx context.Context, a *Application) error {
	return r.tx(ctx).Create(a).Error
}

func (r *gormRepository) Update(ctx context.Context, a *Application) error {
	return r.tx(ctx).Save(a).Error
}

func (r *gormRepository) Delete(ctx context.Context, id string) error {
	return r.tx(ctx).Delete(&Application{}, "id = ?", id).Error
}

func (r *gormRepository) FindByID(ctx context.Context, id string) (*Application, error) {
	var a Application
	if err := r.tx(ctx).First(&a, "id = ?", id).Error; err != nil {
		return nil, mapErr(err)
	}
	return &a, nil
}

func (r *gormRepository) FindByKey(ctx context.Context, projectID, key string) (*Application, error) {
	var a Application
	if err := r.tx(ctx).First(&a, "project_id = ? AND key = ?", projectID, key).Error; err != nil {
		return nil, mapErr(err)
	}
	return &a, nil
}

func (r *gormRepository) List(ctx context.Context, q pagination.Query, filter ListFilter) ([]*Application, int64, error) {
	tx := r.tx(ctx).Model(&Application{})
	if filter.ProjectID != "" {
		tx = tx.Where("project_id = ?", filter.ProjectID)
	}
	if filter.Kind != "" {
		tx = tx.Where("kind = ?", filter.Kind)
	}
	if filter.Status != "" {
		tx = tx.Where("status = ?", filter.Status)
	}
	if filter.OwnerID != "" {
		tx = tx.Where("owner_id = ?", filter.OwnerID)
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
	var items []*Application
	if err := tx.Order(orderBy + " " + q.Order).Offset(q.Offset()).Limit(q.Limit()).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *gormRepository) CreateComponent(ctx context.Context, c *ServiceComponent) error {
	return r.tx(ctx).Create(c).Error
}

func (r *gormRepository) UpdateComponent(ctx context.Context, c *ServiceComponent) error {
	return r.tx(ctx).Save(c).Error
}

func (r *gormRepository) DeleteComponent(ctx context.Context, id string) error {
	return r.tx(ctx).Delete(&ServiceComponent{}, "id = ?", id).Error
}

func (r *gormRepository) ListComponents(ctx context.Context, applicationID string) ([]*ServiceComponent, error) {
	var items []*ServiceComponent
	if err := r.tx(ctx).Where("application_id = ?", applicationID).Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func mapErr(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errs.NotFound("application")
	}
	return err
}

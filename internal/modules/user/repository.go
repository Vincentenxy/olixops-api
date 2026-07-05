package user

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"olixops/internal/platform/database"
	"olixops/pkg/errs"
	"olixops/pkg/pagination"
)

// Repository 是用户持久化接口
type Repository interface {
	Create(ctx context.Context, u *User) error
	Update(ctx context.Context, u *User) error
	Delete(ctx context.Context, id string) error
	FindByID(ctx context.Context, id string) (*User, error)
	FindByUsername(ctx context.Context, username string) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByExternal(ctx context.Context, source, externalID string) (*User, error)
	UpdateLastLogin(ctx context.Context, id, ip string, at time.Time) error
	List(ctx context.Context, filter ListFilter) ([]*User, int64, error)
}

// gormRepository 是 GORM 实现。
type gormRepository struct {
	db *gorm.DB
}

var _ Repository = (*gormRepository)(nil)

func (r *gormRepository) tx(ctx context.Context) *gorm.DB {
	return database.FromContext(ctx, r.db)
}

func (r *gormRepository) Create(ctx context.Context, u *User) error {
	return r.tx(ctx).Create(u).Error
}

func (r *gormRepository) Update(ctx context.Context, u *User) error {
	return r.tx(ctx).Save(u).Error
}

func (r *gormRepository) Delete(ctx context.Context, id string) error {
	return r.tx(ctx).Delete(&User{}, "id = ?", id).Error
}

func (r *gormRepository) FindByID(ctx context.Context, id string) (*User, error) {
	var u User
	if err := r.tx(ctx).First(&u, "id = ?", id).Error; err != nil {
		return nil, mapErr(err)
	}
	return &u, nil
}

func (r *gormRepository) FindByUsername(ctx context.Context, username string) (*User, error) {
	var u User
	if err := r.tx(ctx).First(&u, "username = ?", username).Error; err != nil {
		return nil, mapErr(err)
	}
	return &u, nil
}

func (r *gormRepository) FindByEmail(ctx context.Context, email string) (*User, error) {
	var u User
	if err := r.tx(ctx).First(&u, "email = ?", email).Error; err != nil {
		return nil, mapErr(err)
	}
	return &u, nil
}

func (r *gormRepository) FindByExternal(ctx context.Context, source, externalID string) (*User, error) {
	var u User
	err := r.tx(ctx).First(&u, "source = ? AND external_id = ?", source, externalID).Error
	if err != nil {
		return nil, mapErr(err)
	}
	return &u, nil
}

func (r *gormRepository) UpdateLastLogin(ctx context.Context, id, ip string, at time.Time) error {
	// 只更新两个字段, 避免与并发 Update 产生全行覆盖竞争。
	return r.tx(ctx).Model(&User{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"last_login_at": at,
			"last_login_ip": ip,
		}).Error
}

// ListFilter 是用户列表过滤条件。
type ListFilter struct {
	pagination.Query
	Status  Status `json:"status"`
	Source  string `json:"source"`
	Keyword string `json:"keyword"`
	OrderBy string `json:"order_by"`
	Order   string `json:"order"`
}

func (r *gormRepository) List(ctx context.Context, filter ListFilter) ([]*User, int64, error) {
	tx := r.tx(ctx).Model(&User{})
	if filter.Status != "" {
		tx = tx.Where("status = ?", filter.Status)
	}
	if filter.Source != "" {
		tx = tx.Where("source = ?", filter.Source)
	}
	if filter.Keyword != "" {
		tx = tx.Where("username ILIKE ? OR email ILIKE ? OR display_name ILIKE ?",
			"%"+filter.Keyword+"%", "%"+filter.Keyword+"%", "%"+filter.Keyword+"%")
	}
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	orderBy := "created_at"
	if filter.OrderBy != "" {
		orderBy = filter.OrderBy
	}
	var items []*User
	if err := tx.Order(orderBy + " " + filter.Order).
		Offset(filter.Offset()).
		Limit(filter.Limit()).
		Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func mapErr(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errs.NotFound("user")
	}
	return err
}

// NewRepository 构造默认仓储
func NewRepository(db *gorm.DB) Repository {
	return &gormRepository{
		db: db,
	}
}

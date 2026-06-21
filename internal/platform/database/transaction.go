package database

import (
	"context"

	"gorm.io/gorm"
)

type txKey struct{}

// Tx 用 context 传递事务句柄,任何 Repository 通过 FromContext(ctx, defaultDB) 取得正确的 *gorm.DB。
type Tx struct {
	db *gorm.DB
}

// NewTx 构造事务执行器。
func NewTx(db *gorm.DB) *Tx {
	return &Tx{db: db}
}

// Execute 在事务中执行 fn。fn 内部应通过 database.FromContext(ctx) 获取事务连接。
func (t *Tx) Execute(ctx context.Context, fn func(ctx context.Context) error) error {
	return t.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		ctx := context.WithValue(ctx, txKey{}, tx)
		return fn(ctx)
	})
}

// FromContext 返回 context 中的事务连接,缺省返回 fallback。
func FromContext(ctx context.Context, fallback *gorm.DB) *gorm.DB {
	if ctx == nil {
		return fallback
	}
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok && tx != nil {
		return tx
	}
	return fallback.WithContext(ctx)
}

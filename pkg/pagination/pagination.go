// Package pagination 抽象列表分页参数。
package pagination

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	defaultPage     = 1
	defaultPageSize = 20
	maxPageSize     = 200
)

// Query 是分页查询参数。
type Query struct {
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	Keyword  string `form:"keyword"`
	OrderBy  string `form:"order_by"`
	Order    string `form:"order"` // asc / desc
}

// Normalize 将参数限制在合理范围。
func (q *Query) Normalize() {
	if q.Page <= 0 {
		q.Page = defaultPage
	}
	if q.PageSize <= 0 {
		q.PageSize = defaultPageSize
	}
	if q.PageSize > maxPageSize {
		q.PageSize = maxPageSize
	}
	if q.Order != "desc" {
		q.Order = "asc"
	}
}

// Offset 计算 SQL OFFSET。
func (q Query) Offset() int {
	if q.Page <= 1 {
		return 0
	}
	return (q.Page - 1) * q.PageSize
}

// Limit 返回 SQL LIMIT。
func (q Query) Limit() int { return q.PageSize }

// FromGin 从 Gin context 解析参数,Bind 失败时退回默认值。
func FromGin(c *gin.Context) Query {
	q := Query{}
	if err := c.ShouldBindQuery(&q); err != nil {
		page, _ := strconv.Atoi(c.Query("page"))
		size, _ := strconv.Atoi(c.Query("page_size"))
		q.Page = page
		q.PageSize = size
		q.Keyword = c.Query("keyword")
		q.OrderBy = c.Query("order_by")
		q.Order = c.Query("order")
	}
	q.Normalize()
	return q
}

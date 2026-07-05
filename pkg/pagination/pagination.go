// Package pagination 抽象列表分页基础参数，仅提供页码、页大小能力
package pagination

import (
	"github.com/gin-gonic/gin"
)

const (
	defaultPageNum  = 1
	defaultPageSize = 20
	maxPageSize     = 200
)

type Query struct {
	PageNum  int `json:"pageNum" form:"pageNum" binding:"min=1"`
	PageSize int `json:"pageSize" form:"pageSize" binding:"min=1,max=100"`
}

func (q *Query) Normalize() {
	if q.PageNum < defaultPageNum {
		q.PageNum = defaultPageNum
	}
	if q.PageSize <= 0 {
		q.PageSize = defaultPageSize
	}
	if q.PageSize > maxPageSize {
		q.PageSize = maxPageSize
	}
}

func (q Query) Offset() int {
	if q.PageNum <= 1 {
		return 0
	}
	return (q.PageNum - 1) * q.PageSize
}

func (q Query) Limit() int {
	return q.PageSize
}

// Get 请求使用
// FromGin 从gin url query解析分页参数，解析失败自动使用默认值
func FromGin(c *gin.Context) *Query {
	var q Query
	_ = c.ShouldBindQuery(&q)
	q.Normalize()
	return &q
}

type PageResult[T any] struct {
	Total int64 `json:"total"`
	List  []T   `json:"list"`
}

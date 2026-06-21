package httpx

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 是统一响应体,严格只含 code / msg / data 三个字段。
type Response[T any] struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data T      `json:"data,omitempty"`
}

// PagedData 包装分页返回结构。
type PagedData struct {
	Items    any   `json:"items"`
	Total    int64 `json:"total"`
	PageNum  int   `json:"pageNum"`
	PageSize int   `json:"pageSize"`
}

func (resp *Response[T]) IsSuccess() bool {
	return resp.Code == Ok
}

func (resp *Response[T]) IsFailed() bool {
	return resp.Code != Ok
}

// 成功响应
func OK[T any](c *gin.Context, data T) {
	OKWithMsg(c, metaMap[Ok].msg, data)
}

func OKWithMsg[T any](c *gin.Context, msg string, data T) {
	c.JSON(http.StatusOK, Response[T]{
		Code: Ok,
		Msg:  msg,
		Data: data,
	})
}

// 失败响应
func Fail(c *gin.Context, code int, err error) {
	c.JSON(http.StatusOK, Response[any]{
		Code: code,
		Msg:  err.Error(),
		Data: nil,
	})
}

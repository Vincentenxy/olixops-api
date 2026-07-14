package httpx

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 是统一响应体,严格只含 code / msg / data 三个字段。
type Response[T any] struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data T      `json:"data"`
}

// PagedData 包装分页返回结构。
type PagedData struct {
	Items    any   `json:"items"`
	Total    int64 `json:"total"`
	PageNum  int   `json:"pageNum"`
	PageSize int   `json:"pageSize"`
}

func (resp *Response[T]) IsSuccess() bool {
	return resp.Code == CodeSuccess
}

func (resp *Response[T]) IsFailed() bool {
	return resp.Code != CodeSuccess
}

// 成功响应
func OK[T any](c *gin.Context, data T) {
	OKWithMsg(c, metaMap[CodeSuccess].msg, data)
}

// 自定义消息成功响应
func OKWithMsg[T any](c *gin.Context, msg string, data T) {
	c.JSON(http.StatusOK, Response[T]{
		Code: CodeSuccess,
		Msg:  msg,
		Data: data,
	})
}

// 失败响应
func Fail(c *gin.Context, err error) {
	FailWithCode(c, CodeFail, err)
}

// 带错误码的失败响应
func FailWithCode(c *gin.Context, code int, err error) {
	c.JSON(http.StatusOK, Response[any]{
		Code: code,
		Msg:  err.Error(),
		Data: nil,
	})
}

func Unauthorized(c *gin.Context) {
	c.JSON(http.StatusUnauthorized, Response[any]{
		Code: CodeUnauthorized,
		Msg:  "Unauthorized",
		Data: nil,
	})
}

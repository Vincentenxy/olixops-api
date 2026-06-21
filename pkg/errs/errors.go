// Package errs 提供应用统一错误模型。
//
// 业务层返回的错误统一是 *Error,HTTP 层据此映射 HTTP 状态码与响应体。
// 同一个错误可以用 .WithCause 链式附带原始错误,日志记录时会保留完整 stack。
package errs

import (
	"errors"
	"fmt"
	"net/http"
)

// Code 是业务错误码。
type Code string

const (
	CodeInternal         Code = "INTERNAL"
	CodeInvalidArg       Code = "INVALID_ARGUMENT"
	CodeUnauthorized     Code = "UNAUTHORIZED"
	CodeForbidden        Code = "FORBIDDEN"
	CodeNotFound         Code = "NOT_FOUND"
	CodeAlreadyExists    Code = "ALREADY_EXISTS"
	CodeConflict         Code = "CONFLICT"
	CodeRateLimited      Code = "RATE_LIMITED"
	CodeUnavailable      Code = "UNAVAILABLE"
	CodeDependencyFail   Code = "DEPENDENCY_FAILED"
	CodePreconditionFail Code = "PRECONDITION_FAILED"
)

// Error 是统一错误结构。
type Error struct {
	Code    Code
	Message string
	Cause   error
	Details map[string]any
}

// Error 实现 error 接口。
func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap 返回原始错误,使 errors.Is/As 可用。
func (e *Error) Unwrap() error { return e.Cause }

// WithCause 链接原始错误。
func (e *Error) WithCause(err error) *Error {
	e.Cause = err
	return e
}

// WithDetails 附加结构化信息。
func (e *Error) WithDetails(details map[string]any) *Error {
	e.Details = details
	return e
}

// Newf 创建错误。
func Newf(code Code, format string, args ...any) *Error {
	return &Error{Code: code, Message: fmt.Sprintf(format, args...)}
}

// Wrap 包装为指定 code 的错误。
func Wrap(err error, code Code, message string) *Error {
	if err == nil {
		return nil
	}
	return &Error{Code: code, Message: message, Cause: err}
}

// As 将 err 转换为 *Error;如果原本不是,则返回内部错误包装。
func As(err error) *Error {
	if err == nil {
		return nil
	}
	var e *Error
	if errors.As(err, &e) {
		return e
	}
	return &Error{Code: CodeInternal, Message: err.Error(), Cause: err}
}

// HTTPStatus 把错误码映射成 HTTP 状态码。
func (e *Error) HTTPStatus() int {
	if e == nil {
		return http.StatusOK
	}
	switch e.Code {
	case CodeInvalidArg:
		return http.StatusBadRequest
	case CodeUnauthorized:
		return http.StatusUnauthorized
	case CodeForbidden:
		return http.StatusForbidden
	case CodeNotFound:
		return http.StatusNotFound
	case CodeAlreadyExists, CodeConflict:
		return http.StatusConflict
	case CodeRateLimited:
		return http.StatusTooManyRequests
	case CodePreconditionFail:
		return http.StatusPreconditionFailed
	case CodeUnavailable, CodeDependencyFail:
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}

// 常用快捷构造方法

func InvalidArg(format string, args ...any) *Error { return Newf(CodeInvalidArg, format, args...) }
func Unauthorized(msg string) *Error               { return Newf(CodeUnauthorized, "%s", msg) }
func Forbidden(msg string) *Error                  { return Newf(CodeForbidden, "%s", msg) }
func NotFound(resource string) *Error              { return Newf(CodeNotFound, "%s not found", resource) }
func AlreadyExists(resource string) *Error {
	return Newf(CodeAlreadyExists, "%s already exists", resource)
}
func Internal(format string, args ...any) *Error { return Newf(CodeInternal, format, args...) }

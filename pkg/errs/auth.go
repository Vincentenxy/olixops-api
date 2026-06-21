package errs

import (
	"errors"

	"olixops/internal/platform/auth"
)

// FromAuthError 把 auth 包的 sentinel error 映射为 envelope 友好的 *Error。
//
// 中间件和 handler 共用, 避免到处写 if err == ErrTokenExpired。
// 未识别的 error 兜底为 CodeUnauthorized, 防止泄露内部细节。
func FromAuthError(err error) *Error {
	switch {
	case errors.Is(err, auth.ErrTokenExpired):
		return Unauthorized("token expired")
	case errors.Is(err, auth.ErrInvalidToken):
		return Unauthorized("invalid token")
	case errors.Is(err, auth.ErrInvalidPassword):
		return Unauthorized("invalid credentials")
	case errors.Is(err, auth.ErrUserNotFound):
		return Unauthorized("user not found")
	case errors.Is(err, auth.ErrProviderUnready):
		return &Error{Code: CodeUnavailable, Message: "oauth provider not configured"}
	default:
		return Wrap(err, CodeUnauthorized, "auth failure")
	}
}

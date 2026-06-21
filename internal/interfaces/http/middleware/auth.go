package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"olixops/internal/platform/auth"
	"olixops/pkg/errs"
	"olixops/pkg/httpx"
)

const (
	authHeader   = "Authorization"
	bearerPrefix = "Bearer "
	bearerMinLen = len(bearerPrefix)
)

// Auth 返回 gin 中间件: 校验 Authorization: Bearer <token>,
// 验证通过后把 auth.Subject 注入到 request context (供下游 handler/service 使用)。
//
// 失败统一返回 401 + envelope { code: UNAUTHORIZED, msg, data }。
// errs.FromAuthError 负责把 auth 包的 sentinel error 翻译成对前端友好的文案。
//
// 挂载要求: 必须挂在 Trace 中间件之后, 这样 logger 能拿到 trace-id 与未授权请求关联。
func Auth(issuer auth.TokenIssuer) gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader(authHeader)
		if len(h) < bearerMinLen || !strings.EqualFold(h[:bearerMinLen], bearerPrefix) {
			httpx.Error(c, errs.Unauthorized("missing bearer token"))
			return
		}
		token := strings.TrimSpace(h[bearerMinLen:])
		if token == "" {
			httpx.Error(c, errs.Unauthorized("missing bearer token"))
			return
		}

		sub, err := issuer.Verify(c.Request.Context(), token)
		if err != nil {
			httpx.Error(c, errs.FromAuthError(err))
			return
		}

		ctx := auth.WithSubject(c.Request.Context(), sub)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

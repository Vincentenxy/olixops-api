package middleware

import (
	"olixops/internal/platform/logger"
	"strings"

	"olixops/internal/platform/auth"
	"olixops/pkg/errs"
	"olixops/pkg/httpx"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
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
func Auth(issuer auth.TokenIssuer, cookieManager *auth.CookieManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		var token string

		l := logger.L()
		l.Info("token", zap.String("token", c.Request.Header.Get(authHeader)))

		// get token from header
		h := c.GetHeader(authHeader)
		if len(h) >= bearerMinLen && strings.EqualFold(h[:bearerMinLen], bearerPrefix) {
			token = strings.TrimSpace(h[bearerMinLen:])
		}

		// get token from cookie if not exist in header
		if token == "" {
			cookieToken, err := cookieManager.Get(c, "access_token")
			if err == nil && cookieToken != "" {
				token = cookieToken
			}
		}

		// not login
		if token == "" {
			httpx.Fail(c, errs.Unauthorized("missing token"))
			c.Abort()
			return
		}

		sub, err := issuer.Verify(c.Request.Context(), token)
		if err != nil {
			httpx.Fail(c, errs.FromAuthError(err))
			c.Abort()
			return
		}

		ctx := auth.WithSubject(c.Request.Context(), sub)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

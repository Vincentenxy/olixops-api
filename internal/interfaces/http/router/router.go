// Package router 负责把各业务模块的路由按统一契约组装到 gin.Engine。
//
// 路径约定 (按 [[api-path-layout]] 规范):
//   - /api/pub/v1/...   公开接口,免鉴权
//   - /api/v1/...       内部接口,需鉴权 (priv 分组挂载 Auth 中间件)
//
// 中间件加载顺序 (按调用链先后):
//  1. Trace       — 注入 trace_id, 必须最先
//  2. AccessLog   — 记录每次请求耗时 / status
//  3. Auth (priv) — JWT 校验, 把 Subject 注入 ctx
//  4. (后续) Recovery / Cors / RateLimit 等
package router

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"olixops/internal/interfaces/http/middleware"
	"olixops/internal/platform/auth"
	"olixops/internal/platform/logger"
)

// New 根据已注册模块构造 gin.Engine。
// issuer 用于在 priv 分组挂 Auth 中间件; pub 分组不挂。
func New(modules []ModuleRoutes, issuer auth.TokenIssuer, cookieManager *auth.CookieManager) *gin.Engine {
	r := gin.New()

	// 全局中间件
	r.Use(middleware.Trace())
	r.Use(middleware.AccessLog())

	pub := r.Group("/api/pub")
	pri := r.Group("/api")
	pri.Use(middleware.Auth(issuer, cookieManager))

	for _, m := range modules {
		logger.L().Info("register http module",
			zap.String("module", m.Name()),
		)
		m.RegisterPub(pub)
		m.RegisterPrivate(pri)
	}
	return r
}

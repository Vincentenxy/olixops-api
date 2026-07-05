// Package user 模块对外暴露路由注册入口 (router.ModuleRoutes)。
//
// 路径设计 (按 [[api-path-layout]] / [[api-http-methods]]):
//
//	/api/pub/v1/auth/login     POST  免鉴权登录
//	/api/pub/v1/auth/refresh   POST  用 refresh 换新 token
//	/api/v1/auth/me            GET   当前登录用户
//	/api/v1/auth/logout        POST  登出 (第一阶段 noop)
//	/api/v1/users              POST  创建 / GET 列表
//	/api/v1/users/:id          GET   详情
//	/api/v1/users/:id/update   POST  更新
//	/api/v1/users/:id/delete   POST  删除
//	/api/v1/users/:id/password POST  修改密码
//
// 注意: 严禁 PUT / DELETE, [[api-http-methods]] 规范禁止。
package user

import (
	"olixops/internal/interfaces/http/router"
	"olixops/internal/platform/auth"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"olixops/internal/platform/audit"
)

// Module 是 user 模块对外的路由注册入口, 实现 router.ModuleRoutes。
type Module struct {
	h        *Handler
	recorder audit.Recorder
}

// NewModule 构造 user 模块
func NewModule(h *Handler, recorder audit.Recorder) *Module {
	return &Module{
		h:        h,
		recorder: recorder,
	}
}

func Assemble(db *gorm.DB, recoder audit.Recorder, issuer *auth.JWTIssuer, cookieManager *auth.CookieManager) (*Module, error) {
	serv := NewService(
		NewRepository(db),
		auth.NewBcryptHasher(10),
		issuer,
	)
	handler := NewHandler(serv, recoder, cookieManager)
	return NewModule(handler, recoder), nil
}

// Name 返回模块名, 用于启动日志
func (m *Module) Name() string { return "user" }

// RegisterPub 挂载免鉴权路由到 /api/pub/v1
func (m *Module) RegisterPub(g *gin.RouterGroup) {

	v1 := g.Group("/v1")
	{
		v1.POST("/login", m.h.login)
		v1.POST("/register", m.h.register)

		auth := v1.Group("/auth")
		{
			auth.POST("/refresh", m.h.refresh)
		}

	}

}

// RegisterPrivate 挂载需鉴权路由到 /api/v1
func (m *Module) RegisterPrivate(g *gin.RouterGroup) {

	v1 := g.Group("/v1")
	{
		v1.GET("/logout", m.h.logout)

		// auth
		auth := v1.Group("/auth")
		{
			auth.GET("/me", m.h.me)
		}

		user := v1.Group("/user")
		{

			user.POST("/create", m.h.create)
			user.GET("/list", m.h.list)
		}

	}

}

// 编译期断言: Module 必须满足 router.ModuleRoutes
var _ router.ModuleRoutes = (*Module)(nil)

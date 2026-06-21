package router

import "github.com/gin-gonic/gin"

// ModuleRoutes 是所有 HTTP 模块必须实现的路由注册契约。
//
// 一个模块对应一个业务域 (例如 user / project / health),
// 通过实现 RegisterPub / RegisterPrivate 把本模块的路由挂到对应的 v1 子路由组上。
//
//   - RegisterPub    挂到 /api/pub/v1    免鉴权接口 (healthz、login、register 等)
//   - RegisterPrivate 挂到 /api/v1        需鉴权接口 (CRUD、业务操作等)
//
// 任一方法允许为空实现,表示该模块没有对应分类的接口。
//
// 未来中间件栈 (recovery / cors / accesslog / trace-id / auth) 在 router 层面统一挂载,
// 模块只关心自己的路由声明。
type ModuleRoutes interface {
	// Name 返回模块名,用于启动日志与调试,例如 "health" / "user"。
	Name() string
	RegisterPub(g *gin.RouterGroup)
	RegisterPrivate(g *gin.RouterGroup)
}

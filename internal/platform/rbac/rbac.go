// Package rbac 描述基于角色的权限模型。
//
// 资源以 "domain:resource" 命名,例如 "project:application"。
// 动作以 verb 表达,例如 "read"、"write"、"deploy"。
// 后续可扩展 ABAC,通过自定义 PolicyEvaluator 实现。
package rbac

import (
	"context"
	"errors"
)

var (
	ErrPermissionDenied = errors.New("rbac: permission denied")
	ErrRoleNotFound     = errors.New("rbac: role not found")
)

// 预定义内置角色。业务层可以扩展。
const (
	RoleSystemAdmin  = "system.admin"
	RolePlatformOps  = "platform.ops"
	RoleProjectOwner = "project.owner"
	RoleProjectDev   = "project.dev"
	RoleProjectView  = "project.view"
	RoleGuest        = "guest"
)

// Action 是权限动作枚举。
type Action string

const (
	ActionRead    Action = "read"
	ActionWrite   Action = "write"
	ActionDelete  Action = "delete"
	ActionDeploy  Action = "deploy"
	ActionRelease Action = "release"
	ActionApprove Action = "approve"
	ActionAdmin   Action = "admin"
)

// Resource 表示一个可授权的资源。
type Resource struct {
	Type    string            // 资源类型,如 application、cluster、release_order
	ID      string            // 资源 ID,可为空,代表对该类型的全局动作
	Project string            // 所属项目 ID
	Env     string            // 所属环境
	Labels  map[string]string // 扩展标签,供 ABAC 使用
}

// Decision 是权限判断结果。
type Decision struct {
	Allow  bool
	Reason string
}

// Enforcer 是权限判定器,业务层只依赖这个接口。
type Enforcer interface {
	Check(ctx context.Context, action Action, resource Resource) (Decision, error)
}

// EnforcerFunc 让函数也能实现 Enforcer。
type EnforcerFunc func(ctx context.Context, action Action, resource Resource) (Decision, error)

func (f EnforcerFunc) Check(ctx context.Context, action Action, resource Resource) (Decision, error) {
	return f(ctx, action, resource)
}

// AllowAll 用于本地开发或单元测试。
func AllowAll() Enforcer {
	return EnforcerFunc(func(context.Context, Action, Resource) (Decision, error) {
		return Decision{Allow: true, Reason: "allow_all"}, nil
	})
}

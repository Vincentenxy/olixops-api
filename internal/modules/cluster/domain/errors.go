package domain

import "errors"

// 领域错误 (sentinel errors), service 层用 errors.Is 判断后转成 errs.*。
//
// 按 [[module-boundaries]] 约束, 业务模块不应该 import olixops-api/pkg/errs 来定义错误;
// 应该在每个模块自己定义, 然后在 service 出口处映射。
var (
	// ErrClusterNotFound 集群不存在。
	ErrClusterNotFound = errors.New("cluster: not found")

	// ErrClusterAlreadyExists 集群名称重复。
	ErrClusterAlreadyExists = errors.New("cluster: already exists")

	// ErrKubeconfigInvalid kubeconfig 解析失败或 apiserver 不可达。
	ErrKubeconfigInvalid = errors.New("cluster: kubeconfig invalid or apiserver unreachable")

	// ErrNamespaceNotFound namespace 不存在 (在指定集群)。
	ErrNamespaceNotFound = errors.New("namespace: not found")

	// ErrNamespaceAlreadyExists namespace 已存在。
	ErrNamespaceAlreadyExists = errors.New("namespace: already exists")
)

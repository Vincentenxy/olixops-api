package domain

import "fmt"

// 领域错误码, 与 pkg/errs 配合使用。
//
// handler 层收到领域错误后, 通过 errs.As() 提取 code 返回给前端。

// ErrRepoNotFound 仓库不存在。
type ErrRepoNotFound struct{ ID string }

func (e *ErrRepoNotFound) Error() string { return fmt.Sprintf("helm repo not found: %s", e.ID) }

// ErrRepoAlreadyExists 仓库名称重复。
type ErrRepoAlreadyExists struct{ Name string }

func (e *ErrRepoAlreadyExists) Error() string {
	return fmt.Sprintf("helm repo already exists: %s", e.Name)
}

// ErrReleaseNotFound Release 不存在。
type ErrReleaseNotFound struct{ ID string }

func (e *ErrReleaseNotFound) Error() string { return fmt.Sprintf("helm release not found: %s", e.ID) }

// ErrReleaseAlreadyExists 同名 Release 在该 namespace 已存在。
type ErrReleaseAlreadyExists struct{ Name, Namespace string }

func (e *ErrReleaseAlreadyExists) Error() string {
	return fmt.Sprintf("helm release %q already exists in namespace %q", e.Name, e.Namespace)
}

// ErrChartNotFound 在仓库中找不到指定的 Chart。
type ErrChartNotFound struct{ Repo, Chart string }

func (e *ErrChartNotFound) Error() string {
	return fmt.Sprintf("chart %q not found in repo %q", e.Chart, e.Repo)
}

// ErrChartVersionNotFound 找不到指定的 Chart 版本。
type ErrChartVersionNotFound struct{ Chart, Version string }

func (e *ErrChartVersionNotFound) Error() string {
	return fmt.Sprintf("chart %q version %q not found", e.Chart, e.Version)
}

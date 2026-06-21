// Package audit 描述统一审计日志。
//
// 所有"生产敏感操作"必须通过 Recorder.Record 落库,以便事后追溯。
package audit

import (
	"context"
	"time"
)

// Event 是一次审计事件。
type Event struct {
	ID         string            `json:"id"`
	OccurredAt time.Time         `json:"occurred_at"`
	Actor      string            `json:"actor"`       // 操作人 user id
	ActorName  string            `json:"actor_name"`  // 操作人名称
	Action     string            `json:"action"`      // 操作动作,如 application.deploy
	Resource   string            `json:"resource"`    // 资源类型
	ResourceID string            `json:"resource_id"` // 资源 ID
	Project    string            `json:"project"`     // 所属项目
	Env        string            `json:"env"`         // 所属环境
	IP         string            `json:"ip"`
	UserAgent  string            `json:"user_agent"`
	Status     string            `json:"status"` // success / failure
	Reason     string            `json:"reason"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

// Recorder 是审计落地器。
type Recorder interface {
	Record(ctx context.Context, ev Event) error
}

// RecorderFunc 让函数也能实现 Recorder。
type RecorderFunc func(ctx context.Context, ev Event) error

func (f RecorderFunc) Record(ctx context.Context, ev Event) error {
	return f(ctx, ev)
}

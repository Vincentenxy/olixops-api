package audit

import (
	"context"
	"sync"

	"go.uber.org/zap"
)

// MemRecorder 进程内审计记录器:
//   - 维护一个固定容量的环形缓冲, 满了之后覆盖最旧的 (FIFO)
//   - 每次 Record 同时通过 zap 输出结构化日志, 便于生产环境聚合查询
//
// 设计目的: 第一阶段不接数据库也能跑通审计; 测试可直接断言 Events() 内容。
// 后续阶段接 GORM/ClickHouse 时, 实现新的 Recorder, 业务代码无需改动。
type MemRecorder struct {
	cap  int
	mu   sync.Mutex
	buf  []Event
	head int  // 下一个写入位置
	full bool // 缓冲是否已满过 (用于区分未填满状态)
	log  *zap.Logger
}

// NewMemRecorder 构造进程内 Recorder。
// cap <= 0 时默认 1024; log 可为 nil (nil 时不写 zap)。
func NewMemRecorder(cap int, log *zap.Logger) *MemRecorder {
	if cap <= 0 {
		cap = 1024
	}
	return &MemRecorder{
		cap: cap,
		buf: make([]Event, cap),
		log: log,
	}
}

// Record 记录一条审计事件。
func (r *MemRecorder) Record(_ context.Context, ev Event) error {
	r.mu.Lock()
	r.buf[r.head] = ev
	r.head = (r.head + 1) % r.cap
	if r.head == 0 {
		r.full = true
	}
	r.mu.Unlock()

	if r.log != nil {
		// 同时输出结构化日志, 生产环境可直接被 Loki/ES 聚合
		r.log.Info("audit",
			zap.String("action", ev.Action),
			zap.String("actor", ev.Actor),
			zap.String("actor_name", ev.ActorName),
			zap.String("resource", ev.Resource),
			zap.String("resource_id", ev.ResourceID),
			zap.String("project", ev.Project),
			zap.String("env", ev.Env),
			zap.String("ip", ev.IP),
			zap.String("user_agent", ev.UserAgent),
			zap.String("status", ev.Status),
			zap.String("reason", ev.Reason),
		)
	}
	return nil
}

// Events 返回当前缓冲里的事件 (按时间顺序, 最旧的在前)。
// 仅用于测试 / 调试, 不要在热路径调用。
func (r *MemRecorder) Events() []Event {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []Event
	if r.full {
		out = append(out, r.buf[r.head:]...)
		out = append(out, r.buf[:r.head]...)
	} else {
		out = append(out, r.buf[:r.head]...)
	}
	return out
}

// Len 返回当前缓冲里的事件数量。
func (r *MemRecorder) Len() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.full {
		return r.cap
	}
	return r.head
}

# Agent 模块设计 — Sandbox via Firecracker

> 状态: **设计稿**, 实施中
> 模块: `internal/modules/agent`
> 范围: agent runtime 下的沙箱生命周期管理, 后端用 Firecracker microVMM

---

## 1. 背景与目标

按 [[module-boundaries]] 设计, agent 模块最终会拆成独立服务 `olixops-agent-runtime`。
沙箱 (sandbox) 是 agent 运行时的核心能力 — agent 执行用户代码、调试、长跑任务都需要隔离环境。

需求:

- 通过 **Firecracker** microVMM 提供强隔离 (microVM 级别, 优于 container)
- HTTP API 形式管理沙箱生命周期: create / start / stop / delete / status
- 业务代码**不直接依赖** Firecracker SDK, 走 adapter 接口, 未来可换 runtime (gVisor / kata)

不在本文档范围 (后续任务):

- agent 应用注册 / Session / 工具调用审计 (agent 模块其他子能力)
- 网络 tap / CNI 集成
- 镜像存储 / 拉取

---

## 2. 架构分层

按 [[modular-monolith-layout]], agent 模块内部按 `domain / service / repository / handler / adapter` 切。sandbox 作为子资源复用同一套分层。

```
internal/modules/agent/
├── module.go                      # ModuleRoutes 实现
├── domain/
│   ├── sandbox.go                 # 实体: SandboxSpec / SandboxState / SandboxHandle
│   ├── session.go                 # (其他子资源)
│   └── errors.go                  # 领域错误
├── repository/
│   └── sandbox_repo.go            # 沙箱元数据持久化 (postgres)
├── service/
│   └── sandbox_service.go         # 编排: 调 adapter + repo
├── handler/
│   └── sandbox_handler.go         # /api/v1/sandboxes/* HTTP handler
└── adapter/                       # 外部依赖封装 (本设计核心)
    ├── sandbox_runtime.go         # interface: SandboxRuntime (业务概念)
    ├── firecracker/               # Firecracker 具体实现
    │   ├── client.go              # HTTP 客户端 (net/http + unix socket transport)
    │   ├── process.go             # firecracker 进程管理 (exec.Command + lifecycle)
    │   └── translator.go          # SandboxSpec ↔ Firecracker JSON 配置
    └── noop/                      # 测试用 noop 实现
        └── noop.go
```

**关键约束** ([[module-boundaries]]):

- service 只调 `adapter.SandboxRuntime` 接口, **不 import** firecracker SDK / net/http 调用细节
- adapter 实现持有 firecracker HTTP 客户端, 翻译业务概念 ↔ Firecracker 概念
- 未来换 runtime (gVisor / kata-containers) 改 adapter, 业务零改动

---

## 3. Firecracker HTTP API 特性

**关键事实**, 不要按普通 REST 设计:

| 项 | Firecracker 实际情况 |
|---|---|
| 传输 | **Unix domain socket** (默认 `/tmp/firecracker.sock`) 或 vsock, **不是 TCP** |
| 协议 | HTTP/1.1, 但**不是 RESTful** — 固定几个 endpoint |
| API 端点 | `PUT /machine-config`、`PUT /boot-source`、`PUT /actions`、`GET /actions` |
| 一个 VM 一个进程 | 每个 firecracker 进程 = 一个 microVM, 监听自己的 socket |
| 启动方式 | 我们 Go 服务**自己 fork firecracker binary** (`exec.Command`), 传 `--api-sock` 参数 |
| 错误格式 | 非 2xx 响应 + body 是错误描述 (plain text / JSON 视版本而定) |

生命周期时序:

```
我们 Go 服务                          Firecracker Process
─────────────                          ──────────────────
1. fork firecracker binary
   + --api-sock=/tmp/sb-xxx.sock
   + --config-file=... ─────────────►  进程启动, 监听 unix socket
                                       (此时 VM 未启动)

2. PUT /machine-config              ─►  保存 VM 配置
   { vcpu, memory, kernel, rootfs }

3. PUT /boot-source                 ─►  保存启动源

4. PUT /actions { ActionType: "InstanceStart" }
                                     ─►  VM 真正启动!

5. PUT /actions { ActionType: "SendCtrlResponse" }
                                     ─►  CtrlAltDel 优雅停止

6. 进程退出, 清理 socket 文件
```

---

## 4. 客户端技术选型

### 方案对比

| 方案 | 工作量 | 依赖 | 推荐 |
|---|---|---|---|
| **A. net/http + 自定义 Unix socket Transport** | 中 (~200 行) | 0 | **✅ 推荐** |
| B. firecracker-go-sdk 官方 SDK | 小 (~50 行) | +1 间接依赖 | 不推荐 |

### 选 A 的理由

1. **依赖最小** — 项目推崇少依赖; SDK 只省了 ~150 行调用代码, 不值得引入
2. **完全可控** — 接口适配层把 Firecracker 概念翻译成业务概念, SDK 不会帮你做这个翻译
3. **换 runtime 友好** — 改 adapter 实现, 业务零改动
4. **核心代码只有 30 行** — 自定义 `Transport.DialContext` 支持 unix socket

### net/http + unix socket 的核心代码

```go
// adapter/firecracker/client.go

func newTransport(socketPath string) http.RoundTripper {
    return &http.Transport{
        DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
            var d net.Dialer
            return d.DialContext(ctx, "unix", socketPath)
        },
        // Firecracker 是单连接复用, 短超时即可
        IdleConnTimeout: 30 * time.Second,
    }
}

// PUT helper (所有 firecracker 写操作都是 PUT)
func (c *Client) put(ctx context.Context, path string, body any) error {
    data, _ := json.Marshal(body)
    req, _ := http.NewRequestWithContext(ctx, http.MethodPut,
        "http://unix"+path, bytes.NewReader(data))
    req.Header.Set("Content-Type", "application/json")

    resp, err := c.http.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode >= 300 {
        body, _ := io.ReadAll(resp.Body)
        return fmt.Errorf("firecracker PUT %s: status=%d body=%s",
            path, resp.StatusCode, body)
    }
    return nil
}
```

---

## 5. 适配器接口设计 (核心契约)

**业务概念先行**, 不暴露 Firecracker 概念:

```go
// internal/modules/agent/adapter/sandbox_runtime.go
package adapter

import (
    "context"
    "io"
)

// SandboxRuntime 抽象出沙箱运行时操作, 与具体实现 (firecracker / gvisor / noop) 解耦。
type SandboxRuntime interface {
    // Create 启动一个沙箱进程 (fork firecracker binary + 返回 handle), 但 VM 还未启动。
    // spec 含 vcpu / 内存 / 镜像路径 / 网络 tap 等。
    Create(ctx context.Context, spec SandboxSpec) (*SandboxHandle, error)

    // Start 触发沙箱内 microVM 启动 (firecracker: PUT /actions InstanceStart)。
    Start(ctx context.Context, id string) error

    // Stop 优雅停止 (firecracker: PUT /actions SendCtrlResponse = CtrlAltDel)。
    // 超时后进程被 kill。
    Stop(ctx context.Context, id string, timeout time.Duration) error

    // Delete 销毁沙箱, 释放 socket 路径 / 清理进程资源。
    Delete(ctx context.Context, id string) error

    // Status 查询运行时状态。
    Status(ctx context.Context, id string) (SandboxState, error)
}

// SandboxSpec 描述一个沙箱的静态配置。
type SandboxSpec struct {
    Image       string            // 镜像文件路径 (host 上)
    VCPU        int               // vCPU 数量
    MemoryMB    int               // 内存 MB
    NetworkTap  string            // tap 设备名, 空表示无网络
    KernelArgs  string            // 内核启动参数
    Metadata    map[string]string // 透传给 VM 的元数据
    Env         []string          // 透传环境变量
}

// SandboxHandle 是 Create 之后得到的运行时句柄。
type SandboxHandle struct {
    ID      string // 我们自己生成的内部 ID (用于路由)
    Socket  string // 适配器内部的通信端点 (firecracker: unix socket path)
    PID     int    // 适配器进程 PID (firecracker: firecracker 进程 PID)
}

// SandboxState 是沙箱运行时的状态枚举。
type SandboxState string

const (
    SandboxStatePending SandboxState = "pending"  // 已 Create, 未 Start
    SandboxStateRunning SandboxState = "running"  // VM 正在运行
    SandboxStateStopped SandboxState = "stopped"  // 优雅停止
    SandboxStateFailed  SandboxState = "failed"   // 启动失败 / 异常退出
    SandboxStateUnknown SandboxState = "unknown"  // 查不到 (进程不在)
)
```

**关键设计**:

- `SandboxSpec` 是**业务配置**, 不含 firecracker 的 `boot-source`、`machine-config` 等内部概念
- `SandboxHandle.Socket` 是 adapter 内部细节, 业务代码不应该读它 (firecracker 用 unix socket, gvisor 可能用别的)
- `ID` 是我们生成的内部 ID, 不暴露 firecracker 的 `vm-id`

---

## 6. Firecracker 实现的关键映射

translator 把业务 `SandboxSpec` 翻译成 Firecracker 的 JSON 配置:

```
SandboxSpec                      Firecracker API
───────────                      ───────────────
Image        ── translate ──►    PUT /boot-source { kernel_image_path, boot_args }
                                 PUT /machine-config { vcpu_count, mem_size_mib, ... }
                                 PUT /drives/rootfs { drive_id, path_on_host, is_root_device }

VCPU         ──►                 PUT /machine-config.vcpu_count
MemoryMB     ──►                 PUT /machine-config.mem_size_mib
KernelArgs   ──►                 PUT /boot-source.boot_args
NetworkTap   ──►                 PUT /network-interfaces/{iface_id} { host_dev_name }
```

Create 流程:

```
1. fork firecracker binary
   firecracker --api-sock=/tmp/sb-<uuid>.sock --config-vm-...

2. 等待 socket 文件出现 (最多 5s 超时)

3. 翻译 SandboxSpec, 依次 PUT:
   - /machine-config
   - /boot-source
   - /drives/rootfs
   - /network-interfaces/* (如有)
   - /logger (可选)
   - /machine-options (可选)

4. PUT /actions { ActionType: "InstanceStart" }

5. 保留 {socketPath, PID}, 返回 SandboxHandle{ID: uuid}
```

注意: **第 3 步**必须在第 4 步之前, 否则 firecracker 会拒绝 Start。

---

## 7. 进程管理 (process.go)

`SandboxRuntime.Create` 必须自己 fork firecracker, 不能假设外面有进程。设计要点:

```go
// adapter/firecracker/process.go

type Process struct {
    Cmd    *exec.Cmd
    Socket string
    PID    int
}

// Start 启动 firecracker 进程, 阻塞直到 socket 就绪或超时。
func (p *Process) Start(ctx context.Context, socketPath, firecrackerBin string, vmID string) error {
    // 1. 准备 socket 路径 (临时目录, 进程退出时清理)
    // 2. exec.CommandContext(firecrackerBin, "--api-sock", socketPath, ...)
    // 3. 启动 goroutine 等待进程退出 (用于清理)
    // 4. 轮询 socket 文件是否出现, 超时 5s
    // 5. 返回 *Process
}

// Wait 阻塞等待进程退出, 返回 exit error。
func (p *Process) Wait() error

// Kill 强制 kill, 用于 Stop 超时或 Delete。
func (p *Process) Kill() error
```

**关键点**:

- 必须用 `exec.CommandContext` 让 ctx cancel 能触发 SIGKILL
- socket 文件用临时目录, 进程退出时清理 (避免残留)
- PID 需要记录, 用于 OOM / 异常时 kill

---

## 8. 错误码与状态机

| Adapter 层错误 | 业务层映射 | HTTP |
|---|---|---|
| 进程启动失败 (binary 不存在 / 权限不足) | `errs.CodeDependencyFail` | 503 |
| socket 超时未就绪 | `errs.CodeUnavailable` | 503 |
| PUT /machine-config 4xx (配置错误) | `errs.CodeInvalidArg` | 400 |
| PUT 5xx (firecracker 内部错误) | `errs.CodeDependencyFail` | 503 |
| 进程意外退出 | `errs.CodeUnavailable` | 503 |
| Status 查不到 (进程不在) | `SandboxStateUnknown` | 200 (业务查询返回 unknown 状态) |

状态机:

```
        Create                Start              进程退出/Stop
Pending ────────► (fail) ────► Running ──────────────────► Stopped
   │                ▲           │   ▲                       │
   │                │           │   │                       │
   │                └───────────┘   └─────── Stop ───────────┘
   │
   └─ Create 失败 ──► Failed
```

业务层应该捕获 `Failed` 状态, 不允许再 Start, 必须先 Delete 清理。

---

## 9. HTTP 接口列表

按 [[api-path-layout]] / [[api-http-methods]] 规范 (无参 GET / 带参 POST / 严禁 PUT/DELETE):

| Method | Path | Auth | 说明 |
|---|---|---|---|
| POST | `/api/v1/sandboxes` | 必填 | Create + Start 一体 (异步), 返回 sandbox id |
| GET | `/api/v1/sandboxes` | 必填 | 列表查询 (按状态过滤) |
| GET | `/api/v1/sandboxes/:id` | 必填 | 详情 + 实时状态 |
| POST | `/api/v1/sandboxes/:id/stop` | 必填 | 优雅停止 |
| POST | `/api/v1/sandboxes/:id/start` | 必填 | 已 stop 后重新启动 |
| POST | `/api/v1/sandboxes/:id/delete` | 必填 | 销毁 (Stop + 清理) |

> **不在本期范围**: POST /api/v1/sandboxes/:id/exec (执行命令), POST /api/v1/sandboxes/:id/snapshot (快照)
> 这些是后续能力, 需要先有 sandbox 生命周期稳了再做。

---

## 10. 配置项

`AuthConfig` 已有, 这里新增 `SandboxConfig` (放入 `internal/config/config.go`):

```yaml
sandbox:
  firecracker_bin: /usr/bin/firecracker     # firecracker 可执行文件路径
  socket_dir: /tmp/olixops/sandboxes      # unix socket 临时目录
  default_vcpu: 1
  default_memory_mb: 512
  default_kernel_image: /var/lib/olixops/vmlinux
  default_rootfs_dir: /var/lib/olixops/rootfs/
  create_timeout: 5s                      # socket 就绪超时
  stop_timeout: 10s                       # Stop 超时后强 kill
```

---

## 11. 实施路线 (8 步)

按依赖关系, 每步可独立 PR / 测试:

1. **定义 adapter interface** (`adapter/sandbox_runtime.go`) — 业务概念先行
2. **写 noop 实现** (`adapter/noop/noop.go`) — 测试 + 业务接口对齐
3. **写 firecracker 进程管理** (`adapter/firecracker/process.go`) — exec.Command + 生命周期
4. **写 firecracker HTTP 客户端** (`adapter/firecracker/client.go`) — net/http + unix socket transport
5. **写 translator** (`adapter/firecracker/translator.go`) — SandboxSpec → Firecracker JSON
6. **接 sandbox service** (`service/sandbox_service.go`) — 调 adapter interface
7. **接 sandbox handler** (`handler/sandbox_handler.go`) — 6 个 endpoint
8. **测试**: noop 路径 100% 覆盖; firecracker 路径用 httptest mock HTTP 客户端

每步完成后:

- `go test ./internal/modules/agent/...`
- 编译 `go build ./...`
- 重写 `module.go` 让 `agent` 模块满足 `router.ModuleRoutes` 契约 (按 user 模块范式)

---

## 12. 不在本文档范围

- **agent 应用注册 / Session / 工具调用审计** — agent 模块其他子能力, 后续单开设计
- **镜像存储 / 拉取** — 走 registry 模块, sandbox 直接接收 host 路径
- **网络 / CNI 集成** — tap 设备由 host 预先配置好, sandbox 不管
- **资源配额 / 成本统计** — 第二阶段
- **gVisor / kata-containers 适配** — 现在只 firecracker, 架构留好接口未来加
- **审计 / 计费 / 配额** — 调 `audit.Recorder` 接口, 业务层处理

---

## 13. 参考资料

- [Firecracker API docs](https://github.com/firecracker-microvm/firecracker/blob/main/api_server/swagger/firecracker.yaml)
- [[module-boundaries]] — 业务 / 适配器分层原则
- [[modular-monolith-layout]] — 模块内部分层规范
- [[api-path-layout]] / [[api-http-methods]] — 路由命名规范
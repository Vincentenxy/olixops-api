# Cluster 模块设计

> 状态: **骨架阶段** (skeleton + TODO, 实现待补)
> 模块: `internal/modules/cluster`
> 范围: K8s 集群注册、namespace 管理、节点查询、只读工作负载查询
> 后续可拆: `olixops-cluster`

---

## 1. 背景

按 [[module-boundaries]], cluster 模块最终会拆为独立服务 `olixops-cluster`。
本模块对外提供"已注册 K8s 集群"的统一视图，业务模块（application、deployment、helm）
通过本模块查询集群信息，不再各自维护 Kubeconfig。

## 2. 资源范围

| 资源 | 读写 | 说明 |
|---|---|---|
| Cluster | R/W | 集群元数据（名称、kubeconfig、状态），不连 K8s |
| Namespace | R/W | 通过 k8s client 实际管理 |
| Node | R | 只读查询 |
| Workload (Deployment/Service/Ingress) | R | 只读查询 |

> ⚠️ **写入型 workload 操作**（create deployment、apply yaml）应放到 `application` 或 `helm` 模块。

## 3. 架构分层

按 [[modular-monolith-layout]]:

```
internal/modules/cluster/
├── module.go                      # ModuleRoutes 实现
├── domain/                        # 实体 + 领域错误
├── repository/                    # 集群元数据持久化
├── service/                       # 业务编排
├── handler/                       # HTTP 入口
└── adapter/                       # k8s.io/client-go 封装
    ├── k8s_client.go              # interface
    ├── k8s/                       # 真实实现
    └── noop/                      # 测试用
```

业务代码（service）**只调 adapter interface**，不直接 import k8s SDK。

## 4. 实施路线

8 步独立 commit + 跑通, 见文件清单每个文件的 TODO 注释。

# OlixOps API

OlixOps API 是一个面向新一代 DevOps 平台的后端项目，目标是统一管理服务部署、智能体应用、Kubernetes 集群、Helm
Chart、镜像仓库、统一认证、流水线、环境发布和文件版本等能力。

项目初期建议采用“模块化单体”架构：所有能力先放在同一个代码仓库、同一个后端服务内实现，但从一开始就按领域模块、接口边界、数据模型和事件边界拆清楚。这样前期开发效率高，后续当某个模块的复杂度、团队边界或性能压力上来时，可以平滑拆成独立服务。

## 建设目标

- 统一入口：提供统一 API、统一认证、统一权限、统一审计。
- 多环境管理：支持 `dev`、`test`、`sim`、`prod` 等环境的资源、配置、发布和审批隔离。
- 多类型应用：支持普通服务、后台任务、前端应用、智能体应用、沙箱应用等不同部署形态。
- 云原生集成：管理 Kubernetes 集群、命名空间、工作负载、Helm Release、Ingress、Secret、ConfigMap 等资源。
- DevOps 闭环：打通需求、代码、构建、镜像、部署、发布、回滚、审计和文件版本。
- 可扩展插件化：为 Harbor、OAuth2、Tekton、Helm、Git、对象存储等外部系统预留适配层。

## 开发规范

- 接口规范
    - 除非分享连接的接口之外，其他所有不带参数的请求统一使用GET，带有参数的请求使用POST,不允许使用DELETE/PUT
    - 仅限分享类的接口，可以使用带有参数的get请求
- 命名规范
    - 统一使用驼峰命名法
- 请求返回规范
  - 统一使用如下返回体, **响应体严格只含 code / msg / data 三个字段, 不得包含 trace_id 或其他额外字段**
  ```json
  {
    "code": 0,  // 响应状态码
    "msg": "",  // 提示消息
    "data": {}  // 数据内容
  }
  ```
  - 响应状态码规范
    - 0: 成功
    - -1: 通用失败状态码
    - 1-1000: 与http状态码保持一致
    - 1000以上: 通用业务状态码，业务可以自定义
  - trace-id 不进响应体
    - 链路追踪 ID 通过 **response header `X-Trace-Id`** 回传给前端
    - 前端发请求时携带 `X-Request-Id` header; 后端若无则生成 UUID
    - trace-id 仅在 response header 与后端日志里流转, **严禁塞进 `data` 字段或 envelope 顶层**
  
- 日志规范
  - 全局统一使用日志打印
  - 日志打印格式需要对其，比如层级
  - 不同日志需要添加对应颜色区分
  - 日志需要打印trace-id

- 安全规范
  -

- 测试规范
  - 所有接口必须有测试用例
  - 测试用例需要涵盖所有的情况
  - 测试用例的出入参数都需要放在测试代码中，方便查看
  
- API设计规范
  - 路径设计：/api/[pub]/v1/业务/操作类型
  - pub: 表示无认证，不带pub表示需要认证

## 核心功能发散

### 1. 服务部署管理

面向普通业务服务的部署、升级、回滚和状态观测。

- 应用注册：服务名称、业务线、负责人、仓库地址、技术栈、运行时类型。
- 部署单元：Deployment、StatefulSet、DaemonSet、Job、CronJob、裸进程或外部服务。
- 配置管理：环境变量、ConfigMap、Secret、启动参数、资源限制。
- 发布策略：滚动发布、蓝绿发布、金丝雀发布、暂停/继续发布。
- 运行状态：副本数、Pod 状态、事件、日志入口、健康检查、最近发布时间。
- 回滚能力：按版本、镜像、Helm Release Revision 或发布记录回滚。

### 2. 智能体应用与沙箱管理

面向 Agent 应用、工具调用运行时、隔离执行环境和资源配额。

- Agent 应用注册：模型配置、工具列表、知识库绑定、运行入口、上下文策略。
- 沙箱模板：基础镜像、CPU/内存限制、网络策略、文件系统权限、生命周期。
- 会话管理：Agent Session、任务执行记录、工具调用日志、失败重试。
- 隔离策略：命名空间隔离、NetworkPolicy、临时卷、只读挂载、超时回收。
- 资源治理：并发限制、队列控制、配额统计、成本估算。
- 安全审计：工具调用审计、文件访问审计、外部网络访问审计。

### 3. Kubernetes 集群管理

统一纳管多个 Kubernetes 集群。

- 集群接入：Kubeconfig、ServiceAccount Token、集群连通性检测。
- 集群视图：Node、Namespace、Pod、Workload、Service、Ingress、PVC、Event。
- 命名空间治理：环境绑定、配额、标签、默认 LimitRange、网络策略。
- 多集群调度：不同环境或业务线绑定不同集群。
- 权限隔离：用户只能操作被授权的集群、命名空间和资源。
- 集群巡检：版本、节点健康、资源水位、异常事件、证书过期提醒。

### 4. Helm Chart 集成

支持 Chart 仓库、模板参数、Release 生命周期管理。

- Chart 仓库接入：OCI Registry、HTTP Chart Repository、本地 Chart 包。
- Chart 元数据：名称、版本、应用版本、默认 values、维护人。
- Values 管理：按环境保存 values，支持差异对比、敏感字段脱敏。
- Release 管理：安装、升级、回滚、卸载、历史版本、部署状态。
- 模板渲染：预览 Kubernetes Manifest，支持发布前检查。
- 依赖治理：Chart 依赖、镜像依赖、环境变量和 Secret 依赖。

### 5. 镜像仓库与制品仓库对接

优先支持 Harbor，后续扩展其他 Registry。

- Registry 接入：Harbor、Docker Registry、云厂商镜像仓库。
- 项目与仓库同步：项目、镜像、Tag、Digest、构建时间、扫描状态。
- 镜像选择：部署时选择镜像版本，支持按分支、Commit、Tag 过滤。
- 安全扫描：漏洞等级、修复建议、阻断策略。
- 镜像晋级：从 dev 镜像晋级到 test、sim、prod。
- 清理策略：过期 Tag 清理、保留策略、不可变 Tag。

### 6. OAuth2 统一认证登录

提供统一身份认证和权限体系。

- 登录方式：OAuth2、OIDC、LDAP、企业微信/钉钉/GitLab/GitHub 等扩展。
- 用户同步：用户、组织、部门、团队、角色。
- 权限模型：RBAC 为基础，后续可扩展 ABAC。
- 资源授权：项目、环境、集群、命名空间、应用、流水线、文件。
- Token 管理：Access Token、Refresh Token、会话过期、单点登出。
- 审计日志：登录、授权变更、敏感操作、发布操作。

### 7. Tekton 流水线

以 Tekton 作为云原生流水线执行引擎。

- Pipeline 模板：构建、测试、扫描、镜像推送、部署、通知。
- Task 管理：通用任务库、业务任务、参数化任务。
- PipelineRun 管理：触发、取消、重试、日志、状态、耗时。
- 触发方式：手动触发、Git Webhook、定时触发、需求发布触发。
- 凭据管理：Git 凭据、Registry 凭据、Kube 凭据、Secret 引用。
- 结果回写：镜像 Tag、测试报告、扫描报告、发布记录。

### 8. 需求发布与环境上线

把需求、变更、发布单和环境状态串起来。

- 需求管理：需求编号、标题、负责人、关联服务、关联分支、关联文件。
- 发布单：发布范围、目标环境、发布窗口、审批人、风险等级。
- 环境流转：`dev -> test -> sim -> prod`，支持跳过、阻断、回退。
- 审批流程：测试确认、负责人审批、运维审批、生产发布确认。
- 变更记录：每次上线包含镜像、配置、Chart、文件、数据库脚本等变更内容。
- 发布看板：按环境、应用、需求、发布批次查看状态。

### 9. 文件版本管理

管理配置文件、脚本、部署文件、需求附件或其他版本化文件。

- 文件仓库：按项目、应用、环境、目录组织文件。
- 版本记录：版本号、提交人、提交说明、时间、Hash。
- 差异对比：文本 diff、配置 diff、二进制文件元数据对比。
- 发布绑定：文件版本可绑定到发布单、流水线或部署记录。
- 存储后端：本地存储、对象存储、Git 仓库、数据库元数据。
- 权限与审计：上传、下载、修改、删除、回滚全链路记录。

## 推荐模块拆分

初期可以放在一个 Go 项目中，但按以下领域模块组织。每个模块拥有自己的 service、repository、domain model 和
adapter，避免业务逻辑互相穿透。

```text
olixops
├── cmd
│   └── server              # API 服务入口
├── internal
│   ├── app                 # 应用装配、依赖注入、启动流程
│   ├── config              # 配置加载
│   ├── platform            # 基础设施能力
│   │   ├── auth            # OAuth2/OIDC、Token、会话
│   │   ├── rbac            # 权限、角色、资源授权
│   │   ├── audit           # 审计日志
│   │   ├── database        # 数据库连接、事务
│   │   ├── event           # 领域事件、消息总线
│   │   └── storage         # 文件/对象存储抽象
│   ├── modules
│   │   ├── project         # 项目、团队、业务线
│   │   ├── environment     # 环境、环境变量、环境策略
│   │   ├── application     # 应用与服务模型
│   │   ├── deployment      # 部署、发布、回滚
│   │   ├── agent           # 智能体应用、沙箱、会话
│   │   ├── cluster         # Kubernetes 集群管理
│   │   ├── helm            # Chart、Release、Values
│   │   ├── registry        # Harbor/Registry 对接
│   │   ├── pipeline        # Tekton 流水线
│   │   ├── release         # 需求发布、上线审批
│   │   └── fileversion     # 文件版本管理
│   └── interfaces
│       ├── http            # REST API、路由、中间件
│       ├── grpc            # 后续内部 RPC，可选
│       └── worker          # 异步任务、定时任务
├── pkg                     # 可复用 SDK/工具包，谨慎放置
├── api                     # OpenAPI/Proto 定义
├── deployments             # Helm Chart、K8s YAML、Dockerfile
├── migrations              # 数据库迁移
├── docs                    # 设计文档
└── tests                   # 集成测试、端到端测试
```

## 模块边界与后续拆分方向

建议先保持一个进程、一个数据库，但在代码中提前按“可拆服务”设计边界。

| 模块          | 初期形态                 | 后续可拆为                   |
|-------------|----------------------|-------------------------|
| 认证权限        | 内部 platform 模块       | `olixops-iam`           |
| 应用与部署       | 核心业务模块               | `olixops-deploy`        |
| 集群管理        | K8s adapter + 业务模块   | `olixops-cluster`       |
| Helm 管理     | 独立业务模块               | `olixops-helm`          |
| Registry 对接 | 插件适配模块               | `olixops-registry`      |
| Tekton 流水线  | Pipeline 模块 + Worker | `olixops-pipeline`      |
| 智能体沙箱       | Agent 模块 + 沙箱控制器     | `olixops-agent-runtime` |
| 文件版本        | 存储抽象 + 元数据模块         | `olixops-file`          |
| 发布流程        | 发布单、审批、环境流转          | `olixops-release`       |

拆分时优先选择以下信号：

- 某模块需要独立扩缩容，例如 PipelineRun、日志采集、沙箱执行。
- 某模块依赖重、故障域特殊，例如 Kubernetes Watch、Tekton Controller、Agent Sandbox。
- 某模块团队独立维护，接口稳定，例如 IAM、Registry、File Storage。
- 某模块需要独立安全边界，例如生产发布、凭据管理、沙箱执行。

## 领域模型草案

可以优先抽象以下核心实体：

- `User`：用户。
- `Team`：团队或组织。
- `Project`：业务项目，承载应用、环境和权限。
- `Environment`：环境，如 dev/test/sim/prod。
- `Application`：应用，可以是普通服务、智能体应用、前端应用、任务应用等。
- `ServiceComponent`：应用下的部署组件。
- `Cluster`：Kubernetes 集群。
- `NamespaceBinding`：项目/环境与命名空间绑定关系。
- `DeploymentPlan`：部署计划。
- `DeploymentRecord`：部署记录。
- `ReleaseOrder`：发布单。
- `Requirement`：需求或变更项。
- `PipelineTemplate`：流水线模板。
- `PipelineRun`：流水线执行记录。
- `ChartRepository`：Chart 仓库。
- `HelmRelease`：Helm 发布实例。
- `Registry`：镜像仓库。
- `ImageArtifact`：镜像制品。
- `FileAsset`：文件。
- `FileVersion`：文件版本。
- `AuditLog`：审计日志。
- `Credential`：外部系统凭据，注意加密存储。

## 推荐技术方向

后端技术可以保持简洁，先把领域模型和集成边界跑通。

- 语言：Go。
- API：REST + OpenAPI，内部复杂任务可逐步引入 gRPC。
- Web 框架：Gin、Echo、Fiber 或标准库路由均可，建议优先选团队熟悉的。
- 数据库：PostgreSQL，适合关系模型、JSONB 和审计查询。
- ORM/SQL：GORM、Ent、SQLC 三选一。若追求强类型和可控 SQL，可选 SQLC。
- 缓存：Redis，用于会话、锁、任务状态、短期缓存。
- 异步任务：初期可用数据库任务表，后续接 NATS、RabbitMQ、Kafka。
- K8s 集成：`client-go`、dynamic client、informers。
- Helm 集成：Helm SDK。
- Tekton 集成：Tekton CRD client。
- 认证：OAuth2/OIDC，推荐兼容 Keycloak、Dex、企业 IdP。
- 文件存储：本地开发 + S3/MinIO 抽象。
- 配置：环境变量 + YAML，后续支持配置中心。

## API 设计建议

接口按资源组织，并保持环境、项目、应用几个主维度一致。

```text
/api/v1/auth/*
/api/v1/users
/api/v1/teams
/api/v1/projects
/api/v1/projects/{projectId}/environments
/api/v1/projects/{projectId}/applications
/api/v1/applications/{appId}/deployments
/api/v1/applications/{appId}/releases
/api/v1/clusters
/api/v1/clusters/{clusterId}/namespaces
/api/v1/helm/repositories
/api/v1/helm/releases
/api/v1/registries
/api/v1/pipelines
/api/v1/pipeline-runs
/api/v1/release-orders
/api/v1/files
/api/v1/audit-logs
```

## 最小可行版本

建议第一阶段不要一口气做完所有集成，而是先做出可闭环的主线。

### Phase 1：基础平台

- 用户、团队、项目、环境。
- OAuth2/OIDC 登录。
- RBAC 权限。
- 审计日志。
- 应用注册。
- Kubernetes 集群接入。

### Phase 2：部署闭环

- 应用部署模型。
- 命名空间绑定。
- 镜像仓库接入。
- Helm Chart 接入。
- 部署记录、回滚记录。
- 基础日志和状态查看。

### Phase 3：流水线与发布

- Tekton Pipeline 模板。
- PipelineRun 触发和日志。
- 需求、发布单、审批流。
- 环境流转：dev/test/sim/prod。
- 镜像晋级和配置差异对比。

### Phase 4：智能体与文件版本

- Agent 应用模型。
- Agent 沙箱生命周期。
- 工具调用审计。
- 文件版本管理。
- 文件版本与发布单绑定。

### Phase 5：治理与平台化

- 多集群调度。
- 发布策略增强：蓝绿、金丝雀。
- 成本统计与资源配额。
- 安全扫描、策略阻断。
- 插件化外部系统适配。
- 关键模块拆分为独立服务。

## 开发约定建议

- 业务代码只依赖领域接口，不直接依赖外部 SDK。
- Kubernetes、Helm、Harbor、Tekton、OAuth2 等外部系统统一放在 adapter 层。
- 所有生产敏感操作必须写审计日志。
- 所有跨环境发布必须有明确的部署记录和可回滚依据。
- 凭据、Token、Secret 不直接明文入库。
- 文件版本、Chart Values、发布参数都需要支持 diff。
- 模块间通信初期用内部 service 调用，后续拆分时替换成 RPC 或事件。

## 当前状态

当前项目仍处于初始化阶段，下一步建议先完成：

1. 确定基础 Web 框架和数据库访问方式。
2. 建立标准目录结构。
3. 实现配置加载、日志、中间件、健康检查。
4. 建立用户、项目、环境、应用四个基础模型。
5. 设计 OpenAPI 文档和数据库迁移规范。

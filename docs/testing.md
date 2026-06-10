# 测试文档

本文档说明 SEU OJ 后端测试的组织方式、运行命令和覆盖范围。用户提到的“继承测试”这里按软件工程常用的“集成测试”落地；仓库中目前也没有需要专门做继承层级验证的业务模型。

## 测试类型

### 单元测试

单元测试覆盖不依赖外部服务的纯逻辑和边界处理：

- 配置加载与环境变量覆盖：`internal/config`
- JWT 生成、解析、过期和签名错误：`internal/utils`
- Gin 中间件：JWT、可选 JWT、CORS、限流、角色鉴权、请求耗时、审计辅助函数
- 响应 envelope：`internal/response`
- 观测工具：`internal/observability`
- 沙箱配置、语言规格、Docker 参数拼装、输出截断缓冲区
- 业务纯函数：题目包解析、输出对比、竞赛状态、教学进度等

### 集成测试

集成测试使用 GORM SQLite 内存库，不依赖本机 MySQL、Redis 或 Docker：

- Repository 层：题目、测试点、提交、提交结果 CRUD 和过滤排序
- Service 层：认证、题目生命周期、题解、题目包导入导出、公告、竞赛、论坛、榜单、统计、用户管理
- Router 层：`/api/auth/*` 登录注册流程、管理员题目 CRUD、公开题目详情/统计、健康检查、审计日志写入

SQLite 只用于测试本地业务逻辑和 GORM 调用路径。生产数据库仍是 MySQL。

## 运行命令

在后端目录执行：

```bash
cd seu-oj-backend
go test ./...
```

### Docker 可选集成测试（自动检测）

部分测试会**自动检测 Docker Desktop 是否可用**：

- **Docker 未启动**：相关用例 `Skip`，其余测试照常跑完（约 **65%** 总覆盖率）
- **Docker 已启动**：额外执行沙箱编译运行、判题 Worker、样例运行等用例（总覆盖率约 **72%+**，`judge`/`sandbox` 可达 **85%+/85%+**）

无需额外参数，直接：

```powershell
go test ./... -count=1 "-coverprofile=coverage.out"
go tool cover -func coverage.out
```

强制跳过 Docker 测试（即使 Docker 已开）：

```powershell
$env:SEUOJ_SKIP_DOCKER = "1"
go test ./...
```

首次跑 Docker 测试会自动 `docker pull python:3` / `gcc:13`，可能需要几分钟。

涉及包：

- `internal/sandbox` — Compile / Run / Validate / TLE / RE
- `internal/judge` — Worker 完整判题链路（AC / WA / CE）
- `internal/service` — `RunSampleTests` 在线运行样例

数据库连通性/初始化（读 `config/config.yaml`）：

```bash
go run ./cmd/db-init --ping
go run ./cmd/db-init --check
```

生成覆盖率：

```powershell
cd seu-oj-backend
go test ./... "-coverprofile=coverage.out"
go tool cover -func coverage.out
```

> Windows PowerShell 下 `-coverprofile=coverage.out` 必须加引号，否则 `.out` 会被解析成属性访问。

生成 HTML 覆盖率报告：

```powershell
go tool cover -html coverage.out -o coverage.html
```

`coverage.out` 和 `coverage.html` 是本地生成文件，已加入 `.gitignore`。

详细对比与答辩口径见 [test-coverage-report.md](./test-coverage-report.md)。

## 当前覆盖率基线

2026-06-09 本地验证结果（`go test ./... -count=1`，Docker 已开）：

| 场景 | 总覆盖率 | 说明 |
|------|---------|------|
| Docker **未开** | **~65%** | Docker 用例 Skip，api/service 集成测仍跑 |
| Docker **已开** | **~72%** | 含沙箱 + Worker + 样例运行 + api Handler 直调 |

重点包（Docker 已开，第六轮加深后）：

- `api` **~63%**（原 ~1.4%）
- `service` **~76%**、`repository` **92.3%**
- `sandbox` **85.6%**、`judge` **~85%**、`queue` **91.7%**
- `router` **88.8%**、`middleware` **86.9%**、`cache` **85.7%**
- 高覆盖核心包：`config 96.2%`、`response 100%`、`utils 90.9%`

跨包统计（更贴近真实 HTTP 执行路径）：

```powershell
go test ./... -count=1 "-coverpkg=./..." "-coverprofile=coverage-all.out"
go tool cover -func coverage-all.out
```

详细对比与答辩口径见 [test-coverage-report.md](./test-coverage-report.md)。

## 后续扩展建议

新增业务时优先按以下顺序补测试：

1. 先测纯函数和参数校验。
2. Repository/Service 使用 SQLite 内存库覆盖正常路径、权限错误、not found、重复数据。
3. API/Router 用 `httptest` 走真实 Gin 路由，验证鉴权、响应 envelope 和关键副作用。
4. Docker 判题、Redis 队列、MySQL 方言行为单独放到需要外部服务的端到端测试，不混入默认 `go test ./...`。

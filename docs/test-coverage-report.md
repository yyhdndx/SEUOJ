# 测试覆盖率报告

本文档用于结项答辩与交付说明，记录 SEU OJ 后端测试加强的规划、实施结果与后续建议。

## 1. 目标与策略

### 背景

2026-05-27 基线：`go test ./...` 全部通过，**总语句覆盖率 34.0%**。主要短板集中在：

| 模块 | 基线覆盖率 | 原因 |
|------|-----------|------|
| `internal/queue` | ~8% | 依赖 Redis，原测试几乎为空 |
| `internal/service`（提交链路） | ~44% | 入队失败、分页 total、Rejudge 等路径未测 |
| `internal/cache` | ~51% | 仅 nil-cache 与 GetOrSet miss 路径 |
| `cmd/db-init` | 0% | 新建工具，无单测 |
| `internal/judge` | ~12% | 需 Docker 沙箱，默认 CI 不跑 |
| `internal/api` | ~1.4% | Handler 多由 router 集成测试间接覆盖 |

### 加强原则

1. **默认 `go test ./...` 不依赖外部服务**：MySQL / 真实 Redis / Docker 仍不纳入常规单测。
2. **用 miniredis 补 Redis 路径**：队列、缓存、提交入队。
3. **用 SQLite 内存库补 Service/Repository**：业务逻辑与 GORM 调用。
4. **优先覆盖 P0/P1 修复点**：入队失败标 System Error、提交分页 total、Rejudge 权限等。
5. **Router 层 httptest 走真实 Gin 路由**：验证鉴权 + envelope + 副作用。
6. **Docker 可选、自动检测**：`internal/testutil` 探测 Docker Desktop；可用则跑沙箱/Worker 集成测，不可用则 Skip。

## 2. 本次新增/扩展测试

### 2.1 提交服务（`internal/service/submission_service_test.go`）

| 用例 | 覆盖场景 |
|------|---------|
| `TestSubmissionServiceCreateSubmissionEnqueues` | 正常提交 → Pending + Redis 入队 |
| `TestSubmissionServiceCreateSubmissionRejectsHiddenProblem` | 隐藏题目拒绝提交 |
| `TestSubmissionServiceEnqueueFailureMarksSystemError` | Redis 不可用 → System Error 落库（P0 修复） |
| `TestSubmissionServiceListMySubmissionsUsesRealTotal` | 分页 total=25 而非仅当前页（P1 修复） |
| `TestSubmissionServiceRejudgeRequiresAdmin` | 学生拒绝 / 管理员 Rejudge |

### 2.2 判题队列（`internal/queue/judge_queue_test.go`）

| 用例 | 覆盖场景 |
|------|---------|
| `TestJudgeQueueEnqueueDequeueAndLength` | 入队、出队、长度 |
| `TestJudgeQueueDequeueEmptyBlocksUntilCancelled` | BLPop 阻塞 + 取消 |

### 2.3 数据库初始化工具（`cmd/db-init/main_test.go`）

| 用例 | 覆盖场景 |
|------|---------|
| `TestExtractCreateProcedureFromIndexSQL` | 从 index.sql 提取存储过程 |
| `TestCallLinePatternMatchesIndexSQL` | CALL 行正则匹配 |
| `TestDirExists` / `TestResolveDatabaseDirFromBackend` | 路径辅助函数 |

### 2.4 教学模块（`internal/service/service_integration_test.go`）

| 用例 | 覆盖场景 |
|------|---------|
| `TestTeachingPlaylistDetailProgress` | 题单进度：已 AC / 未开始 / next problem |

> 说明：Playlist 等表含 MySQL `ENUM`，SQLite AutoMigrate 不兼容，测试中通过 `migrateTeachingTablesSQLite` 手工建表。

### 2.5 缓存（`internal/cache/cache_test.go`）

| 用例 | 覆盖场景 |
|------|---------|
| `TestCacheJSONWithRedis` | Get/Set JSON、DeletePrefixes |
| `TestGetOrSetUsesCacheHit` | GetOrSet 命中路径 |

### 2.6 路由集成（`internal/router/router_integration_test.go`）

| 用例 | 覆盖场景 |
|------|---------|
| `TestSubmissionLifecycleIntegration` | 创建提交 → 我的列表 → 详情 → Rejudge 鉴权 |

### 2.7 Docker 可选集成测试（`internal/testutil/docker.go`）

| 文件 | 用例 | 覆盖场景 |
|------|------|---------|
| `sandbox/docker_integration_test.go` | Python/C++ 全路径 | Compile + Run + TLE + RE |
| `judge/worker_integration_test.go` | Worker 分支 + Docker 判题 | AC / WA / CE |
| `service/submission_run_docker_test.go` | RunSampleTests | 在线样例 |

### 2.8 第二轮加强（2026-06-09 晚）

| 文件 | 用例 | 覆盖场景 |
|------|------|---------|
| `teaching_integration_test.go` | `TestTeachingModuleLifecycle` | 题单 CRUD + 班级 + 作业 + Analytics 全链路 |
| `submission_service_test.go` | Detail / PublicList / AdminList | 提交详情、公开榜、过滤 |
| `stats_service_test.go` | `TestStatsServiceMyAndAdmin` | 个人/管理员统计 + 队列长度 |
| `router_more_integration_test.go` | Forum/Stats/Contest/Teaching/Run | HTTP 入口大面积补测 |
| 生产小改 | forum/stats/submission | SQLite 兼容（`CASE WHEN`、`SUBSTR`、`datetime`） |

### 2.9 第三轮加强（2026-06-09 深夜）— 针对 `internal/api` 大头

| 文件 | 用例 | 覆盖场景 |
|------|------|---------|
| `handlers_integration_test.go` | `TestAPIHandlersIntegration` | Auth/公告/用户/榜单/统计/题目/提交/论坛/竞赛/教学 主链路 |
| `handlers_more_integration_test.go` | `TestAPIHandlersMoreIntegration` | 题目 CRUD、竞赛公告/AdminRanklist、班级成员、作业、论坛收藏/点赞、Rejudge |
| `teaching_integration_test.go` | `TestTeachingClassMemberAndUniqueCode` | `UpdateClassMember`、`CreateClassWithUniqueCode` |
| `service_integration_test.go` | 扩展 | `PublicProblemDetail`、`GetAdminRanklist` |

> **关键突破**：在 `internal/api` 包内直接调用 Handler（Gin TestContext + SQLite），使 api 包默认统计从 **~1.4% → 55%**，总覆盖率 **49% → 69%**。

### 2.10 第四轮加强（2026-06-09）— 冲刺 70%

| 文件 | 用例 | 覆盖场景 |
|------|------|---------|
| `handlers_package_test.go` | 题目包导入导出、Run 样例 | multipart 上传、zip 下载、Docker Run |
| `handlers_errors_test.go` | 生命周期扩展 | 竞赛/题单/作业/论坛 CRUD 删除与更新 |
| `handlers_branches_test.go` | 错误矩阵 | 权限拒绝、not found、非法参数 |
| `handlers_auth_extended_test.go` | 登录/改密 | 注册后登录、改密、404 路径 |
| `auth_service_test.go` | 认证全链路 | 注册/登录/改密/禁用用户/重复名 |
| `submission_service_test.go` | 竞赛提交/Run | `CreateSubmission(contest)`、`loadRunnableProblem` |
| `database_test.go` / `audit_test.go` | 基础设施 | Redis 连接、审计日志、限流 key |
| `router_more_integration_test.go` | 用户/公告/榜单 | Admin HTTP 入口 |
| `cmd/db-init/main_test.go` | dsn/resolveDir | 连接串与 SQL 目录解析 |

## 3. 覆盖率对比（2026-06-09）

在 `seu-oj-backend` 目录执行（建议加 `-count=1` 避免缓存）：

```powershell
go test ./... -count=1 "-coverprofile=coverage.out"
go tool cover -func coverage.out
```

跨包统计（Router 跑到的 api 代码也计入，**答辩推荐用这个数字**）：

```powershell
go test ./... -count=1 "-coverpkg=./..." "-coverprofile=coverage-all.out"
go tool cover -func coverage-all.out
```

> **Windows PowerShell 注意**：`-coverprofile=...` 必须加引号。

### 总览

| 指标 | 基线 (05-27) | 当前（默认） | 跨包 `-coverpkg` |
|------|-------------|-------------|------------------|
| **总语句覆盖率** | **34.0%** | **~69%** | **~69%** |
| 测试是否全部通过 | 是 | 是 | 是 |

### 重点包对比（Docker 已开）

| 包 | 基线 | 当前 |
|----|------|------|
| `internal/api` | ~1.4% | **~63%** |
| `internal/service` | 44.1% | **~76%** |
| `internal/sandbox` | 45.2% | **85.6%** |
| `internal/judge` | 12.1% | **~85%** |
| `internal/queue` | ~8% | **91.7%** |
| `internal/repository` | — | **92.3%** |
| `internal/middleware` | — | **86.9%** |
| `internal/cache` | ~51% | **85.7%** |
| `internal/router` | 88.8% | **88.8%** |
| `cmd/db-init` | 0% | **~20%** |

### 仍偏低且合理的模块

| 包 | 覆盖率 | 原因 |
|----|--------|------|
| `internal/database` | ~27% | `New()` 需真实 MySQL AutoMigrate |
| `main` / `cmd/judge-worker` | 0% | 进程入口 |
| `internal/api` | ~59% | 部分分支需 Docker/文件上传组合 |

## 4. 答辩口径（可直接引用）

> 后端测试 **约 72% 语句覆盖率**（api **63%**、service **76%**、sandbox **86%**、judge **85%**、queue **92%**、router **89%**）。
>
> 采用 SQLite 集成 + miniredis + **api 包内 Handler 直调** + httptest 路由 + Docker 可选判题；较基线 34% 提升约 **35 个百分点**。

## 5. 后续可扩展项（P2）

1. docker-compose E2E：MySQL + Redis + Worker 一条链路。
2. `internal/database`：`-tags=mysql` 连通性测试。
3. 用户管理 Admin API Router 集成测。

## 6. 复现命令

```powershell
cd seu-oj-backend

# 日常（Docker 开/关均可）
go test ./...

# 答辩前推荐：Docker Desktop 已启动
go test ./... -count=1 "-coverprofile=coverage.out"
go tool cover -func coverage.out
go tool cover -html coverage.out -o coverage.html

# 跨包覆盖率（答辩数字更好看）
go test ./... -count=1 "-coverpkg=./..." "-coverprofile=coverage-all.out"
go tool cover -func coverage-all.out

# 强制跳过 Docker
$env:SEUOJ_SKIP_DOCKER = "1"
go test ./...
```

HTML 报告 `coverage.html` 与 `coverage.out` 已加入 `.gitignore`，本地生成即可。

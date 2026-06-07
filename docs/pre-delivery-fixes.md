# 结项前修复记录

本文档记录 SEU OJ 交付前针对审计报告（P0/P1）所做的修复与文档补齐。

修复日期：2026-06-06

---

## 一、急需修复（P0）

### 1. 判题入队失败导致永久 Pending

**问题**：`CreateSubmission` / `RejudgeSubmission` 先写库再入 Redis 队列；Redis 不可用时留下 `Pending` 记录且不会被 Worker 处理。

**修复**（`internal/service/submission_service.go`）：
- 入队失败时调用 `markEnqueueFailure`，将状态改为 `System Error`，错误信息为 `judge queue unavailable`
- 若标记失败则删除刚创建的记录（兜底）
- 重判入队失败同样标记为 `System Error`，避免假 Pending

**答辩注意**：Submit 失败时检查 Redis 与 Judge Worker 是否已启动。

### 2. 限流 Retry-After 计算错误 + 测试失败

**问题**：`ratelimit.go` 在被限流时使用 `time.Until(bucket.ResetAt)`（墙钟），单元测试使用固定历史时间导致 `Retry-After` 为负值，CI 失败。

**修复**：改为 `bucket.ResetAt.Sub(now)`，负值钳制为 0。

### 3. 数据库初始化不完整

**问题**：`AutoMigrate` 原先不包含全部核心表；无统一 init 流程；Windows `mysql.exe` 常因插件问题无法导入 SQL。

**修复**：
- `internal/database/db.go`：AutoMigrate 增加核心表
- 新增 **`cmd/db-init`**：Go 工具读 `config.yaml` 导入 SQL，**无需 mysql CLI**
- 新增 `scripts/dev.sh`、`scripts/dev.ps1` 一键启动
- 新增 [database-init.md](./database-init.md)、[scripts/README.md](../scripts/README.md)

### 4. 配置示例缺失

**问题**：`.gitignore` 预留了 `config.example.yaml` 但文件不存在；JWT 默认为开发密钥。

**修复**：新增 `seu-oj-backend/config/config.example.yaml`，答辩前复制为 `config/config.yaml` 并修改 `jwt_secret`。

### 5. CodeMirror 依赖未安装

**问题**：前端 importmap 依赖 `CodeMirror/node_modules`，仓库未包含。

**修复**：
- [demo-checklist.md](./demo-checklist.md) 增加 npm install、一键脚本说明
- 新增 [scripts/dev.sh](../scripts/dev.sh)（Git Bash 验证通过）

**答辩前执行**：

```bash
bash scripts/dev.sh
```

---

## 二、较为需要改进（P1）

### 1. 前端调试文案泄露

**问题**：题目页样例区显示 `sort_order` / `active`；比赛题目页显示 `X sample case(s) are available.`。

**修复**：
- `seu-oj-frontend/js/problems.js`：移除样例副标题中的内部字段
- `seu-oj-frontend/js/contests.js`：移除样例数量调试文案

### 2. 提交列表分页 total 不准确

**问题**：`listMySubmissions` 在第一页且 `pageSize<=20` 时用 `ListRecentByUserID`，`total` 取当前页条数而非真实总数。

**修复**：统一走 `ListByUserID`（含 COUNT）。

### 3. 提交代码无长度上限

**问题**：DTO 仅 `required`，恶意超大代码可占满资源。

**修复**：`CreateSubmissionRequest` / `RunSubmissionRequest` 的 `code` 字段增加 `max=65535`。

### 4. 演示 admin 账号缺失

**问题**：seed 仅有 teacher/student，管理页演示需手动注册。

**修复**：`seed_classes_assignments.sql` 增加 `demo_admin` / `A-DEMO-2026`（密码 `123456`）。

### 5. 部署与答辩文档缺口

**修复**：
- 更新 [README.md](../README.md)、[demo-checklist.md](./demo-checklist.md)
- 新增 [scripts/README.md](../scripts/README.md)、[database-init.md](./database-init.md)
- 数据库/启动脚本统一读 `config/config.yaml`

---

## 二点五、一键启动脚本（2026-06 补充）

**背景**：Windows 上 PowerShell 执行策略限制 + `mysql.exe` 插件报错，手动步骤过多。

**方案**：
- `bash scripts/dev.sh`：Git Bash 下一键检查、db-init、npm、后台启动 Web/Worker
- `go run ./cmd/db-init`：替代 mysql CLI 做 schema/seed 导入
- 配置统一来自 `config/config.yaml`

**已验证流程**（Git Bash / Windows）：

```bash
bash scripts/dev.sh
# [OK] MySQL reachable
# [SKIP] Database already initialized
# [OK] Web server started → logs/web.log
# [OK] Judge worker started → logs/worker.log

bash scripts/stop-dev.sh
```

### 6. 论坛教师权限（设计保留，文档说明）

**现状**：`teacher` 可编辑/删除任意帖子（含置顶/锁帖），集成测试 `TestForumService` 明确覆盖此行为。

**结论**：作为教学 OJ 的 moderation 能力保留，不在此次修改行为；答辩口径见 [defense-qa.md](./defense-qa.md)。

### 7. Worker 崩溃后队列消息丢失（已知限制）

**现状**：Redis `BLPop` 后若 Worker 进程崩溃，提交可能永久 Pending。

**缓解**：
- 答辩前确认 Worker 稳定运行
- 管理员后台可对卡住提交执行「重判」
- 后续可选：定时扫描 Pending 超时记录并重新入队

---

## 三、验证清单

```bash
# 推荐：一键启动（Git Bash）
bash scripts/dev.sh
tail -f logs/web.log logs/worker.log
bash scripts/stop-dev.sh

# 仅检查/初始化
bash scripts/dev.sh --setup-only

# 数据库连通性
cd seu-oj-backend && go run ./cmd/db-init --ping

# 后端测试
go test ./...
```

浏览器访问 `http://127.0.0.1:8080/`，按 [demo-checklist.md](./demo-checklist.md) 回归。

---

## 四、未纳入本次修复的 backlog（P2）

- docker-compose 一键部署
- Judge Worker 多 goroutine 并行判题
- 服务端代码草稿、自定义测试输入（STORY-003）
- 首页 portal 化（STORY-004）
- E2E 自动化回归
- 限流改为 Redis 分布式实现

详见审计报告与 `stories/open/` 目录。

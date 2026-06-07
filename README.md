# SEU OJ

SEU OJ 是一个面向课程教学、日常训练与竞赛组织的轻量级 Online Judge 系统。

当前仓库包含：
- `seu-oj-backend`：Go 后端，提供 API、判题调度、静态资源托管
- `seu-oj-frontend`：静态前端 SPA，按业务域拆分 JS/CSS
- `docs`：接口文档、需求建模、答辩材料与协作文档

## 当前能力概览

系统当前已经具备这些核心模块：
- 用户注册、登录、JWT 鉴权、角色管理
- 题库、题解、题目统计、管理员题目 CRUD
- 多语言提交与 Docker 沙箱判题
- 比赛、报名、榜单、封榜、赛后练习、比赛公告
- 题单、班级、作业/考试、教师管理页
- 论坛、公告、排行榜、提交管理

详细功能说明见：
- [当前功能总览](./docs/current-features.md)
- [后端接口文档](./docs/api.md)
- [需求建模与用例图](./docs/requirements-modeling.md)
- [PPT 需求部分建议](./docs/ppt-requirements.md)

## 运行方式

推荐先看：
- [展示前检查清单](./docs/demo-checklist.md)
- [数据库初始化指南](./docs/database-init.md)
- [结项前修复记录](./docs/pre-delivery-fixes.md)
- [后端接口文档](./docs/api.md)

### 一键启动（推荐）

**第一步**：编辑 `seu-oj-backend/config/config.yaml`（数据库、Redis 等）。可从 `config.example.yaml` 复制。

**第二步**：确保 MySQL、Redis、Docker 可用，然后在仓库根目录：

```bash
# Git Bash / WSL / Linux（推荐，已在 Windows Git Bash 验证）
bash scripts/dev.sh
```

常用命令：

```bash
bash scripts/dev.sh --setup-only   # 仅检查/初始化，不启动
bash scripts/stop-dev.sh           # 停止服务
tail -f logs/web.log logs/worker.log
```

浏览器访问：http://127.0.0.1:8080/

**Windows PowerShell**（若 Git Bash 不可用）：

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\dev.ps1
# 或双击 scripts\start-dev.bat
```

脚本会自动：检查环境 → 读 `config.yaml` → 用 `go run ./cmd/db-init` 初始化数据库（无需 mysql CLI）→ 安装 CodeMirror → 启动 Web + Worker。

详见 [scripts/README.md](./scripts/README.md)。

### 手动分步（不用一键脚本时）

1. 配置：复制并编辑 `seu-oj-backend/config/config.yaml`
2. 数据库：`cd seu-oj-backend && go run ./cmd/db-init`（详见 [database-init.md](./docs/database-init.md)）
3. CodeMirror：`cd seu-oj-frontend/CodeMirror && npm install`
4. 启动（需两个终端）：
   ```bash
   cd seu-oj-backend
   go run .
   go run ./cmd/judge-worker
   ```

### 日常启动

优先使用 `bash scripts/dev.sh` 或 `dev.ps1` / `start-dev.bat`。

手动启动时需**同时**运行 Web 与 Judge Worker，否则 Submit 会卡在 Pending。

## 文档索引

- [结项前修复记录](./docs/pre-delivery-fixes.md)：P0/P1 问题修复说明
- [数据库初始化指南](./docs/database-init.md)：MySQL schema 与 seed 执行顺序
- [开发脚本说明](./scripts/README.md)：一键启动 `dev.sh` / `dev.ps1`
- [STORY-001 交付说明（比赛页 / 榜单 UX）](./stories/deliveries/STORY-001-contest-ranklist-ux.md)
- [STORY-005 交付说明（题单训练进度 / Playlists）](./stories/deliveries/STORY-005-playlist-training-progress.md)
- [docs/api.md](./docs/api.md)：后端接口文档
- [docs/current-features.md](./docs/current-features.md)：当前实现功能总览
- [docs/demo-checklist.md](./docs/demo-checklist.md)：展示前功能检查清单
- [docs/requirements-modeling.md](./docs/requirements-modeling.md)：需求建模说明
- [docs/usecase-overall.puml](./docs/usecase-overall.puml)：总体用例图
- [docs/usecase-teaching-split.puml](./docs/usecase-teaching-split.puml)：教学模块用例图

## 协作建议

前端已经按业务域拆分：
- `js/problems.js`
- `js/submissions.js`
- `js/contests.js`
- `js/teaching.js`
- `js/admin.js`
- `js/forum.js`

样式也已拆分：
- `css/base.css`
- `css/problem.css`
- `css/submission.css`
- `css/contest.css`
- `css/teaching.css`
- `css/forum.css`

建议组员按业务域并行维护，而不是多人同时修改同一个入口文件。

# 展示前检查清单

本文档用于答辩或课堂展示前的快速回归。

相关文档：
- [database-init.md](./database-init.md) — 数据库初始化
- [scripts/README.md](../scripts/README.md) — 一键启动脚本
- [pre-delivery-fixes.md](./pre-delivery-fixes.md) — 结项修复说明

---

## 0. 最快启动（推荐）

在仓库根目录，**先确保 MySQL / Redis / Docker 已可用**，然后：

```bash
# Git Bash（Windows 推荐）
bash scripts/dev.sh
```

脚本会自动：读 `config/config.yaml` → 检测/初始化数据库 → 安装 CodeMirror（若缺失）→ 后台启动 Web + Worker。

- 浏览器访问：http://127.0.0.1:8080/
- 查看日志：`tail -f logs/web.log logs/worker.log`
- 停止服务：`bash scripts/stop-dev.sh`

PowerShell 用户见 [scripts/README.md](../scripts/README.md)（需 `ExecutionPolicy Bypass` 或 `start-dev.bat`）。

---

## 1. 基础环境

- [ ] 已在 `seu-oj-backend/config/config.yaml` 配置好 `database`（与后端共用）
- [ ] MySQL 可连接（`cd seu-oj-backend && go run ./cmd/db-init --ping`）
- [ ] 数据库已初始化（`go run ./cmd/db-init --check`，或由 `dev.sh` 自动完成）
- [ ] Redis 已启动（默认 `127.0.0.1:6379`，见 config.yaml 的 `redis.addr`）
- [ ] Docker 已启动，`gcc:13` 镜像可用
- [ ] Web + Judge Worker 均已运行（`dev.sh` 或手动两个终端）

手动等价步骤（仅当不用一键脚本时）：

```bash
# 1. 配置
cp seu-oj-backend/config/config.example.yaml seu-oj-backend/config/config.yaml

# 2. 数据库（读 config.yaml，无需 mysql 客户端）
cd seu-oj-backend && go run ./cmd/db-init

# 3. CodeMirror
cd ../seu-oj-frontend/CodeMirror && npm install

# 4. 启动（两个终端）
cd ../../seu-oj-backend
go run .
go run ./cmd/judge-worker
```

---

## 2. 演示账号（seed 导入后，密码均为 `123456`）

| 角色 | 用户名 | 学工号 |
|------|--------|--------|
| 管理员 | demo_admin | A-DEMO-2026 |
| 教师 | demo_teacher | T-DEMO-2026 |
| 学生 | demo_alice | S-DEMO-001 |

---

## 3. 首页与登录

- [ ] 首页可正常渲染
- [ ] 注册、登录、退出登录正常
- [ ] 导航栏跳转正常
- [ ] Console 无 CodeMirror 模块 404

---

## 4. 题库主链路

- [ ] 题目列表可显示多道题
- [ ] 题目详情可正常渲染 Markdown，CodeMirror 有语法高亮
- [ ] 样例区不显示 `sort_order` / `active` 等内部字段
- [ ] `Run` 正常返回样例结果
- [ ] `Submit` 能从 `Pending` 进入最终状态（**Worker 必须运行**）
- [ ] 提交详情页可轮询

---

## 5. 比赛模块

- [ ] 比赛列表可打开
- [ ] 比赛详情可展示题目、榜单、公告
- [ ] 可报名比赛
- [ ] 比赛题目页可提交
- [ ] 榜单可正常显示

---

## 6. 教学模块

- [ ] 题单列表与详情正常（含训练进度）
- [ ] 班级列表与详情正常
- [ ] 作业详情正常
- [ ] 教师班级页分析卡正常
- [ ] 教师作业页筛选正常

---

## 7. 论坛与公告

- [ ] 论坛列表正常
- [ ] 帖子详情正常
- [ ] 可发帖、回帖
- [ ] 公告页可打开

---

## 8. 管理与教师页面

- [ ] 管理员题目页可打开（`demo_admin` 登录）
- [ ] 管理员比赛页可打开
- [ ] 教师题单、班级、作业页可打开

---

## 9. 浏览器检查

- [ ] 无明显空白页
- [ ] 无严重 Console 红字
- [ ] 静态资源 `/js/*` `/css/*` 正常加载

---

## 10. 故障排查速查

| 现象 | 可能原因 | 处理 |
|------|----------|------|
| Submit 一直 Pending | Worker 或 Redis 未启动 | 确认 `logs/worker.log`、启动 Redis |
| Submit 报 enqueue failed | Redis 连接失败 | 检查 config.yaml 的 redis；记录会标 System Error |
| 题目页无代码高亮 | CodeMirror 未安装 | `dev.sh` 会自动 npm install |
| 登录/注册失败 | DB 未初始化或 config 错误 | `go run ./cmd/db-init --ping` |
| DB 初始化报 mysql 插件错误 | Windows mysql.exe 损坏 | **改用** `go run ./cmd/db-init`，勿用 mysql CLI |
| Run/Submit 编译失败 | Docker 未运行 | 启动 Docker，确认 `gcc:13` 镜像 |
| PowerShell 禁止运行脚本 | 执行策略限制 | 用 `bash scripts/dev.sh` 或 `start-dev.bat` |

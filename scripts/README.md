# 开发脚本说明

本目录提供 SEU OJ 的一键检查、初始化与启动脚本。

**核心约定**：MySQL 连接信息统一读 `seu-oj-backend/config/config.yaml`；数据库导入使用 Go 工具 `cmd/db-init`，**不依赖本机 mysql 命令**。

---

## 快速开始（已验证：Git Bash on Windows）

```bash
# 在仓库根目录
bash scripts/dev.sh

# 浏览器打开
# http://127.0.0.1:8080/

# 看日志
tail -f logs/web.log logs/worker.log

# 停止
bash scripts/stop-dev.sh
```

---

## dev.sh — Git Bash / WSL / Linux / macOS（推荐）

```bash
bash scripts/dev.sh                 # 检查 + 初始化 + 后台启动
bash scripts/dev.sh --setup-only    # 只做初始化，不启动服务
bash scripts/dev.sh --force-db-init # 强制重新导入数据库
bash scripts/dev.sh --skip-docker-pull
bash scripts/stop-dev.sh            # 停止 Web + Worker
bash start-dev.sh                   # 根目录快捷入口
```

### 启动方式

`dev.sh` 在后台启动两个进程，日志写入：

| 文件 | 内容 |
|------|------|
| `logs/web.log` | Web 服务（`:8080`） |
| `logs/worker.log` | Judge Worker |
| `logs/web.pid` / `logs/worker.pid` | 进程 PID |

### 典型输出（成功）

```
[OK] MySQL reachable (host:3306/dbname)
[SKIP] Database 'xxx' already initialized
[SKIP] CodeMirror node_modules ready
[OK] Web server started (PID ..., log: logs/web.log)
[OK] Judge worker started (PID ..., log: logs/worker.log)
```

---

## dev.ps1 — Windows PowerShell

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\dev.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\dev.ps1 -SetupOnly
powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\stop-dev.ps1
```

或双击 `scripts/start-dev.bat`。

与 `dev.sh` 不同：`dev.ps1` 会**新开两个 PowerShell 窗口**分别运行 Web 与 Worker，便于直接看控制台输出。

> 若直接 `.\scripts\dev.ps1` 报「禁止运行脚本」，请用 `Bypass` 或改用 `bash scripts/dev.sh`。

---

## 脚本会自动做什么

| 步骤 | 条件 | 动作 |
|------|------|------|
| 读配置 | 始终 | 从 `config/config.yaml` 读取 database / redis |
| 检查 MySQL | 始终 | `go run ./cmd/db-init --ping` |
| 配置文件 | 无 config.yaml | 从 config.example.yaml 复制 |
| 数据库 | users 表无数据 | `go run ./cmd/db-init` |
| CodeMirror | 无 node_modules | `npm install` |
| Docker 镜像 | 无 gcc:13 | `docker pull gcc:13` |
| 启动服务 | 非 setup-only | 后台（sh）或新窗口（ps1） |

---

## 数据库工具 cmd/db-init

```bash
cd seu-oj-backend
go run ./cmd/db-init --ping
go run ./cmd/db-init --check
go run ./cmd/db-init
go run ./cmd/db-init --force
```

包装脚本：

- `seu-oj-backend/database/init.sh`
- `seu-oj-backend/database/init.ps1`

---

## 前置要求

| 依赖 | 说明 |
|------|------|
| Go | 后端、db-init、判题 Worker |
| MySQL | 配置在 config.yaml，可远程 |
| Redis | 判题队列，默认 127.0.0.1:6379 |
| Docker | Run/Submit 判题，需 gcc:13 镜像 |
| npm | 首次安装 CodeMirror（dev.sh 可自动执行） |

**不需要**：本机 mysql CLI（Windows 上常因插件问题不可用）。

---

## 覆盖 config.yaml（可选）

```bash
export DB_HOST=127.0.0.1
export DB_USER=root
export DB_PASSWORD=secret
export DB_NAME=seu_oj
bash scripts/dev.sh
```

PowerShell 参数：`-MySQLHost`、`-MySQLPassword` 等（见 `dev.ps1` 头部注释）。

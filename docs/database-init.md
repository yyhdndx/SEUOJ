# 数据库初始化指南

本文档说明 SEU OJ 首次部署时如何初始化 MySQL 数据库。

## 背景

- 后端通过 GORM `AutoMigrate` 可创建/补齐部分表，但**完整 schema 与演示数据建议用 SQL 导入**。
- 所有连接信息写在 **`seu-oj-backend/config/config.yaml`** 的 `database:` 段，与 Go 后端、`dev.sh` 共用。
- **不需要本机安装 mysql 命令行客户端**；推荐使用 Go 工具 `cmd/db-init`。

---

## 前置条件

1. MySQL 8.x 可访问（本地或远程均可）
2. 已配置 `config/config.yaml`（可从 `config.example.yaml` 复制）
3. 本机已安装 Go

示例配置：

```yaml
database:
  host: "127.0.0.1"      # 或远程 IP
  port: "3306"
  user: "root"
  password: "your_password"
  name: "seu_oj"         # 库名可自定义，如 seuoj
```

---

## 方式一：一键脚本（推荐）

在仓库根目录：

```bash
bash scripts/dev.sh --setup-only
```

会自动检测：若 `users` 表无数据则导入 schema + seed；若已初始化则跳过。

---

## 方式二：Go 工具 db-init（手动）

```bash
cd seu-oj-backend

go run ./cmd/db-init --ping    # 测试 MySQL 连接
go run ./cmd/db-init --check   # 检查是否已初始化（exit 0 = 已就绪）
go run ./cmd/db-init           # 首次导入（读 config.yaml）
go run ./cmd/db-init --force   # 强制重新导入（慎用）
```

包装脚本（等价于上面）：

```bash
bash seu-oj-backend/database/init.sh
powershell -ExecutionPolicy Bypass -File seu-oj-backend/database/init.ps1
```

### 工具说明

| 命令 | 作用 |
|------|------|
| `--ping` | `SELECT 1`，验证 config.yaml 中的连接可用 |
| `--check` | 检查目标库是否存在 `users` 表且有数据 |
| `--force` | 忽略已初始化状态，重新执行全部 SQL |
| （无参数） | 未初始化时导入；已初始化则跳过 |

---

## 方式三：手动 mysql 客户端（可选）

仅当本机 mysql 客户端正常时可用。Windows 若报 `mysql_native_password cannot be loaded`，请改用方式一/二。

在 `seu-oj-backend/database` 目录，按下列顺序执行：

### Schema（建表）

| 顺序 | 文件 |
|------|------|
| 1 | `user.sql` |
| 2 | `problems.sql` |
| 3 | `problem_testcases.sql` |
| 4 | `contests.sql` |
| 5 | `contest_problems.sql` |
| 6 | `contest_registrations.sql` |
| 7 | `contest_announcements.sql` |
| 8 | `submissions.sql` |
| 9 | `submission_results.sql` |
| 10 | `forum_topics.sql` |
| 11 | `forum_replies.sql` |
| 12 | `audit_logs.sql` |

### Seed（演示数据）

| 顺序 | 文件 |
|------|------|
| 1 | `seed_more_problems.sql` |
| 2 | `seed_classes_assignments.sql` |
| 3 | `seed_contests.sql` |
| 4 | `seed_forum.sql` |
| 5 | `seed_playlists.sql` |
| 6 | `seed_problem_solutions.sql` |
| 7 | `index.sql` |

---

## 演示账号（seed 导入后，密码均为 `123456`）

| 用户名 | 学工号 | 角色 |
|--------|--------|------|
| demo_admin | A-DEMO-2026 | admin |
| demo_teacher | T-DEMO-2026 | teacher |
| demo_alice | S-DEMO-001 | student |
| demo_bob | S-DEMO-002 | student |
| demo_cindy | S-DEMO-003 | student |

---

## 启动后 AutoMigrate

SQL 导入后，首次 `go run .` 会额外迁移 GORM 管理的扩展表（题单、班级、作业、论坛点赞收藏等），并对已有表做列补齐。

---

## 常见问题

**Q: 数据库配置写在哪？**  
A: `seu-oj-backend/config/config.yaml` 的 `database:` 段。`dev.sh`、`db-init`、Go 后端读同一份文件。

**Q: 只跑 `go run .` 不初始化可以吗？**  
A: AutoMigrate 能建部分表，但没有演示题目/账号，答辩仍需 seed。

**Q: seed 能重复执行吗？**  
A: 大部分 seed 幂等；`--force` 会重跑全部 SQL，远程库慎用。

**Q: 题单 seed 没有数据？**  
A: `seed_playlists.sql` 需要库中已有 teacher/admin，请先导入 `seed_classes_assignments.sql`。

**Q: Windows mysql 客户端报错怎么办？**  
A: 不要用 `mysql` CLI，改用 `go run ./cmd/db-init` 或 `bash scripts/dev.sh`。

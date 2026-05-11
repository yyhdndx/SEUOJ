# STORY-005 交付说明：题单训练进度（Playlists）

## 1. 数据库与表结构（实现前梳理）

以下表由 GORM `AutoMigrate` 或既有 SQL 维护，题单训练逻辑主要涉及 **playlists**、**playlist_problems**、**problems**、**submissions**。

### 1.1 `playlists`

| 字段 | 类型（概念） | 说明 |
|------|----------------|------|
| `id` | BIGINT PK AI | 题单主键 |
| `title` | VARCHAR(255) | 标题 |
| `description` | MEDIUMTEXT | 描述 |
| `visibility` | ENUM | `public` / `private` / `class` |
| `created_by` | BIGINT FK → users | 创建者 |
| `created_at` / `updated_at` | DATETIME | 时间戳 |

### 1.2 `playlist_problems`

| 字段 | 说明 |
|------|------|
| `playlist_id` | 题单 ID |
| `problem_id` | 题目主键（非 display_id） |
| `display_order` | 题单内顺序，从 1 递增 |

### 1.3 `problems`

| 字段 | 说明 |
|------|------|
| `id` | 题目主键（教师端表单填写的 problem_id） |
| `display_id` | 展示用题号 |
| `difficulty` | TINYINT：`0` 未知，`1` 简单，`2` 中等，`3` 困难 |

### 1.4 `submissions`（聚合进度时用到的列）

| 字段 | 说明 |
|------|------|
| `user_id` / `problem_id` | 归属用户与题目 |
| `status` | 判题结果；**Accepted** 与代码中其它模块一致（大小写敏感） |
| `created_at` | 提交时间 |
| `contest_id` | 比赛提交；题单进度统计中 **仅统计 `contest_id IS NULL`**，与班级分析等逻辑一致 |

### 1.5 设计选择说明

- **不在数据库新增表**：进度由提交记录实时聚合，避免同步状态表。
- **单次 IN 查询**：按题单内 `problem_id` 列表拉取当前用户相关提交，避免 N+1。
- **可选登录**：公开路由上增加「可选 JWT」解析；无 token 或 token 无效时视为游客，不报错。

---

## 2. 数据库操作记录（脚本）

### 2.1 新增文件

- 路径：`seu-oj-backend/database/seed_playlists.sql`
- 性质：幂等演示数据，可重复执行。
- **题单标题与描述均为英文 ASCII**，避免在 `latin1` 列上写入中文导致 **ERROR 1366**，也避免与连接字符集不一致时的 **ERROR 1267**。幂等键为固定英文 `title` 字符串（如 `Demo Basic Syntax Training`）。

### 2.2 脚本行为摘要

1. `START TRANSACTION`
2. 选取 `@teacher_id`：`users.role IN ('teacher','admin')`，优先 `teacher`。
3. 对 4 个固定标题的题单：
   - 若不存在则 `INSERT INTO playlists`；
   - `UPDATE` 描述与可见性（刷新元数据）；
   - `DELETE FROM playlist_problems WHERE playlist_id = ?`；
   - 按 `display_id` 子查询 `problems.id`，再 `INSERT INTO playlist_problems`（`display_order` 连续）。
4. `COMMIT` 后附带校验用 `SELECT`。

### 2.3 题单与题目（依赖 display_id）

| 题单标题（DB 中英文字符串） | visibility | display_id 顺序 |
|-----------------------------|------------|-----------------|
| `Demo Basic Syntax Training` | public | 1003, 1004, 1005, 1006 |
| `Demo Loops and Arrays` | public | 1004, 1005, 1006, 1007 |
| `Demo Strings Practice` | public | 1003, 1005, 1006, 1007 |
| `Demo Class Homework Playlist` | class | 1003–1007 |

**依赖**：建议先执行 `seed_more_problems.sql`，保证至少存在 **1003–1007**。

**已规避的常见错误**：

| 情况 | 行为 |
|------|------|
| 无 teacher/admin | `@teacher_id` 为 NULL，不插入新 `playlists`；每条 `playlist_problems` 的 `INSERT ... SELECT` 带 `AND @pl_* IS NOT NULL`，避免插入 `playlist_id = NULL` |
| 某 `display_id` 不存在 | 对应行插入 0 条，题单题目变少，不中断事务 |
| 列字符集为 latin1 | 脚本内无非 ASCII 字面量，`title`/`description` 可写入 latin1 |
| 与旧中文种子混用 | 旧题单标题不同，不会冲突；若需删掉旧中文演示数据需手工处理 |

### 2.4 建议本地执行顺序（MySQL 客户端）

**方式 A：仓库内 PowerShell 脚本（推荐，顺序执行两个 SQL）**

脚本 `run_seed_playlists.ps1` 默认从 **`seu-oj-backend/config/config.yaml`** 的 **`database:`** 段读取 `host`、`port`、`user`、`password`、`name`；只有你在命令行显式传入参数时才会覆盖对应项。

在 `seu-oj-backend/database` 目录下：

```powershell
# 全部使用 config.yaml（含其中的 password，若为空可再设 MYSQL_PWD）
.\run_seed_playlists.ps1
```

若题目已存在、只想重跑题单种子：

```powershell
.\run_seed_playlists.ps1 -SkipMoreProblems
```

仅用环境变量覆盖密码（不写进 yaml 时）：

```powershell
$env:MYSQL_PWD = '你的密码'
.\run_seed_playlists.ps1
```

密码优先级：**`-Password` > `MYSQL_PWD` > config.yaml 的 `database.password`**。

**PowerShell 执行策略**：若出现「禁止运行脚本」，可执行
`powershell -ExecutionPolicy Bypass -File .\run_seed_playlists.ps1`，或对当前用户放宽：
`Set-ExecutionPolicy -Scope CurrentUser -ExecutionPolicy RemoteSigned`。

**Bash（Git Bash / WSL）**：同目录提供 `run_seed_playlists.sh`，默认同样读取 `config/config.yaml`：

```bash
cd seu-oj-backend/database
chmod +x run_seed_playlists.sh   # 首次
./run_seed_playlists.sh
./run_seed_playlists.sh --skip-more-problems
```

**Git Bash 下 ERROR 2059**（`Authentication plugin 'mysql_native_password' cannot be loaded`）：多为本机 `mysql` 客户端过新或安装不完整，缺少认证插件。可选：

1. 使用 **MySQL 8.0** 安装目录里的 `mysql.exe`，并设置：
   `MYSQL_CLI="/c/Program Files/MySQL/MySQL Server 8.0/bin/mysql.exe" ./run_seed_playlists.sh`
2. 已安装 **Docker** 时：
   `export MYSQL_DOCKER_IMAGE=mysql:8.0`
   再执行 `./run_seed_playlists.sh`（容器内客户端连远程库，需本机能访问数据库网络）。
3. 有服务端权限时，可将用户改为 `caching_sha2_password`（与 DBA 操作相关，此处不展开）。

脚本同时接受 `--skip-more-problem`（单数）作为 `--skip-more-problems` 的别名。

**方式 B：mysql 客户端 + 重定向（CMD）**

```bat
cd /d E:\codes\advanced-ds\seuoj\seu-oj-backend\database
mysql -h 主机 -P 3306 -u 用户 -p --default-character-set=utf8mb4 seuoj < seed_more_problems.sql
mysql -h 主机 -P 3306 -u 用户 -p --default-character-set=utf8mb4 seuoj < seed_playlists.sql
```

**方式 C：已登录 mysql 交互式**

```text
SOURCE E:/codes/advanced-ds/seuoj/seu-oj-backend/database/seed_more_problems.sql;
SOURCE E:/codes/advanced-ds/seuoj/seu-oj-backend/database/seed_playlists.sql;
```

（路径按你本机调整；Windows 下也可用正斜杠。）

### 2.5 验收用 SQL（与 Story 一致）

```sql
SELECT id, title, visibility, created_by FROM playlists
WHERE title LIKE 'Demo %'
ORDER BY id;

SELECT p.title, pp.display_order, pr.display_id, pr.title AS problem_title
FROM playlist_problems pp
JOIN playlists p ON p.id = pp.playlist_id
JOIN problems pr ON pr.id = pp.problem_id
WHERE p.title LIKE 'Demo %'
ORDER BY p.id, pp.display_order;
```

---

## 3. 后端改动摘要

| 模块 | 说明 |
|------|------|
| `internal/middleware/optional_jwt.go` | 可选 JWT：有合法 Bearer 则写入 `user_id` / `role`，否则继续 |
| `internal/router/router.go` | `GET /api/playlists/:id` 挂载 `OptionalJWTAuth` |
| `internal/api/teaching.go` | `PlaylistDetail` 从上下文读取可选用户，传入 Service |
| `internal/dto/teaching.go` | `PlaylistProgress`；扩展 `PlaylistProblemItem`；`PlaylistDetailResponse.Progress` |
| `internal/service/teaching_service.go` | 题单题目联表带出 `difficulty`；按用户聚合提交；计算 `progress` 与 `next_problem_id` |

### 3.1 进度口径

- `solved_count`：本题在题单中至少一次 `Accepted`（非比赛提交）。
- `attempted_count`：有提交且当前状态不是 `accepted` 的题目数（仅 `attempted`，不含 `not_started`）。
- `next_problem_*`：顺序上第一道非 `accepted`；若全部 `accepted` 则指向第一题（复习）。

---

## 4. 前端改动摘要

| 文件 | 说明 |
|------|------|
| `seu-oj-frontend/js/teaching.js` | 题单详情训练布局、进度条、继续练习、题目表列；`validatePlaylistProblemIdsInput`；教师选题参考行展示 id / display_id / 难度 |
| `seu-oj-frontend/css/teaching.css` | 进度卡片、进度条、难度 pill、表格横向滚动 |

公开题单请求仍走 `apiFetch`：登录时自动带 Bearer，后端返回个人进度；未登录无 token，后端返回全零与 `not_started`。

---

## 5. 文档

- `docs/api.md` 第 9.2 节已更新为上述契约说明。

---

## 6. 验证命令（本地）

```powershell
Set-Location seu-oj-backend
go test ./...
```

```powershell
node --check seu-oj-frontend/js/teaching.js
```

浏览器：`#/playlists`、`#/playlists/:id`、`#/teacher/playlists`（创建/编辑校验）。

---

## 7. 已知风险 / 未做范围

- `class` 可见性题单的班级维度权限未在本 Story 扩展（与 Story 说明一致）。
- 提交统计不含比赛场 `contest_id` 非空的记录；若需「比赛 AC 也算题单进度」需另开需求。

---

## 8. 文件清单

- 新增：`seu-oj-backend/database/seed_playlists.sql`
- 新增：`seu-oj-backend/internal/middleware/optional_jwt.go`
- 修改：`seu-oj-backend/internal/router/router.go`
- 修改：`seu-oj-backend/internal/api/teaching.go`
- 修改：`seu-oj-backend/internal/dto/teaching.go`
- 修改：`seu-oj-backend/internal/service/teaching_service.go`
- 修改：`seu-oj-frontend/js/teaching.js`
- 修改：`seu-oj-frontend/css/teaching.css`
- 修改：`docs/api.md`

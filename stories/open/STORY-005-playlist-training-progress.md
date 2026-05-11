# STORY-005-playlist-training-progress

- Status: `todo`
- Priority: `P1`
- Assignee: `unassigned`
- Reviewer: `unassigned`
- Created At: `2026-04-30`
- Target Version: `v0.1`

## Background

当前题单模块已经具备基础能力：
- 学生 / 游客可以浏览公开题单列表和题单详情
- 教师可以创建、编辑、删除题单
- 班级作业可以引用题单
- 题单中的题目已经有固定顺序

但目前题单更像一个“静态题目集合”，还没有成为真正的训练路径。主要问题包括：
- 学生进入题单后只能看到题目列表，不知道自己完成到哪里
- 题单详情页没有显示每道题的个人完成状态
- 没有“继续练习”入口，用户需要自己判断下一题做什么
- 题单没有汇总进度，例如 `3 / 10 Accepted`
- 题目行信息较少，没有和题库中的难度字段、提交状态打通
- 教师创建题单仍依赖手动输入 problem id，使用成本较高

这个 story 的核心目标是先把学生侧题单从“列表页”升级为“训练页”。教师侧组题体验可以在本 story 中做轻量优化，但不应抢占主线。

## Goal

将题单详情页改造成可持续训练的页面，让学生进入题单后能快速回答：
- 这个题单有多少题？
- 我已经通过了多少题？
- 哪些题没有开始？
- 哪些题提交过但还没通过？
- 下一题应该做哪一道？

完成后，题单模块应具备明确的训练感：
- 题单详情页展示个人进度
- 每道题展示个人状态
- 支持一键继续练习
- 题目信息与题库字段保持一致
- 未登录用户仍能正常浏览公开题单，但不展示个人进度

## Scope

本次包含：

### 1. 题单详情增加用户进度

登录用户打开题单详情时，需要返回当前用户在该题单中的训练进度。

建议增加字段：
- `problem_count`
- `solved_count`
- `attempted_count`
- `next_problem_id`
- `next_problem_display_id`
- `progress_percent`

进度口径：
- `solved_count`：题单中至少有一次 `Accepted` 提交的题目数
- `attempted_count`：题单中有过提交但没有 `Accepted` 的题目数
- `problem_count`：题单总题数
- `progress_percent`：`solved_count / problem_count`
- `next_problem_id`：题单顺序中第一道未 Accepted 的题目

游客或未登录用户：
- 可以继续访问公开题单
- `solved_count`、`attempted_count` 可以返回 `0`
- 每题状态显示为 `not_started` 或不返回个人状态
- 前端需要显示“登录后查看个人进度”的轻提示

### 2. 每道题增加训练状态

题单详情中的每个题目需要增加个人状态字段。

建议字段：
- `status`
- `last_submission_id`
- `last_submission_status`
- `last_submitted_at`
- `accepted_at`
- `difficulty`

状态枚举建议：
- `not_started`：没有任何提交
- `attempted`：有提交，但没有 Accepted
- `accepted`：至少有一次 Accepted

展示要求：
- `accepted` 使用成功状态标签
- `attempted` 使用警告 / 待处理状态标签
- `not_started` 使用中性状态标签
- 题目难度使用 `easy / medium / hard` 标签

### 3. 后端查询逻辑改造

后端需要在题单详情查询时聚合当前用户提交情况。

建议实现方式：
- 保留现有 `GET /api/playlists/:id`
- 如果请求带有登录态，则在详情中附带个人进度
- 如果没有登录态，则只返回公开题单基础信息

提交状态聚合逻辑：
- 先查询题单中的所有 `problem_id`
- 再查询当前用户对这些题目的提交
- 按 `problem_id` 聚合：
  - 是否存在 `Accepted`
  - 最近一次提交状态
  - 最近一次提交时间
  - 最近一次提交 ID
  - 首次或最近 Accepted 时间

需要注意：
- 不要为每道题单独查询提交，避免 N+1 查询
- 查询应只限制在当前题单的问题集合中
- 当前题单题目数一般不大，可以用 `WHERE problem_id IN ?`

### 4. 题单详情页前端重构

调整 `#/playlists/:id` 页面，让它从简单详情页变成训练工作台。

页面建议结构：
- 顶部：题单标题、可见性、返回按钮
- 进度区：总题数、已通过、已尝试、完成百分比
- 主操作：继续练习按钮
- 描述区：题单说明
- 题目区：按顺序展示题目状态

题目列表建议列：
- 顺序
- Display ID
- 标题
- 难度
- 我的状态
- 最近提交
- 操作

操作按钮：
- 未通过题目显示 `Solve`
- 已通过题目显示 `Review`
- 最近提交可跳转到提交详情或提交列表筛选页，如果当前项目已有对应路由则复用

### 5. 继续练习入口

题单详情页需要提供一个明显的 `Continue Practice` 入口。

逻辑：
- 如果存在未 Accepted 的题目，跳转到第一道未 Accepted 题
- 如果全部 Accepted，跳转到第一道题，按钮文案可为 `Review Playlist`
- 未登录用户点击时跳转到第一道题即可，不强制登录

跳转方式：
- 复用现有题目详情路由：`#/problems/:problem_id`

### 6. 公开题单列表轻量增强

题单列表页可以做轻量增强，但不作为主要复杂功能。

建议增加：
- 题目数量更突出
- 可见性标签更清晰
- 描述截断保持统一
- 登录用户如果后端已返回进度，可展示简短进度，例如 `3 / 10`

如果列表接口暂不返回个人进度，本 story 可以只改详情页，不强制修改列表接口。

### 7. 教师侧组题体验轻量优化

教师侧本次不要求完整拖拽选题器，但建议至少降低手动输入错误。

可以做这些轻量优化：
- 输入框 placeholder 更明确，说明这里需要数据库 `problem_id`
- 在可选题目列表中同时显示：
  - `id`
  - `display_id`
  - `title`
  - `difficulty`
- 创建 / 保存前在前端检查：
  - 题目 ID 是否为空
  - 是否有重复 ID
  - 是否存在非数字内容
- 错误提示应明确告诉教师是哪一个 ID 有问题

完整的可视化选题器建议放到后续 story。

### 8. 手动插入演示题单数据

本 story 需要准备可演示的题单数据，否则题单训练进度页面缺少稳定验证入口。

当前仓库中没有看到专门的 playlist seed SQL，因此本次需要补充一个可手动执行、可重复执行的题单 seed 脚本。

建议新增：
- `seu-oj-backend/database/seed_playlists.sql`

题单数据至少包含 3 个：
- `基础语法训练`：`public`，用于游客和普通用户访问
- `循环与数组训练`：`public`，用于展示多题训练进度
- `课堂作业题单示例`：`class`，用于教师端和后续作业绑定场景

每个题单建议包含 4 到 6 道题，题目来源优先使用当前已有题目：
- `1001` 到 `1025` 的已有题目
- 本地已补充的 `1026` 到 `1030` 题目

插入时不要直接依赖题目表自增 `id`。推荐通过 `display_id` 查询实际 `problem_id`，再写入 `playlist_problems`。

脚本要求：
- 可以重复执行
- 不重复创建同名题单
- 重跑时可以刷新题单题目顺序
- `created_by` 使用已有教师或管理员用户 ID
- 如果本地没有教师账号，需要在脚本注释中说明先创建教师 / 管理员账号

建议脚本风格：
- 使用事务
- 先查 `users` 中的教师或管理员作为 `created_by`
- 通过 `display_id` 查询题目 ID
- 先插入或更新 `playlists`
- 再删除并重建对应 `playlist_problems`

## Out of Scope

本次不包含：
- 拖拽式题单编辑器
- 题单收藏 / 点赞 / 评论
- 根据用户能力推荐题单
- 题单学习路径树
- 班级作业统计大屏重做
- 判题逻辑修改
- 提交记录页面大改
- 新增复杂权限体系

## Related Modules

- `seu-oj-backend/internal/model/playlist.go`
- `seu-oj-backend/internal/model/playlist_problem.go`
- `seu-oj-backend/internal/model/problem.go`
- `seu-oj-backend/internal/model/submission.go`
- `seu-oj-backend/internal/dto/teaching.go`
- `seu-oj-backend/internal/service/teaching_service.go`
- `seu-oj-backend/internal/api/teaching.go`
- `seu-oj-backend/internal/router/router.go`
- `seu-oj-frontend/js/teaching.js`
- `seu-oj-frontend/css/teaching.css`
- `seu-oj-backend/database/seed_playlists.sql`
- `docs/api.md`

## Suggested Files

后端建议修改：
- `seu-oj-backend/internal/dto/teaching.go`
- `seu-oj-backend/internal/service/teaching_service.go`
- `seu-oj-backend/internal/api/teaching.go`

前端建议修改：
- `seu-oj-frontend/js/teaching.js`
- `seu-oj-frontend/css/teaching.css`

数据库演示数据建议新增：
- `seu-oj-backend/database/seed_playlists.sql`

通常不需要修改：
- `seu-oj-backend/internal/router/router.go`

除非决定新增独立接口，例如 `GET /api/playlists/:id/progress`。

## API Design Notes

优先方案：扩展现有题单详情接口。

### `GET /api/playlists/:id`

公开访问时返回基础题单信息。

登录访问时额外返回 `progress` 和每题个人状态。

建议响应结构：

```json
{
  "id": 1,
  "title": "Week 1 Training",
  "description": "Basic loops and conditionals",
  "visibility": "public",
  "created_by": 2,
  "created_at": "2026-04-30T10:00:00+08:00",
  "updated_at": "2026-04-30T10:00:00+08:00",
  "progress": {
    "problem_count": 10,
    "solved_count": 3,
    "attempted_count": 2,
    "progress_percent": 30,
    "next_problem_id": 8,
    "next_problem_display_id": "1008"
  },
  "problems": [
    {
      "problem_id": 3,
      "display_order": 1,
      "display_id": "1003",
      "title": "A + B Problem",
      "difficulty": "easy",
      "status": "accepted",
      "last_submission_id": 15,
      "last_submission_status": "Accepted",
      "last_submitted_at": "2026-04-30T10:30:00+08:00",
      "accepted_at": "2026-04-30T10:30:00+08:00"
    }
  ]
}
```

兼容性要求：
- 原有字段不能删除
- 前端应兼容没有 `progress` 字段的响应
- 未登录请求不能因为缺少用户 ID 而失败

## Data Preparation Tasks

### 1. 新增题单 seed 脚本

新增 `seu-oj-backend/database/seed_playlists.sql`，用于本地开发、课堂演示和验收。

脚本需要至少创建这些题单：

| 题单标题 | Visibility | 建议题目 | 用途 |
| --- | --- | --- | --- |
| 基础语法训练 | `public` | `1001, 1002, 1026, 1027` | 游客 / 学生浏览公开题单 |
| 循环与数组训练 | `public` | `1003, 1004, 1028, 1030` | 展示继续练习和部分完成进度 |
| 字符串与综合练习 | `public` | `1005, 1029, 1010, 1011` | 展示不同难度和较长列表 |
| 课堂作业题单示例 | `class` | `1026, 1027, 1028, 1029, 1030` | 教师端和作业绑定预留 |

如果某些 `display_id` 在本地库不存在，可以替换为当前数据库中存在的题目，但需要保持：
- 至少 3 个公开题单
- 至少 1 个 `class` 题单
- 每个题单至少 3 道题

### 2. Seed 脚本幂等要求

脚本需要满足：
- 重复执行不会创建重复题单
- 重复执行会刷新题单描述、可见性和题目顺序
- 如果题目不存在，应尽量在脚本注释中说明依赖，而不是静默插入空题单
- `playlist_problems.display_order` 从 `1` 开始连续递增

推荐实现方式：

```sql
START TRANSACTION;

SELECT id INTO @teacher_id
FROM users
WHERE role IN ('teacher', 'admin')
ORDER BY FIELD(role, 'teacher', 'admin'), id
LIMIT 1;

-- 如果 @teacher_id 为空，需要先创建教师或管理员账号。

-- 示例：通过 display_id 查题目实际 id，再插入 playlist_problems。
-- 具体 SQL 可以按 seed_more_problems.sql 的幂等风格实现。

COMMIT;
```

### 3. 演示提交数据准备

为了验证 `not_started / attempted / accepted` 三种状态，建议手动准备或通过页面产生提交记录：
- 一个用户完全没有提交过某个题单
- 一个用户对某道题提交过但未通过
- 一个用户对某道题已有 `Accepted`

不建议在本 story 里强行插入大量 submission seed。优先通过真实运行 / 提交流程产生数据，这样也能顺便验证题目页和判题链路。

## Backend Tasks

### 1. DTO 扩展

在 `teaching.go` 中增加：
- `PlaylistProgress`
- 扩展 `PlaylistProblemItem`
- 扩展 `PlaylistDetailResponse`

建议字段：
- `PlaylistDetailResponse.Progress *PlaylistProgress`
- `PlaylistProblemItem.Difficulty string`
- `PlaylistProblemItem.Status string`
- `PlaylistProblemItem.LastSubmissionID *uint64`
- `PlaylistProblemItem.LastSubmissionStatus string`
- `PlaylistProblemItem.LastSubmittedAt *time.Time`
- `PlaylistProblemItem.AcceptedAt *time.Time`

### 2. API 层读取可选登录态

当前公开题单详情是公开接口。需要支持“未登录可访问，登录则带个人进度”。

建议：
- 在 `PlaylistDetail` handler 中尝试读取用户 ID 和角色
- 如果没有登录态，不返回错误
- 如果有登录态，传入 service 用于计算进度

注意：
- 不能把公开题单详情改成强制登录
- 不能影响游客浏览公开题单

### 3. Service 层增加进度聚合

在 `GetPlaylistDetail` 中：
- 查询题单和题单题目
- 题目查询需要带出 `p.difficulty`
- 当 `requestUserID > 0` 时，额外查询 submissions
- 将提交状态合并到 `problems`
- 计算 `progress`

建议拆出私有函数：
- `buildPlaylistProgress(problems []dto.PlaylistProblemItem) dto.PlaylistProgress`
- `loadPlaylistSubmissionStatus(userID uint64, problemIDs []uint64) (map[uint64]playlistProblemStatus, error)`

### 4. 查询性能要求

实现时避免：
- 对每个题目单独查一次 submissions
- 在前端循环请求提交记录

建议使用一次查询获取当前用户在题单内所有题目的提交：

```sql
SELECT id, problem_id, status, created_at
FROM submissions
WHERE user_id = ?
  AND problem_id IN (?)
ORDER BY problem_id ASC, created_at DESC
```

然后在 Go 代码中按 `problem_id` 聚合。

### 5. 权限保持现状

访问规则保持当前语义：
- `public` 题单：游客可看
- `private` 题单：创建者或管理员可看
- `class` 题单：本 story 不扩展班级可见性规则

如果发现当前 `class` 可见性逻辑不完整，只记录为风险，不在本 story 内扩大处理。

## Frontend Tasks

### 1. 重构 `renderPlaylistDetail`

目标页面结构：
- Header：标题、可见性、返回按钮
- Progress Summary：总题数 / 已通过 / 已尝试 / 完成百分比
- Primary Action：继续练习
- Description：题单描述
- Problem Table：题目训练状态列表

要求：
- 进度摘要优先展示，不要埋在描述下面
- 继续练习按钮在首屏可见
- 空题单仍然显示友好空状态

### 2. 新增渲染辅助函数

建议新增：
- `renderPlaylistProgress(progress, problems)`
- `renderPlaylistProblemStatus(status)`
- `renderPlaylistDifficulty(difficulty)`
- `findNextPlaylistProblem(detail)`
- `getPlaylistContinueHref(detail)`

这些函数放在 `teaching.js` 中即可，不需要新建文件。

### 3. 题目表格交互

每行题目至少包含：
- 序号
- display id
- 标题链接
- 难度
- 我的状态
- 最近提交
- 操作按钮

链接规则：
- 标题和操作按钮跳转到 `#/problems/:problem_id`
- 最近提交如果有 `last_submission_id`，优先跳转提交详情；如果当前项目没有详情路由，则跳转提交列表

### 4. 未登录态处理

未登录用户打开公开题单时：
- 不显示错误
- 可以看题目列表
- 状态区域显示“登录后查看个人进度”
- 继续练习跳转第一道题

### 5. 教师侧轻量校验

在创建 / 编辑题单提交前：
- 检查是否至少有一个题目 ID
- 检查是否有重复 ID
- 检查是否全部是正整数
- 出错时使用 `setFlash` 明确提示

可选增强：
- 可选题目列表显示 `display_id` 和 `difficulty`
- 降低教师只看到数据库 id 时的困惑

## UI / Style Tasks

在 `teaching.css` 中补充题单训练相关样式：
- 进度摘要卡片
- 进度条
- 状态标签
- 难度标签
- 继续练习按钮区域
- 移动端表格或卡片式降级布局

要求：
- 不新增大面积新主题色
- 复用已有 `status-pill`、`detail-card`、`data-table` 风格
- 桌面端信息密度适中
- 移动端不能出现明显横向溢出

## Expected Behavior

完成后应满足：
- 登录用户打开题单详情能看到自己的完成进度
- 每道题能看出是未开始、已尝试还是已通过
- 点击继续练习能进入第一道未通过题
- 全部通过后继续练习变成复习入口
- 未登录用户仍能浏览公开题单
- 教师创建题单时明显减少 ID 输入错误

## Acceptance Criteria

- [ ] `GET /api/playlists/:id` 对未登录用户仍可访问公开题单
- [ ] 登录用户访问题单详情时返回进度信息
- [ ] 每道题返回 `difficulty` 和个人训练状态
- [ ] 题单详情页展示总题数、已通过数、已尝试数和完成百分比
- [ ] 题单详情页提供可用的继续练习按钮
- [ ] 未开始、已尝试、已通过三类状态视觉上可区分
- [ ] 题目行可以跳转到对应题目详情页
- [ ] 教师创建 / 编辑题单时能拦截空 ID、重复 ID、非法 ID
- [ ] 前端在没有 `progress` 字段时不会崩溃
- [ ] 后端查询没有明显 N+1 问题
- [ ] 新增可重复执行的 `seed_playlists.sql`
- [ ] 本地数据库至少存在 3 个可用于演示的题单
- [ ] API 文档同步更新

## Verification

建议至少完成这些验证：

### 数据准备验证

执行题单 seed 脚本后检查：

```sql
SELECT id, title, visibility, created_by
FROM playlists
ORDER BY id DESC;

SELECT p.title, pp.display_order, pr.display_id, pr.title AS problem_title
FROM playlist_problems pp
JOIN playlists p ON p.id = pp.playlist_id
JOIN problems pr ON pr.id = pp.problem_id
ORDER BY p.id, pp.display_order;
```

检查点：
- 至少有 3 个公开题单
- 至少有 1 个 `class` 题单
- 每个题单至少有 3 道题
- 每个题单题目顺序从 1 开始连续
- `#/playlists` 能看到公开题单
- `#/teacher/playlists` 能看到当前教师或管理员创建的题单

### 后端验证

```powershell
cd seu-oj-backend
go test ./...
```

手动检查：
- 未登录请求公开题单详情
- 登录普通用户请求公开题单详情
- 登录教师请求自己创建的题单详情
- 普通用户请求 private 题单应被拒绝

### 前端验证

```powershell
node --check seu-oj-frontend/js/teaching.js
```

浏览器检查：
- `#/playlists`
- `#/playlists/:id`
- `#/teacher/playlists`
- `#/teacher/playlists/:id`

重点场景：
- 用户没有提交过题单内任何题目
- 用户提交过但没有 Accepted
- 用户部分题目 Accepted
- 用户全部题目 Accepted
- 游客访问公开题单

## Risks / Notes

- 当前公开题单详情接口可能没有经过登录中间件，因此需要小心处理“可选登录态”。
- 如果当前 JWT 中间件不支持可选登录态，可以优先新增独立接口 `GET /api/playlists/:id/progress`，但需要前端多请求一次。
- `class` 可见性目前可能没有完整绑定班级成员权限，本 story 不建议扩大处理。
- 如果题目难度字段在当前分支还没有完全落库，需要先完成 STORY-003 中 difficulty 的后端字段和 DTO 贯通。
- 提交状态聚合需要注意 `Accepted` 的大小写与当前系统实际状态值保持一致。
- 题单 seed 依赖已有教师 / 管理员账号；如果本地数据库没有这类账号，需要先注册或手动更新用户角色。
- 题单 seed 不应硬编码 `problem_id`，应优先通过 `display_id` 查询，避免不同环境自增 ID 不一致。

## Handover Notes

建议实现顺序：

1. 先确认本地有教师或管理员账号。
2. 新增并执行 `seed_playlists.sql`，准备 3 到 4 个可演示题单。
3. 扩展 DTO 和 service，确保题单详情能返回 `difficulty`。
4. 加入登录用户的提交状态聚合。
5. 完成 `progress` 计算和 `next_problem_id`。
6. 更新题单详情前端页面。
7. 增加教师侧 ID 输入校验。
8. 更新 `docs/api.md`。
9. 执行 Go 测试和前端语法检查。

如果时间有限，最低交付范围是：
- 后端返回题目难度和个人状态
- 前端展示进度和继续练习
- 保持游客访问公开题单不受影响
- 至少手动插入 3 个题单，保证页面有稳定演示数据

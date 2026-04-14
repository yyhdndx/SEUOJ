# STORY-003-problem-workbench-redesign

- Status: `todo`
- Priority: `P0`
- Assignee: `unassigned`
- Reviewer: `unassigned`
- Created At: `2026-04-14`
- Target Version: `v0.1`

## Background

当前题目详情页已经具备基本做题能力，包括：
- 题面展示
- 样例运行
- 代码提交
- 最近提交查看

但从实际使用、演示效果和后续继续维护的角度看，题目页仍存在明显问题：

1. 标题区域层次不清楚，题目名称不够突出
2. 题目缺少 `difficulty` 字段，题目信息表达不完整
3. 左侧内容混排，题面、统计、题解、最近提交没有形成稳定分区
4. 页面中仍保留开发阶段遗留信息，例如：
   - `1 sample case(s) are available.`
   - `sort_order=1, active=true`
5. 题面整体字号偏小，阅读体验一般
6. 右侧代码区仍是基础文本输入框，缺少代码高亮和更好的编辑体验
7. 测试结果区域只有在点击 `Run` 后才出现，没有固定工作区结构
8. 当前不支持用户自定义测试输入
9. 当前 draft 主要依赖本地存储，缺少服务端草稿能力，导致跨设备或刷新后恢复能力有限

这个 story 的目标不是做“局部修补”，而是把题目页明确收敛为一个可持续维护的做题工作台。

## Goal

将题目详情页重构为更清晰、更稳定的双栏工作台，整体体验接近 LeetCode 这类在线做题界面，但保持当前项目的轻量实现方式。

完成后，题目页应满足以下要求：
- 左侧用于阅读和理解题目
- 右侧用于编写、运行和提交代码
- 题面信息层次清楚，阅读成本低
- 代码工作区结构稳定，不依赖临时弹出的结果面板
- 每个用户在每道题上都能恢复最近一次草稿代码

## Scope

本次包含：

### 1. 标题与基础信息重构
- 强化题目标题视觉层级
- 调整副标题信息排列
- 为题目新增并展示 `difficulty`
  - 推荐值：`easy / medium / hard`

### 2. 左侧题面区重构
将左侧内容拆成独立区块，而不是继续大块混排。建议区块包括：
- Description
- Problem Stats
- Solutions
- My Recent Submissions

要求：
- 各区块边界清楚
- 区块顺序稳定
- 区块标题明确

### 3. 清理不适合用户展示的开发遗留内容
需要移除或隐藏这些信息：
- `sample case count` 这类调试性文案
- `sort_order`
- `active`
- 其他内部状态或开发遗留字段

### 4. 题面阅读体验优化
- 调整题目页字体与间距
- 让题面内容更适合长时间阅读
- 避免信息堆叠过于紧密

### 5. 右侧代码编辑区升级
右侧不再只是普通文本框，而要具备基本代码编辑体验。至少实现：
- 代码高亮
- 更适合在线写代码的输入体验
- 清晰的编辑区域边界

实现上允许采用轻量编辑器方案，但不要求一步到位做成完整 IDE。

### 6. 测试结果区重构
测试结果区改为常驻面板，而不是只有点击 `Run` 之后才出现。

要求：
- 未运行代码时显示默认提示，例如“请先运行代码”
- 运行后显示样例测试结果
- 保持题目页右侧工作区结构稳定

### 7. 自定义测试输入
支持用户自行输入测试内容并执行本地运行。

要求：
- 自定义输入不要求 `Expected`
- 仅用于本地运行测试
- 不进入正式评测结果

### 8. 服务端草稿能力
为每个用户-题目组合增加服务端草稿存储能力。

要求：
- 打开题目页时默认加载该用户该题最近草稿
- 草稿至少保存：语言、代码、更新时间
- 草稿来源优先以最近一次提交后的代码为基础
- 不再只依赖浏览器本地存储

## Out of Scope

本次不包含：
- 重写判题逻辑
- 引入多文件项目提交
- 引入多人协同编辑
- 重做题库列表页
- 引入题目收藏、做题笔记、标签系统
- 一步到位实现完整 IDE 功能

## Related Modules

- `seu-oj-backend/internal/model/problem.go`
- `seu-oj-backend/internal/model/submission.go`
- `seu-oj-backend/internal/dto/problem.go`
- `seu-oj-backend/internal/dto/submission.go`
- `seu-oj-backend/internal/service/problem_service.go`
- `seu-oj-backend/internal/service/submission_service.go`
- `seu-oj-backend/internal/api/problem.go`
- `seu-oj-backend/internal/api/submission.go`
- `seu-oj-backend/internal/database/db.go`
- `seu-oj-frontend/js/problems.js`
- `seu-oj-frontend/css/problem.css`
- `docs/api.md`

## Suggested Files

- `seu-oj-backend/internal/model/problem.go`
- `seu-oj-backend/internal/model/submission.go`
- `seu-oj-backend/internal/service/problem_service.go`
- `seu-oj-backend/internal/service/submission_service.go`
- `seu-oj-backend/internal/api/problem.go`
- `seu-oj-backend/internal/api/submission.go`
- `seu-oj-backend/internal/database/db.go`
- `seu-oj-frontend/js/problems.js`
- `seu-oj-frontend/css/problem.css`
- `docs/api.md`

## Data / Schema Notes

### 1. 题目难度字段
建议在 `problems` 表增加：
- `difficulty`

推荐取值：
- `easy`
- `medium`
- `hard`

### 2. 草稿存储方案
草稿不建议继续只依赖浏览器本地存储。

候选方案：
1. 在 `submissions` 表中增加草稿相关字段
2. 单独新增 `problem_drafts` 表

推荐方案：
- 单独新增 `problem_drafts` 表

建议字段：
- `id`
- `user_id`
- `problem_id`
- `language`
- `code`
- `updated_at`

采用单独表的原因：
- draft 与正式 submission 生命周期不同
- 正式提交记录不应混入“未提交草稿”语义
- 后续更适合做自动保存、恢复和覆盖更新

## Page Entry

重点检查这些页面：
- `#/problems/:id`
- `#/contests/:contestId/problems/:problemId`
- 题目页中的 `Run`
- 题目页中的 `Submit`

## Expected Behavior

完成后，页面至少应满足：
- 题目标题明显突出，难度清晰可见
- 左侧按模块分区展示，而不是内容混排
- 页面中不再出现开发遗留字段和内部提示
- 右侧代码区具备代码高亮或等价在线编辑体验
- 测试结果区始终存在，未运行时显示默认提示
- 用户可以输入自定义测试内容并运行
- 页面加载时，默认恢复该用户该题最近草稿
- 普通题目页和比赛题目页保持一致的工作流结构

## Acceptance Criteria

- [ ] 题目标题区域层次清晰，难度字段已接入并展示
- [ ] 左侧模块明确拆分为题面、统计、题解、最近提交等区块
- [ ] 页面中不再显示 `sample case count`、`sort_order`、`active` 等开发遗留内容
- [ ] 右侧代码区具备代码高亮或等价在线编辑体验
- [ ] 测试结果区未运行时能稳定显示默认提示
- [ ] 用户可输入自定义测试内容并运行
- [ ] 每个用户打开题目时能加载该题最近草稿
- [ ] 普通题目页和比赛题目页都已回归检查
- [ ] 前端脚本语法检查和后端构建通过

## Verification

建议至少完成这些验证：
- 打开普通题目页，检查题目结构、难度、统计、题解、最近提交
- 打开比赛题目页，确认工作台结构一致
- 测试默认样例 `Run`
- 测试自定义输入 `Run`
- 提交一份代码后刷新页面，确认草稿能重新载入
- 检查页面中不再出现开发期残留提示
- 运行：

```powershell
node --check "D:\desk\软件工程\SEUOJ\seu-oj-frontend\js\problems.js"
cd "D:\desk\软件工程\SEUOJ\seu-oj-backend"
$env:GOCACHE="D:\desk\软件工程\SEUOJ\seu-oj-backend\.gocache"
go build ./...
```

## Suggested Frontend Tech

本 story 允许引入轻量前端渲染增强技术，但不建议重构整站框架。

推荐技术：
- `CodeMirror`：用于题目页代码编辑器升级
- `markdown-it` 或 `marked`：用于题面、题解等 Markdown 渲染统一
- `dayjs`：用于题目页和最近提交时间格式化

技术约束：
- 不将当前前端整体重构为 React/Vue
- 不引入重型后台 UI 框架
- 优先采用可局部接入、对现有静态 SPA 侵入较小的方案

## Risks / Notes

- 代码编辑器升级可能引入额外静态资源或依赖，需要控制集成复杂度
- 如果采用 Monaco/CodeMirror，需要确认当前静态前端结构是否适合接入
- 服务端 draft 方案需要明确与正式 submission 的边界
- 题目页重构范围较大，必须注意不要破坏现有提交、运行和比赛上下文逻辑

## Handover Notes

这是一个跨前后端的 story，建议拆成两个实现子任务：
1. 后端字段与草稿数据能力
2. 前端题目工作台重构

如果时间有限，优先级建议如下：
1. `difficulty` 字段
2. 左侧区块重构
3. 常驻 Run Result 面板
4. 服务端 draft 能力
5. 代码编辑器升级
6. 自定义测试输入

## Architecture Decision Note

关于“题目描述、用户代码、富文本内容是否改用 MongoDB”这一点，本项目当前阶段不建议切换到 MongoDB。

推荐继续使用 MySQL，原因如下：
- 当前系统主体是强关系模型，用户、题目、提交、比赛、班级之间关系明确
- 题目描述、题解、论坛正文、草稿代码这类内容使用 MySQL 的 `TEXT` / `MEDIUMTEXT` 已足够
- 引入 MongoDB 会增加额外的数据存储、运维和同步成本，不适合当前轻量团队
- 文本内容较长并不自动意味着应切换到文档数据库

结论：
- 当前阶段继续使用 MySQL
- 题目描述、题解、论坛正文、代码草稿均继续放在 MySQL 中
- 只有在未来出现明确的非结构化数据扩张场景时，再单独评估 MongoDB



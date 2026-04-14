# STORY-002-forum-search-and-zones

- Status: `todo`
- Priority: `P1`
- Assignee: `unassigned`
- Reviewer: `unassigned`
- Created At: `2026-04-14`
- Target Version: `v0.1`

## Background

当前论坛模块已经具备这些基础能力：
- 发帖
- 回帖
- 置顶
- 锁帖
- 从题目页 / 比赛页进入讨论区

但论坛首页仍然偏简陋，主要问题包括：
- 帖子列表缺少更明确的分区感
- 用户无法通过关键字快速定位讨论
- 题目讨论、比赛讨论和综合讨论混在一起时，可读性一般
- 当前论坛更像“能发帖”，还不像“可持续使用的讨论区”

## Goal

增强论坛首页的信息组织能力，使论坛列表页可以用于日常浏览、快速定位和课堂展示。

目标是让用户在论坛首页能够快速回答这几个问题：
- 这条帖子属于哪一类讨论？
- 是否和某道题或某场比赛有关？
- 我能否通过关键字快速找到它？

## Scope

本次包含：
- 为论坛列表增加关键字搜索
- 增加或强化论坛分区展示与筛选入口
- 优化帖子卡片中的元信息展示，例如：
  - 类型
  - 是否置顶
  - 是否锁帖
  - 关联题目 / 比赛
  - 作者与回复数
- 检查题目页、比赛页进入论坛讨论的上下文是否仍然正确
- 回归帖子详情页加载和回帖流程

## Out of Scope

本次不包含：
- 新增点赞、收藏、通知系统
- 重做论坛数据库结构
- 引入复杂的全文检索方案
- 大规模重做论坛详情页样式

## Related Modules

- `seu-oj-frontend/js/forum.js`
- `seu-oj-frontend/css/forum.css`
- `seu-oj-backend/internal/api/forum.go`
- `seu-oj-backend/internal/service/forum_service.go`

## Suggested Files

- `seu-oj-frontend/js/forum.js`
- `seu-oj-frontend/css/forum.css`
- `seu-oj-backend/internal/api/forum.go`
- `seu-oj-backend/internal/service/forum_service.go`
- `docs/api.md`

## Page Entry

重点检查这些页面：
- `#/forum`
- `#/forum/topics/:id`
- `#/problems/:id` 中的 Discuss 入口
- `#/contests/:id` 中的 Discuss 入口

## Expected Behavior

完成后，论坛首页至少应满足：
- 用户可通过关键字快速筛选帖子
- 用户可明显区分：
  - General discussion
  - Problem discussion
  - Contest discussion
- 帖子卡片元信息比当前更完整
- 从题目页或比赛页进入论坛时，相关讨论能更容易定位
- 帖子详情页和回帖功能保持可用

## Acceptance Criteria

- [ ] 论坛列表支持关键字检索或等价的快速过滤
- [ ] 论坛分区筛选入口清晰可用
- [ ] 题目讨论、比赛讨论、综合讨论的帖子能被区分显示
- [ ] 帖子详情加载正常，回帖流程不受影响
- [ ] 前端语法检查和后端构建通过

## Verification

建议至少完成这些验证：
- 在论坛列表页验证关键字搜索
- 验证不同分区筛选结果是否正确
- 从题目页点击 Discuss，确认能正确进入对应讨论上下文
- 从比赛页点击 Discuss，确认能正确进入对应讨论上下文
- 打开一篇帖子并发送回复，确认详情页未被改坏
- 运行：

```powershell
node --check "D:\desk\软件工程\SEUOJ\seu-oj-frontend\js\forum.js"
cd "D:\desk\软件工程\SEUOJ\seu-oj-backend"
$env:GOCACHE="D:\desk\软件工程\SEUOJ\seu-oj-backend\.gocache"
go build ./...
```

## Suggested Frontend Tech

本 story 可引入轻量渲染技术以提升论坛文本展示和信息层次。

推荐技术：
- `markdown-it` 或 `marked`：用于帖子正文、回复内容的统一 Markdown 渲染
- `dayjs`：用于帖子时间、回复时间的统一格式化

技术约束：
- 不重构论坛整体交互模型
- 不引入重型富文本编辑器作为本次默认方案

## Risks / Notes

- 如果采用前端过滤方案，需要在 delivery 中说明其适用范围
- 如果补了后端查询参数，需要同步更新接口文档
- 不要因为筛选功能改动而影响帖子详情页、回帖页和上下文跳转

## Handover Notes

当前论坛模块已经具备基本交互能力，本 story 的重点是把论坛首页从“可用”提升到“便于持续使用”。如果涉及后端接口变更，delivery 中必须明确说明新增了哪些参数和返回字段。



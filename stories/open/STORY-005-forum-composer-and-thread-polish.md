# STORY-005-forum-composer-and-thread-polish

- Status: `todo`
- Priority: `P1`
- Assignee: `unassigned`
- Reviewer: `unassigned`
- Created At: `2026-04-22`
- Target Version: `v0.1`

## Background

在 `STORY-002` 完成后，论坛首页已经具备了基础的搜索、分区和帖子列表能力，但整体交互仍然偏“功能堆叠”，距离常见 OJ 讨论区的使用习惯还有明显差距。

当前主要问题包括：
- 新建讨论表单直接常驻在列表页，占用较大空间，打断浏览节奏
- `Zone ID` 设计不自然，普通用户不会通过输入 ID 来组织讨论
- 分区按钮与搜索区域的布局关系不够稳定，横向对齐较弱，影响观感
- 帖子卡片依赖小按钮进入详情，交互不够直接
- 时间展示过于原始，`ISO` 时间字符串在列表页信息噪音较高
- 帖子预览区仍以纯文本为主，缺少 Markdown 呈现能力
- 帖子预览和详情页缺少点赞 / 收藏等基础互动入口

这些问题不一定影响“能不能用”，但会明显影响：
- 论坛首页的整洁度
- 用户浏览和发帖的顺手程度
- 课堂展示效果
- 论坛是否像一个真正可持续使用的讨论区

## Goal

进一步打磨论坛首页和帖子交互，让论坛更接近常见 OJ 讨论区的使用体验。

本次 story 需要实现的核心目标是：
- 发帖入口更轻量，不再用大表单压住列表页
- 分区与搜索区更整洁、更自然
- 帖子卡片更像“可点击的内容单元”，而不是按钮集合
- 帖子内容与预览具备更好的 Markdown 表达能力
- 点赞 / 收藏等轻互动在预览和详情中都可见、可用

## Scope

本次包含：
- 将论坛首页中的“新建讨论”区域改为悬浮入口
  - 使用悬浮按钮、浮动入口或等价的轻量入口
  - 点击后进入单独的发帖界面、弹层或侧边编辑界面
  - 不再让完整发帖表单常驻在讨论列表页中
- 移除论坛列表页和发帖入口中的 `Zone ID` 输入设计
  - 不再要求用户手动输入 `problem_id / contest_id`
  - 题目讨论 / 比赛讨论应主要依赖上下文入口进入
  - 分区选择应尽量使用清晰的按钮、标签或上下文预填，而不是原始 ID 输入
- 重构论坛首页分区按钮与搜索区的布局
  - 优化横向对齐与留白关系
  - 保证桌面端和移动端都更自然
  - 使四个分区按钮与下方搜索区形成一致的视觉节奏
- 重构帖子卡片交互
  - 移除 `View` 和 `More in Zone` 这类列表级按钮
  - 点击帖子卡片主体任意区域即可进入帖子详情
  - 需要保留的互动控件应避免误触发跳转
- 优化列表页帖子元信息展示
  - 时间统一改为更简洁的日期格式，例如 `YYYY-MM-DD`
  - 列表页不再显示 `last reply`
  - 保留必要但不过载的作者与状态信息
- 支持帖子内容的 Markdown 呈现
  - 帖子详情必须支持 Markdown 渲染
  - 帖子预览区也应支持安全、可控的 Markdown 预览或截断渲染
- 为帖子增加点赞 / 收藏能力
  - 在帖子预览卡片中显示点赞数、收藏数及当前用户状态
  - 在帖子详情页中也显示并支持操作
  - 交互应有明确的激活态 / 未激活态
- 补齐与上述改动相关的前后端接口、返回字段和文档说明

补充建议项：
- 发帖悬浮入口应支持草稿保留或最少限度的误关闭保护
- 帖子卡片应增加明确的 hover / active 反馈，强化“整卡可点击”感知
- 点赞 / 收藏按钮在列表页与详情页的交互语义保持一致

## Out of Scope

本次不包含：
- 引入完整富文本编辑器
- 增加通知中心、消息提醒或站内信
- 设计复杂的热门帖子算法
- 实现楼中楼 / 嵌套回复
- 重做整个论坛后端数据模型以外的其他业务模块

## Related Modules

- `seu-oj-frontend/js/forum.js`
- `seu-oj-frontend/css/forum.css`
- `seu-oj-frontend/js/problems.js`
- `seu-oj-frontend/js/contests.js`
- `seu-oj-backend/internal/api/forum.go`
- `seu-oj-backend/internal/service/forum_service.go`
- `seu-oj-backend/internal/dto/forum.go`
- `seu-oj-backend/internal/model/...`
- `docs/api.md`

## Suggested Files

- `seu-oj-frontend/js/forum.js`
- `seu-oj-frontend/css/forum.css`
- `seu-oj-frontend/js/problems.js`
- `seu-oj-frontend/js/contests.js`
- `seu-oj-backend/internal/api/forum.go`
- `seu-oj-backend/internal/service/forum_service.go`
- `seu-oj-backend/internal/dto/forum.go`
- 如需新增点赞 / 收藏表，再补充对应 `model`、迁移和 SQL 文件
- `docs/api.md`

## Page Entry

重点检查这些页面：
- `#/forum`
- `#/forum/topics/:id`
- `#/problems/:id` 中的 Discuss 入口
- `#/contests/:id` 中的 Discuss 入口

## Expected Behavior

完成后，论坛页至少应满足：
- 列表页默认专注于“浏览帖子”，而不是一进来就看到大块发帖表单
- 发帖入口轻量、明确，点击后才进入编辑界面
- 用户无需理解 `Zone ID` 也能完成分区浏览和发帖
- 帖子卡片点击路径更自然，整卡即可进入详情
- 列表元信息更克制，日期更易读
- 帖子内容和预览具备 Markdown 表达能力
- 点赞 / 收藏在预览与详情中都可见、可操作

## Acceptance Criteria

- [ ] 论坛首页不再常驻完整发帖表单，而是改为悬浮入口或等价轻量入口
- [ ] `Zone ID` 输入从论坛首页筛选区和发帖入口中移除
- [ ] 分区按钮与搜索区在桌面端和移动端的布局更整齐，视觉对齐明显改善
- [ ] 列表页移除 `View` / `More in Zone` 按钮，点击帖子卡片主体即可进入详情
- [ ] 列表页时间改为简洁日期格式，且不再展示 `last reply`
- [ ] 帖子详情支持 Markdown 渲染，列表预览也支持安全、可控的 Markdown 预览
- [ ] 点赞 / 收藏在帖子预览卡片和详情页中都可见、可用，并显示状态与计数
- [ ] 从题目页和比赛页进入论坛时，分区上下文仍然正确
- [ ] 前端语法检查通过；如涉及后端接口或表结构变更，后端构建也通过

## Risks / Notes

- 移除 `Zone ID` 后，题目 / 比赛讨论的上下文绑定必须依赖页面入口或明确的预填逻辑，不能把能力一起删掉
- 帖子预览的 Markdown 渲染必须控制安全性和截断方式，避免破坏列表布局
- 点赞 / 收藏功能如果新增后端表结构，delivery 中必须明确说明新增字段、接口和约束
- 整卡可点击后，按钮类交互需要做好事件隔离，避免点赞 / 收藏时误跳详情

## Handover Notes

这个 story 仍然以论坛模块为核心，但重点从“能搜索、能分区”转向“更像一个自然可用的 OJ 讨论区”。

实现时建议优先级如下：
1. 发帖入口改为悬浮式入口
2. 删除 `Zone ID` 并重构分区 / 搜索区布局
3. 整卡点击与时间信息瘦身
4. Markdown 预览支持
5. 点赞 / 收藏能力

如果点赞 / 收藏需要新增后端接口：
- 需要在 `docs/api.md` 中补充接口说明
- 需要在 delivery 中明确返回字段和交互状态字段

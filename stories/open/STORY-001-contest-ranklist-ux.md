# STORY-001-contest-ranklist-ux

- Status: `done`
- Priority: `P1`
- Assignee: `unassigned`
- Reviewer: `unassigned`
- Created At: `2026-04-14`
- Target Version: `v0.1`

## Background

当前比赛模块已经具备比赛详情、榜单、公告和比赛题目页，核心流程可以运行。但榜单页面仍然偏基础，主要问题包括：
- 题目列较多时，横向阅读成本高
- 排名、用户信息和每题结果之间的层次不够清楚
- 当前登录用户虽然已有一定高亮，但在长榜单中仍不够醒目
- 封榜后的展示提示还不够直观

这个问题不会阻塞比赛功能本身，但会明显影响课堂展示效果，也会降低后续继续扩展榜单的可维护性。

## Goal

提升比赛榜单页面的可读性和可展示性，使其在以下场景下仍然易于使用：
- 比赛题目数量较多
- 参赛人数较多
- 存在封榜状态
- 需要快速定位当前用户成绩

本 story 的目标不是重写榜单逻辑，而是把已有榜单结果展示得更清楚、更稳定。

## Scope

本次包含：
- 优化比赛榜单表格布局，使长表格更易读
- 为榜单首列（排名、用户名、 solved / penalty）增加更清晰的视觉层级
- 强化当前登录用户在榜单中的高亮显示
- 优化每题单元格的状态表现，区分：
  - 已通过
  - 已尝试未通过
  - 封榜后有隐藏提交
- 检查比赛详情页到榜单页的跳转和展示一致性
- 保证榜单在桌面端展示稳定，不出现明显错位

## Out of Scope

本次不包含：
- 修改榜单计算逻辑
- 新增比赛规则或新的比赛类型
- 修改后端比赛数据结构
- 新增榜单导出功能
- 重做比赛详情页整体布局

## Related Modules

- `seu-oj-frontend/js/contests.js`
- `seu-oj-frontend/css/contest.css`
- `seu-oj-frontend/app.js`

## Suggested Files

- `seu-oj-frontend/js/contests.js`
- `seu-oj-frontend/css/contest.css`

## Page Entry

重点检查这些页面：
- `#/contests`
- `#/contests/:id`
- `#/contests/:id` 中的 ranklist 区块

如果榜单单独有跳转入口，也需要一起回归。

## Expected Behavior

完成后，页面至少应满足：
- 用户进入比赛详情页后，能快速识别榜单区域
- 榜单第一列比当前更稳定、更易读
- 当前登录用户在榜单中能够被快速定位
- 题目列较多时，表头和题目单元格不会造成明显阅读混乱
- 封榜状态下，隐藏提交的存在应有明显提示，但不泄露不该显示的结果

## Acceptance Criteria

- [x] 比赛榜单在题目列较多时仍然可正常阅读
- [x] 当前登录用户在榜单中有明显高亮
- [x] 封榜、通过、未通过、尝试中的状态区分清晰
- [x] 不影响比赛详情页、比赛题目页和榜单页的现有跳转
- [x] 前端脚本语法检查通过

交付说明见：[stories/deliveries/STORY-001-contest-ranklist-ux.md](../deliveries/STORY-001-contest-ranklist-ux.md)。

## Verification

建议至少完成这些验证：
- 打开一个已有比赛页面，确认榜单正常显示
- 用普通用户登录，确认当前用户高亮存在
- 检查封榜比赛和未封榜比赛各一场
- 运行：

```powershell
node --check "D:\desk\软件工程\SEUOJ\seu-oj-frontend\js\contests.js"
```

## Risks / Notes

- 长表格样式改动不要影响普通表格页面
- 如果采用 sticky 方案，需要注意当前父容器的 overflow 结构
- 不要为了榜单样式优化去改动后端榜单逻辑

## Handover Notes

当前比赛榜单已经可用，但更偏“功能跑通”。本 story 是典型的前端展示优化任务，适合独立完成。修改完成后，建议在 delivery 中附上改动前后的截图说明。

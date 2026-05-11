# STORY-004 交付说明 by zhujie

- Story: `STORY-004`
- Author: `zhujie`
- Date: `2026-04-26`
- Version: `v1`

## 概要

本次将首页重构为角色感知的 dashboard，突出用户下一步动作，并将继续工作、快捷入口、最近活动和次级摘要重新分层。

## 已完成修改

- 重新组织 `#/` 首页结构，形成顶部身份与下一步动作、角色化 Quick Actions、Continue Work、Recent Activity、次级摘要的层次。
- 为普通用户、教师、管理员和游客提供不同的首页重点入口，没有新增后端接口，也没有调整全局导航结构。
- 将最近一次提交、最近失败原因、最近 Accepted 记录、继续/复用提交动作合并到同一个 Continue Work 主区域。
- 将系统统计、公告、论坛、比赛、题目/教学摘要降级为次级信息，避免抢占主工作流视觉重心。
- 移除了旧首页中偏调试性质的账户快照信息，例如 token 状态和 API Base。

## 修改文件

- `seu-oj-frontend/app.js`
- `seu-oj-frontend/css/base.css`
- `stories/open/STORY-004-home-dashboard-restructure.md`

## 验证方式

```text
1. 检查首页渲染逻辑仍然只使用已有 API，没有新增后端接口依赖。
2. 检查普通用户、教师、管理员和游客在首页 helper 中的分支逻辑。
3. 执行前端脚本语法检查，覆盖本次改动入口脚本和 story 要求检查的 general-pages 脚本。
```

```powershell
node --check seu-oj-frontend/app.js
node --check seu-oj-frontend/js/general-pages.js
```

## 剩余工作

- 建议后续使用真实普通用户、教师、管理员账号分别进入首页做一次浏览器人工验收。

## 风险 / 已知问题

- 本次 dashboard 只复用现有接口，因此教师侧没有新增统计数字，主要以教学快捷入口承载角色差异。
- 公告和论坛摘要是 best-effort 加载；如果对应接口失败，首页仍会渲染，但不会展示这些摘要。

## Reviewer 关注点

本次有意移除了旧首页的 Workflow 表格和偏调试性质的账户快照。这些内容更适合放在导航、个人页、管理页或具体功能页中；首页现在只保留入口和摘要，不再承担详细页职责。

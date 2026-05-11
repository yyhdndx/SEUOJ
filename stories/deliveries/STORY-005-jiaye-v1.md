# STORY-005 delivery by Jiaye

- Story: `STORY-005-forum-composer-and-thread-polish`
- Author: `Jiaye0618`
- Date: `2026-05-02`
- Version: `v1`

## Summary

完成 STORY-005 全部 5 项任务：发帖悬浮入口、Zone ID 移除与布局重构、整卡点击与时间瘦身、Markdown 预览/详情渲染、点赞/收藏功能。前后端均已完成并通过构建验证。

## Completed Changes

- 论坛首页发帖入口从常驻表单改为 FAB 悬浮按钮 + 弹层编辑器，支持误关闭保护
- 移除论坛筛选区和发帖弹层中的 Zone ID 输入，分区选择仅通过按钮/下拉框完成
- 重构分区按钮与搜索区布局为统一的 `.forum-filter-grid`，桌面端和移动端均对齐
- 移除帖子卡片上的 View / More in Zone 按钮，整卡点击进入详情，like/fav 按钮通过 `stopPropagation` 隔离
- 列表页时间统一为 `YYYY-MM-DD` 格式，移除 last reply 展示
- 帖子详情和列表预览均支持 Markdown 渲染（详情用 `renderMarkdownBlock`，预览用 `renderMarkdownPreview` 带安全截断）
- 新增点赞/收藏功能：两张关联表、4 个 API 端点、列表卡片和详情页按钮、optimistic UI 即时反馈
- 新增 `OptionalJWTAuth` 中间件，使公开论坛接口可识别已登录用户并水合 `is_liked`/`is_favorited` 状态
- 题目页和比赛页的 Discuss 按钮正确传递 scope_type/scope_id 上下文到论坛

## Files Changed

### 后端 (9 files)

| 文件 | 操作 | 说明 |
|------|------|------|
| `seu-oj-backend/internal/model/forum_topic_like.go` | 新建 | 点赞关联表模型，`(topic_id, user_id)` 唯一索引 |
| `seu-oj-backend/internal/model/forum_topic_favorite.go` | 新建 | 收藏关联表模型，`(topic_id, user_id)` 唯一索引 |
| `seu-oj-backend/internal/model/forum_topic.go` | 修改 | 新增 `LikeCount`、`FavoriteCount` 反范式计数字段 |
| `seu-oj-backend/internal/database/db.go` | 修改 | AutoMigrate 注册新模型 + ALTER TABLE 兼容旧库 |
| `seu-oj-backend/internal/dto/forum.go` | 修改 | DTO 新增 `LikeCount`、`FavoriteCount`、`IsLiked`、`IsFavorited` 字段 |
| `seu-oj-backend/internal/service/forum_service.go` | 修改 | 新增 Like/Unlike/Favorite/Unfavorite 方法、hydration 逻辑、`invalidate` 缓存清除 |
| `seu-oj-backend/internal/api/forum.go` | 修改 | 新增 4 个 handler；`ListTopics`/`TopicDetail` 改用 `getOptionalUserID` |
| `seu-oj-backend/internal/api/submission.go` | 修改 | 新增 `getOptionalUserID` 辅助函数（不写 HTTP response） |
| `seu-oj-backend/internal/middleware/jwt.go` | 修改 | 新增 `OptionalJWTAuth` 中间件 |
| `seu-oj-backend/internal/router/router.go` | 修改 | 注册 4 条 auth 路由；2 条 public 路由应用 OptionalJWTAuth |

### 前端 (4 files)

| 文件 | 操作 | 说明 |
|------|------|------|
| `seu-oj-frontend/js/forum.js` | 修改 | FAB 弹层编辑器、整卡点击、Markdown 预览/渲染、点赞/收藏按钮与交互、日期格式化 |
| `seu-oj-frontend/css/forum.css` | 修改 | FAB 样式、卡片 hover 态、filter-grid 布局、like/fav 按钮激活态样式 |
| `seu-oj-frontend/js/problems.js` | 修改 | Discuss 入口传递 scope 上下文 |
| `seu-oj-frontend/js/contests.js` | 修改 | Discuss 入口传递 scope 上下文 |

## Verification

```text
1. 启动后端: cd seu-oj-backend && go run .
2. 打开 http://127.0.0.1:8080/#/forum
3. 登录后验证:
   - 论坛首页右下角显示 + 号 FAB 按钮，点击弹出 New Topic 弹层
   - 弹层中无 Zone ID 输入，仅 Title / Zone (下拉) / Content (Markdown)
   - 弹层按 Escape 或点击遮罩关闭，有内容时提示确认丢弃
   - 分区按钮 All / General / Problem / Contest 布局整齐
   - 搜索区 keyword + zone 下拉 + Search / Reset 横向对齐
   - 帖子整卡可点击进入详情，hover 有视觉反馈
   - 日期显示为 YYYY-MM-DD 格式，无 last reply
   - 列表卡底部显示 ♥ N / ★ N 点赞收藏按钮
   - 未登录时按钮无 is-active 态，点击触发登录提示
   - 登录后点击 ♥ → 变红计数+1，再次点击 → 恢复灰色计数-1
   - ★ 按钮同理（金色激活态）
   - 进入详情页，sidebar 显示 Likes/Favorites 计数
   - header 显示 Like/Save 按钮，可正常切换
   - 详情页 Markdown 内容正确渲染
4. 从题目页 Discuss 和比赛页 Discuss 进入论坛，分区上下文正确
5. 后端构建: go build ./... (通过)
6. 前端语法检查: node --check js/forum.js (通过)
```

```powershell
cd seu-oj-backend && go build ./...
node --check ../seu-oj-frontend/js/forum.js
```

## Remaining Work

- `docs/api.md` 中补充点赞/收藏接口说明（新增 4 个端点）
- 可考虑详情页 like/fav 操作后局部更新而非全量 re-render

## Risks / Known Issues

- 点赞/收藏计数器采用反范式化存储（`like_count`/`favorite_count` 在 `forum_topics` 表），极端并发下计数可能与实际行数微小偏差；已有 `GREATEST(x - 1, 0)` 保护防止负数
- `OptionalJWTAuth` 不阻断未登录请求，未登录用户仍可浏览论坛但无法点赞/收藏
- 缓存 TTL 为 20s（列表）/ 30s（详情），点赞后需等待缓存过期或二次操作才能看到其他人操作的计数变化

## Notes For Reviewer

- 重点注意 `OptionalJWTAuth` 的设计：不写 HTTP response、不 Abort，仅在 token 有效时设置 context。这解决了原 `getContextUserID` 写双份 JSON 导致 "invalid response: 200" 的 bug
- `getOptionalUserID` 放在 `submission.go` 但实际被 `forum.go` 引用，因为两者在同一 package，无需重复定义
- 前端 like/fav 采用 optimistic UI：先切换 DOM 再发 API，失败由 catch 的 `setFlash` 提示
- `forum.js` 中 `toggleForumReaction` 是列表和详情共用的核心函数，详情页操作后调用 `renderForumTopicDetail(id)` 全量刷新以保证 sidebar 和 header 计数一致

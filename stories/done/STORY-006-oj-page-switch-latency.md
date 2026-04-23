# STORY-006-oj-page-switch-latency

- Status: `done`
- Priority: `P1`
- Assignee: `codex`
- Reviewer: `unassigned`
- Created At: `2026-04-23`
- Target Version: `v0.x`

## Background

OJ 在从其他页面切换回首页或常用列表页时，存在接口请求耗时较高的问题。初始加载耗时可以接受，但缓存热后仍出现部分接口超过 100ms 到 700ms 的延迟，影响页面切换体验。

日志显示主要慢点包括：
- `/api/stats/me` 冷加载可达到 600ms 以上
- `/api/contests` 在短时间内仍可能出现 200ms 以上请求
- `/api/submissions/my` 在 60 秒内仍可能 miss，原因是不同页面使用了不同 `page_size`
- `/api/submissions/32` 热缓存后已基本稳定在 0ms 到 1ms

## Goal

降低 OJ 页面切换时的后端接口延迟，重点优化首页相关接口在缓存热状态下的响应时间，并提供后端可观测性用于定位后续慢请求。

## Scope

本次包含：
- 收敛数据库索引，保留对实际查询路径有帮助的组合索引，减少低价值单列索引
- 引入 Redis JSON 缓存层，支持通用 `GetOrSet` 与缓存失效
- 为后端请求增加 timing middleware 和 `Server-Timing` 记录
- 缓存首页、题目、比赛、统计、榜单、公告、论坛、提交记录等读接口
- 优化 `/api/submissions/my` 和 `/api/submissions/{id}` 的热路径
- 将首页与列表页常见的不同 `page_size` 请求收敛到共享缓存
- 优化比赛列表 N+1 查询
- 优化用户统计 `/api/stats/me` 的多次 count 查询
- 在提交创建、重判、判题状态更新后失效相关缓存

## Out of Scope

本次不包含：
- 前端 timing 调试代码，最终已从 `seu-oj-frontend/app.js` 回退
- 改造前端页面结构或交互样式
- 引入新的业务接口协议
- 对管理端重查询路径做完整缓存覆盖
- 对所有冷启动慢查询做彻底 SQL 重写

## Related Modules

- `seu-oj-backend/internal/cache`
- `seu-oj-backend/internal/middleware`
- `seu-oj-backend/internal/observability`
- `seu-oj-backend/internal/service`
- `seu-oj-backend/internal/repository`
- `seu-oj-backend/internal/api`
- `seu-oj-backend/internal/router`
- `seu-oj-backend/internal/judge`
- `seu-oj-backend/internal/model`
- `seu-oj-backend/database`

## Suggested Files

- `seu-oj-backend/internal/cache/cache.go`
- `seu-oj-backend/internal/middleware/timing.go`
- `seu-oj-backend/internal/observability/timing.go`
- `seu-oj-backend/internal/router/router.go`
- `seu-oj-backend/internal/service/submission_service.go`
- `seu-oj-backend/internal/service/problem_service.go`
- `seu-oj-backend/internal/service/contest_service.go`
- `seu-oj-backend/internal/service/stats_service.go`
- `seu-oj-backend/internal/service/ranklist_service.go`
- `seu-oj-backend/internal/service/forum_service.go`
- `seu-oj-backend/internal/service/announcement_service.go`
- `seu-oj-backend/internal/repository/submission_repository.go`
- `seu-oj-backend/internal/judge/worker.go`
- `seu-oj-backend/database/index.sql`

## Completed Changes

- 新增后端 timing middleware，统一输出请求耗时和 `Server-Timing`。
- 新增 Redis 缓存工具，封装 cache get、DB load、cache set 的计时。
- 将首页关键接口缓存 TTL 收敛到 60 秒，避免 15 秒到 20 秒短 TTL 导致的频繁 miss。
- 将 `/api/problems` 第一页无筛选查询统一到 `recent:limit100` 缓存，再按请求 `page_size` 切片返回。
- 将 `/api/contests` 第一页无筛选查询统一到 `recent:limit50` 缓存，再按请求 `page_size` 切片返回。
- 将 `/api/submissions/my` 第一页无筛选查询统一到 `recent:limit100` 缓存，解决 home 的 `page_size=5` 和题目页的 `page_size=100` 互相 miss。
- 将 `/api/submissions/{id}` 详情缓存 60 秒，并在权限校验前从缓存加载详情。
- `/api/stats/me` 从多次 count 查询收敛为一次聚合查询，降低冷 miss 成本。
- 比赛列表批量统计报名人数和题目数量，移除 per-contest count。
- 比赛题目加载批量查询题目 brief，移除 N+1 题目读取。
- 判题 worker 在提交状态变化后失效提交详情、提交列表、统计、榜单和题目统计缓存。
- 索引脚本和 GORM model index tag 收敛为实际查询路径需要的组合索引。
- 前端调试 timing 代码已移除，最终改动局限在后端。

## Acceptance Criteria

- [x] 首页缓存热后，`/api/problems`、`/api/contests`、`/api/stats/overview`、`/api/stats/me` 应优先命中缓存。
- [x] 60 秒内从首页切换到题目列表或比赛列表，不应因为 `page_size` 不同导致常见第一页无筛选查询重新打 DB。
- [x] `/api/submissions/my?page=1&page_size=5` 和 `/api/submissions/my?page=1&page_size=100` 共享最近提交缓存。
- [x] `/api/submissions/{id}` 热缓存后应接近 0ms 到 1ms。
- [x] 后端日志能通过 `server_timing` 区分 `cache_get`、`db`、`cache_set` 和 `total`。
- [x] 提交创建、重判、判题状态变化后，相关缓存能被失效。
- [x] 前端不保留 timing console 日志和额外 request timing 逻辑。
- [x] 后端测试通过。

## Verification

已执行：

```powershell
cd seu-oj-backend
go test ./...
```

结果：
- `go test ./...` 通过
- `git diff --check` 通过
- `seu-oj-frontend/app.js` 已恢复到当前分支基线，前端 timing 相关 diff 为空

日志验证结论：
- `/api/submissions/32` 热缓存后基本为 `0ms`
- `/api/submissions/my` 热缓存后基本为 `0ms` 到 `3ms`
- 原主要剩余慢点集中在 `/api/stats/me`、`/api/contests` 和不同 `page_size` 造成的缓存键不一致，本次已针对这些路径收敛

## Risks / Notes

- 缓存 TTL 目前偏向页面切换体验，部分统计数据允许最多约 60 秒延迟。
- 缓存失效使用 prefix 删除，数据量很大时需要关注 Redis `SCAN + DEL` 的成本。
- `Server-Timing` 和后端 timing 日志建议保留，后续可通过配置控制详细日志输出。
- 首页冷启动仍可能较慢，本 story 的目标是接受初始加载成本，优化缓存热后的页面切换体验。

## Handover Notes

后续如果再次出现页面切换慢，优先查看日志中的 `server_timing`：
- `cache_get;desc="hit"` 但 `total` 高，优先检查网络或响应序列化
- `cache_get;desc="miss"` 且 `db` 高，优先检查 TTL、失效策略或缓存键是否过细
- 同一业务数据如果被不同页面用不同 query 请求，应优先考虑共享基础缓存后切片返回

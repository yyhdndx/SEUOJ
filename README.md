# SEU OJ

SEU OJ 是一个面向课程教学、日常训练与竞赛组织的轻量级 Online Judge 系统。

当前仓库包含：
- `seu-oj-backend`：Go 后端，提供 API、判题调度、静态资源托管
- `seu-oj-frontend`：静态前端 SPA，按业务域拆分 JS/CSS
- `docs`：接口文档、需求建模、答辩材料与协作文档

## 当前能力概览

系统当前已经具备这些核心模块：
- 用户注册、登录、JWT 鉴权、角色管理
- 题库、题解、题目统计、管理员题目 CRUD
- 多语言提交与 Docker 沙箱判题
- 比赛、报名、榜单、封榜、赛后练习、比赛公告
- 题单、班级、作业/考试、教师管理页
- 论坛、公告、排行榜、提交管理

详细功能说明见：
- [当前功能总览](./docs/current-features.md)
- [后端接口文档](./docs/api.md)
- [需求建模与用例图](./docs/requirements-modeling.md)
- [PPT 需求部分建议](./docs/ppt-requirements.md)

## 运行方式

推荐先看：
- [展示前检查清单](./docs/demo-checklist.md)
- [后端接口文档](./docs/api.md)

当前项目运行时通常需要：
- MySQL
- Redis
- Docker
- Web 服务：`go run .`
- Judge Worker：`go run ./cmd/judge-worker`

后端目录：
```powershell
cd "D:\desk\软件工程\SEUOJ\seu-oj-backend"
go run .
```

Worker 目录：
```powershell
cd "D:\desk\软件工程\SEUOJ\seu-oj-backend"
go run ./cmd/judge-worker
```

浏览器访问：
```text
http://127.0.0.1:8080/
```

## 文档索引

- [docs/api.md](./docs/api.md)：后端接口文档
- [docs/current-features.md](./docs/current-features.md)：当前实现功能总览
- [docs/demo-checklist.md](./docs/demo-checklist.md)：展示前功能检查清单
- [docs/requirements-modeling.md](./docs/requirements-modeling.md)：需求建模说明
- [docs/usecase-overall.puml](./docs/usecase-overall.puml)：总体用例图
- [docs/usecase-teaching-split.puml](./docs/usecase-teaching-split.puml)：教学模块用例图

## 协作建议

前端已经按业务域拆分：
- `js/problems.js`
- `js/submissions.js`
- `js/contests.js`
- `js/teaching.js`
- `js/admin.js`
- `js/forum.js`

样式也已拆分：
- `css/base.css`
- `css/problem.css`
- `css/submission.css`
- `css/contest.css`
- `css/teaching.css`
- `css/forum.css`

建议组员按业务域并行维护，而不是多人同时修改同一个入口文件。

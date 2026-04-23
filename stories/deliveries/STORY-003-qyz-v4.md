# STORY-XXX delivery by NAME

- Story: `STORY-003`
- Author: `qyz`
- Date: `YYYY-MM-DD`
- Version: `v4`

## Completed Changes

- 向题目表中增加了难度字段(seu-oj-backend/database/problems.sql)，并修改了对应的前/后端处理逻辑
- 固定了题目页面中的页签
- 修改了题目列表界面的题目信息

## Files Changed

- `seu-oj-backend/database/problems.sql`
- `seu-oj-backend/internal/dto/problem.go`
- `seu-oj-backend/internal/model/problem.go`
- `seu-oj-backend/internal/service/problem_service.go`
- `seu-oj-frontend/app.js`
- `seu-oj-frontend/css/base.css`
- `seu-oj-frontend/css/problem.css`
- `seu-oj-frontend/js/admin.js`
- `seu-oj-frontend/js/contests.js`
- `seu-oj-frontend/js/problems.js`
- `seu-oj-frontend/js/submissions.js`

## Verification

首先安装CodeMirror所需依赖

```text
1. cd ./seu-oj-backend
2. go run .
```

## Remaining Work

- Solutions功能未实现
- 代码缓存不会随语言改变，当前仅能缓存一份代码
- 当代码为空时直接Run/Submit会出现报错框，可消去但很丑

## Risks / Known Issues

## Notes For Reviewer

None

# STORY-XXX delivery by NAME

- Story: `STORY-003`
- Author: `qyz`
- Date: `YYYY-MM-DD`
- Version: `v4`

## Completed Changes

- 数据库中添加了难度字段，修改了后端对难度字段的渲染逻辑
- 修改了题目列表界面，删去了自增id列和创建时间，添加了难度字段
- 修复了题目页面{题目详情、题解、提交记录}页签滚动的问题
- 优化了代码为空时运行/提交的提示页面
- 更改了代码缓存逻辑，现在不同语言的代码分开缓存，切换语言时对应切换

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


```bash
cd ./seu-oj-backend
go run .
```

## Remaining Work

- Solutions功能未实现
- 提交界面可以优化

## Risks / Known Issues=

## Notes For Reviewer

None

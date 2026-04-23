# STORY-XXX delivery by NAME

- Story: `STORY-003`
- Author: `qyz`
- Date: `YYYY-MM-DD`
- Version: `v3`

## Completed Changes

- 将代码编辑区改为Code Mirror
- 去掉了详情、题解页签处的滚动条

## Files Changed

- `seu-oj-frontend/CodeMirror/package-lock.json`
- `seu-oj-frontend/CodeMirror/package.json`
- `seu-oj-frontend/css/problem.css`
- `seu-oj-frontend/index.html`
- `seu-oj-frontend/js/admin.js`
- `seu-oj-frontend/js/codemirror-loader.js`
- `seu-oj-frontend/js/contests.js`
- `seu-oj-frontend/js/general-pages.js`
- `seu-oj-frontend/js/problems.js`

## Verification

首先安装CodeMirror所需依赖

```bash
cd ./seu-oj-frontend/CodeMirror
npm install
```

检查前端结果
```text
1. cd ./seu-oj-backend
2. go run .
3. 点击Run 按钮
```

## Remaining Work

- Solutions功能未实现
- 代码缓存不会随语言改变，当前仅能缓存一份代码
- 当代码为空时直接Run/Submit会出现报错框且无法消去

## Risks / Known Issues

- 当前Run使用本地环境

## Notes For Reviewer

None

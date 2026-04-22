# STORY-001 交付说明：比赛详情页 / 榜单 / 题目区 UX

- Story：`stories/open/STORY-001-contest-ranklist-ux.md`
- 交付日期：2026-04-21
- 涉及前端：`seu-oj-frontend/js/contests.js`、`seu-oj-frontend/css/contest.css`

## 1. 需求理解与目标

在**不修改后端榜单计算逻辑**、**不改动榜单 API 数据结构**的前提下，提升比赛详情页在「公告 / 题单 / 我的提交 / 榜单」区域的可用性与课堂展示效果，并强化封榜提示与当前用户定位能力。

## 2. 方案概要

### 2.1 比赛详情主内容区改为标签页

- 将原先纵向堆叠的区块合并为**同一卡片内**的顶部标签栏 + 下方面板切换；**标签栏紧跟页面标题行下方**（标题行仍单独展示比赛名与操作按钮）。
- 标签包括：**Contest Overview**（原 Description + Overview 两栏）、Contest Announcements、Problem Set、My Contest Submissions、Standings。
- **默认标签**：Contest Overview（无 `?tab=` 时）。
- **深链**：支持 `#/contests/:id?tab=overview|announcements|problems|submissions|standings`；进入题单使用 `#/contests/:id?tab=problems`。
- **标签切换与地址栏**：使用 `history.replaceState` 更新 `location.hash` 中的 `?tab=`，避免每次点标签触发 `hashchange` 导致整页重新拉取接口；定时轮询仍会完整刷新页面数据。

### 2.2 题单：行式 / 卡片式与整行进入题目

- **行式表格**：去掉单独的「Solve」列；表格行 `tr.contest-problem-click-row` 可点击（及键盘 Enter/Space）跳转到 `#/contests/:id/problems/:problem_id`。
- **卡片式**：响应式网格，每张卡片展示题号、标题、判题模式与时空限制；整张卡片为链接。
- **默认布局策略**：题目数量 **≥ 6** 时，若用户未在 `localStorage` 中保存过偏好，则默认 **卡片式**；否则默认行式。用户可通过 **Row layout / Card layout** 切换，偏好键：`seuoj_contest_problem_layout_<contestId>`。
- **管理员比赛详情页**的题单区复用同一套组件与交互，并在渲染后调用相同的布局初始化逻辑。

### 2.3 「我的比赛提交」行可点

- 最近提交表格每行增加 `contest-submission-click-row`，整行跳转到 `#/submissions/:id`（不再依赖第一列链接）。

### 2.4 榜单表格与分页

- **列结构**（与 ACM 规则及现有字段对齐）：
  - **Rank**：名次。
  - **Participant**：用户名 + 学号/账号行（`userid`）。
  - **Total**：上行 **解题数**（`solved_count`），下行 **总罚时说明**（总 penalty 分钟，含规则内罚时）。
  - **每题一列**：上行沿用原 `contestRankCellLabel` 语义（`+` / `+k` / `-k` / `?` / `–` 及封榜 `*`）；下行在 **Accepted** 时显示该题 **penalty_minutes**（分钟），否则为 `–`。
- **分页**：客户端按固定页长 **25** 条切页；右下角分页条为「共 N 页」+ `« < 1 2 … 10 11 > »` 风格，中间页码采用经典省略号算法（`getNumericPaginationRange`）。
- **当前用户摘要条**：在榜单上方展示「Your standing」（未登录或未上榜则提示文案）；**按钮**在点击后自动翻到用户所在页、`scrollIntoView` 并短时高亮对应行（`contest-rank-row-flash`）。

### 2.5 封榜与用户名颜色

- **封榜显著提示**：在榜单区域顶部增加 **FROZEN** 横幅（`contest-freeze-banner`），并保留简短脚注说明 `*` 含义。
- **用户名颜色**（全表一致，表达「封榜前 / 封榜后」氛围，而非逐人状态）：
  - 未封榜（`ranklist_frozen === false`）：`.contest-ranklist-live .contest-rank-username` 为**绿色系**。
  - 已封榜：`.contest-ranklist-frozen .contest-rank-username` 为**红色系**。

## 3. 实现与文件清单

| 文件 | 变更说明 |
|------|-----------|
| `seu-oj-frontend/js/contests.js` | 比赛详情标签页、题单双布局、提交行点击、榜单新表格/分页/摘要/封榜条；管理员页题单与榜单接入；常量与分页/URL 辅助函数。 |
| `seu-oj-frontend/css/contest.css` | 标签、题单卡片网格、可点行、封榜横幅、榜单双层单元格、分页条、用户名颜色、高亮动画等样式。 |
| `README.md` | 增加本交付文档索引链接。 |

**未改动**：后端 `ContestRanklistResponse` 及榜单计算逻辑；其他业务域脚本。

## 4. 验收与自测建议

1. 打开 `#/contests/:id`，确认默认显示 **Contest Overview**，标签切换正常。
2. 访问 `#/contests/:id?tab=standings`，确认直接进入榜单标签。
3. 有权限时打开 **Problem Set**：行式整行可进题；卡片式多列排布；切换布局后刷新页面偏好仍保留。
4. **My Contest Submissions**：点击任意数据行进入提交详情。
5. 榜单：总列与每题列均为两行信息；分页在人数较多时出现且 `«` `»` 边界正确；登录用户点击摘要可定位到本人行。
6. 封榜比赛：出现 FROZEN 横幅，用户名为红色系；未封榜为绿色系。
7. 语法检查：`node --check seu-oj-frontend/js/contests.js`（已通过）。

## 5. 已知限制与后续可做

- 分页为**纯前端**切片，极大规模榜单时首次仍会一次性传输完整列表（与既有 API 行为一致）。若将来需要服务端分页，需另开 story 并改 API。
- 标签切换使用 `replaceState`，浏览器历史中的「上一页」对 `?tab=` 的恢复行为与 `pushState` 方案不同；若产品明确要求「返回即回到上一标签」，可再评估改为 `pushState` 并接 `popstate`。

## 6. Story 状态

建议在评审通过后：

- 将 `stories/open/STORY-001-contest-ranklist-ux.md` 中 Acceptance Criteria 勾选为完成，并将 Status 更新为 `done`（或按组内流程移至 `stories/done/`）。

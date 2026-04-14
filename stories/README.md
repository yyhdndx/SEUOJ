# Story 协作规则

本目录用于项目组内部的轻量协作管理，替代外部共享文档，统一记录待开发事项和交付说明。

## 目标

- 明确每个开发事项的边界与验收标准
- 记录每次交付的实际修改内容与验证方式
- 便于项目负责人分配任务、验收结果和归档

## 目录结构

```text
stories/
  README.md
  open/
  done/
  deliveries/
  templates/
    story-template.md
    delivery-template.md
```

目录说明：
- `open/`：待开发或开发中的 story
- `done/`：已验收完成的 story
- `deliveries/`：每次交付说明，按 story 编号归档
- `templates/`：story 和 delivery 模板

## 角色分工

### 项目负责人
职责：
- 创建 story
- 定义开发范围与验收标准
- 指定相关模块或参考文件
- 审核交付结果
- 验收完成后归档 story

### 组员
职责：
- 领取 `open/` 中的 story
- 按 story 说明完成开发
- 提交 delivery 文件
- 在 delivery 中说明改动范围、验证方式、剩余问题和风险

## 命名规则

### Story 文件命名
文件位置：`stories/open/` 或 `stories/done/`

示例：
```text
STORY-001-teacher-dashboard.md
STORY-002-contest-ranklist-polish.md
```

命名要求：
- 前缀固定为 `STORY-编号`
- 编号使用三位数，从 `001` 开始递增
- 文件名后半部分使用简短英文短语描述主题
- 文件名仅使用小写英文、数字和连字符

### Delivery 文件命名
文件位置：`stories/deliveries/`

示例：
```text
STORY-001-zhangsan-v1.md
STORY-001-lisi-v2.md
```

命名要求：
- 必须包含对应的 `STORY-编号`
- 必须包含提交人姓名或缩写
- 同一提交人的重复补充使用 `v1 / v2 / ...` 递增

## Story 编写规则

每个 story 必须包含：
- 背景
- 目标
- 范围
- 不在本次范围内的内容
- 相关模块
- 建议修改文件
- 验收标准
- 风险或注意事项
- 交接说明

编写要求：
- 一个 story 只处理一个明确的问题
- 不将多个不相关的大功能混在同一个 story 中
- 验收标准必须可判断、可检查

### 合理的 story 示例
- 教师作业页增加学生搜索和进度筛选
- 比赛榜单页增加 sticky header
- 论坛帖子列表支持按分区过滤

### 不合适的 story 示例
- 完善教学模块
- 优化前端
- 修所有 bug

## Delivery 编写规则

组员完成 story 后，必须新增一份 delivery 文件。

delivery 文件必须说明：
- 对应的 story
- 提交人
- 实际完成的修改内容
- 涉及文件
- 验证方式
- 剩余工作
- 风险或已知问题
- 给 reviewer 的说明

编写要求：
- 不使用“已完成”这类空泛表述
- 必须给出实际修改文件路径和验证方法
- 如果只完成部分内容，必须明确说明未完成项

## 推荐流程

### 1. 创建 story
项目负责人在 `stories/open/` 下新增 story 文件。

### 2. 领取 story
组员开始开发前，将 story 中的 `Assignee` 更新为本人，并将状态改为 `in_progress`。

### 3. 完成开发
组员在代码中完成对应修改。

### 4. 提交 delivery
组员在 `stories/deliveries/` 下新增 delivery 文件，记录本次交付。

### 5. 验收
项目负责人检查：
- 功能是否达到验收标准
- delivery 是否完整清晰
- 是否影响现有模块

### 6. 归档
验收通过后：
- 将 story 从 `open/` 移动到 `done/`
- 将状态更新为 `done`

## 状态约定

story 状态统一使用：
- `todo`
- `in_progress`
- `review`
- `done`
- `blocked`

不再额外扩展其他状态。

## 与 Git 的关系

story 机制用于记录任务和交付，不替代版本控制。

实际代码变更仍以以下内容为准：
- Git 提交
- 代码 diff
- 构建与验证结果

## 粒度建议

建议负责人优先拆分出 0.5 天到 1 天可完成的 story。

推荐原则：
- 单个 story 不跨太多模块
- 单个 story 的验收标准应可快速验证
- story 数量应足以支持并行开发，但避免过细导致维护成本过高

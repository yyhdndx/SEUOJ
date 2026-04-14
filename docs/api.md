# SEU OJ Backend API Documentation

本文档基于当前代码仓库中的后端路由与 DTO 整理，目标是提供一份适合开发联调和组内协作使用的接口文档。

建议配合阅读：
- [当前功能总览](./current-features.md)
- [展示前检查清单](./demo-checklist.md)
- [需求建模说明](./requirements-modeling.md)

当前接口文档覆盖这些模块：
- 认证与用户
- 题库、题解、提交、判题
- 比赛、榜单、公告
- 教学模块：题单、班级、作业、教师接口
- 论坛与管理接口

- 后端基地址：`http://127.0.0.1:8080`
- API 前缀：`/api`
- 统一返回格式：

```json
{
  "code": 0,
  "message": "ok",
  "data": {}
}
```

- 失败返回示例：

```json
{
  "code": 1,
  "message": "invalid request parameters",
  "data": null
}
```

## 1. 鉴权说明

需要登录的接口必须携带：

```http
Authorization: Bearer <token>
```

角色约定：
- `student`：普通用户/学生
- `teacher`：教师
- `admin`：管理员

权限中间件：
- `JWTAuth`：需要登录
- `RequireAdmin`：仅管理员
- `RequireTeacherOrAdmin`：教师或管理员

## 2. 认证模块 `/api/auth`

### 2.1 注册
- 方法：`POST`
- 路径：`/api/auth/register`
- 权限：公开

请求体：

```json
{
  "username": "alice",
  "userid": "213220001",
  "password": "123456"
}
```

字段约束：
- `username`：3~20
- `userid`：3~50
- `password`：至少 6 位

### 2.2 登录
- 方法：`POST`
- 路径：`/api/auth/login`
- 权限：公开

请求体：

```json
{
  "username": "alice",
  "password": "123456"
}
```

成功响应中的 `data`：

```json
{
  "token": "jwt-token",
  "user": {
    "id": 1,
    "username": "alice",
    "userid": "213220001",
    "role": "student",
    "status": "active"
  }
}
```

### 2.3 当前用户信息
- 方法：`GET`
- 路径：`/api/auth/me`
- 权限：登录

### 2.4 更新个人资料
- 方法：`PUT`
- 路径：`/api/auth/profile`
- 权限：登录

请求体：

```json
{
  "username": "alice",
  "userid": "213220001"
}
```

### 2.5 修改密码
- 方法：`PUT`
- 路径：`/api/auth/password`
- 权限：登录

请求体：

```json
{
  "current_password": "123456",
  "new_password": "654321"
}
```

## 3. 健康检查与统计

### 3.1 健康检查
- 方法：`GET`
- 路径：`/api/health`
- 权限：公开

### 3.2 全站概览统计
- 方法：`GET`
- 路径：`/api/stats/overview`
- 权限：公开

### 3.3 当前用户统计
- 方法：`GET`
- 路径：`/api/stats/me`
- 权限：登录

### 3.4 管理员统计
- 方法：`GET`
- 路径：`/api/stats/admin`
- 权限：管理员

### 3.5 总榜
- 方法：`GET`
- 路径：`/api/ranklist`
- 权限：公开

## 4. 公告模块 `/api/announcements`

### 4.1 公告列表
- 方法：`GET`
- 路径：`/api/announcements`
- 权限：公开
- 查询参数：
  - `page`
  - `page_size`

### 4.2 公告详情
- 方法：`GET`
- 路径：`/api/announcements/:id`
- 权限：公开

## 5. 题目模块 `/api/problems`

### 5.1 题目列表
- 方法：`GET`
- 路径：`/api/problems`
- 权限：公开
- 查询参数：
  - `page`
  - `page_size`
  - `keyword`

响应中的列表项字段：
- `id`
- `display_id`
- `title`
- `judge_mode`
- `time_limit_ms`
- `memory_limit_mb`
- `visible`
- `created_at`

### 5.2 题目详情
- 方法：`GET`
- 路径：`/api/problems/:id`
- 权限：公开

响应包含：
- 题面信息
- 样例信息
- `sample` 类型测试点
- 公开题解 `solutions`

### 5.3 题目统计
- 方法：`GET`
- 路径：`/api/problems/:id/stats`
- 权限：公开

### 5.4 题解列表
- 方法：`GET`
- 路径：`/api/problems/:id/solutions`
- 权限：公开

## 6. 提交模块 `/api/submissions`

### 6.1 运行样例
- 方法：`POST`
- 路径：`/api/submissions/run`
- 权限：登录

请求体：

```json
{
  "problem_id": 3,
  "contest_id": 1,
  "language": "cpp",
  "code": "#include <iostream>\nusing namespace std;\nint main(){int a,b;cin>>a>>b;cout<<a+b;return 0;}"
}
```

说明：
- `contest_id` 可省略
- `language` 支持：`cpp`、`c`、`java`、`python3`、`go`、`rust`
- 该接口只运行样例，不入队，不落库

### 6.2 创建提交
- 方法：`POST`
- 路径：`/api/submissions`
- 权限：登录

请求体与 `run` 相同。

成功响应：

```json
{
  "submission_id": 101,
  "status": "Pending"
}
```

### 6.3 我的提交列表
- 方法：`GET`
- 路径：`/api/submissions/my`
- 权限：登录
- 查询参数：
  - `page`
  - `page_size`
  - `problem_id`
  - `contest_id`
  - `status`

### 6.4 提交详情
- 方法：`GET`
- 路径：`/api/submissions/:id`
- 权限：登录

说明：
- 学生只能看自己的提交
- 管理员可看任意提交

响应包含：
- `code`
- `status`
- `compile_info`
- `error_message`
- testcase 结果 `results`
- `contest_id`
- `is_practice`

## 7. 比赛模块 `/api/contests`

### 7.1 比赛列表
- 方法：`GET`
- 路径：`/api/contests`
- 权限：公开
- 查询参数：
  - `page`
  - `page_size`
  - `keyword`
  - `status`：`upcoming | running | ended`

### 7.2 比赛详情
- 方法：`GET`
- 路径：`/api/contests/:id`
- 权限：公开

### 7.3 比赛榜单
- 方法：`GET`
- 路径：`/api/contests/:id/ranklist`
- 权限：公开

### 7.4 比赛公告列表
- 方法：`GET`
- 路径：`/api/contests/:id/announcements`
- 权限：公开

### 7.5 比赛公告详情
- 方法：`GET`
- 路径：`/api/contests/:id/announcements/:announcement_id`
- 权限：公开

### 7.6 当前用户在比赛中的状态
- 方法：`GET`
- 路径：`/api/contests/:id/me`
- 权限：登录

返回字段包括：
- `registered`
- `contest_status`
- `can_register`
- `can_view_problems`
- `can_submit`
- `practice_enabled`

### 7.7 报名比赛
- 方法：`POST`
- 路径：`/api/contests/:id/register`
- 权限：登录

### 7.8 比赛题目列表
- 方法：`GET`
- 路径：`/api/contests/:id/problems`
- 权限：登录

### 7.9 比赛内题目详情
- 方法：`GET`
- 路径：`/api/contests/:id/problems/:problem_id`
- 权限：登录

## 8. 论坛模块 `/api/forum`

### 8.1 主题列表
- 方法：`GET`
- 路径：`/api/forum/topics`
- 权限：公开
- 查询参数：
  - `page`
  - `page_size`
  - `keyword`
  - `scope_type`：`general | problem | contest`
  - `scope_id`

### 8.2 主题详情
- 方法：`GET`
- 路径：`/api/forum/topics/:id`
- 权限：公开

### 8.3 发帖
- 方法：`POST`
- 路径：`/api/forum/topics`
- 权限：登录

请求体：

```json
{
  "title": "How should I solve Problem 1002?",
  "content": "I want to discuss the idea.",
  "scope_type": "problem",
  "scope_id": 3
}
```

### 8.4 编辑帖子
- 方法：`PUT`
- 路径：`/api/forum/topics/:id`
- 权限：登录（作者或管理员/教师）

请求体：

```json
{
  "title": "Updated title",
  "content": "Updated content",
  "is_pinned": true,
  "is_locked": false
}
```

说明：
- `is_pinned` / `is_locked` 通常由教师或管理员控制

### 8.5 删除帖子
- 方法：`DELETE`
- 路径：`/api/forum/topics/:id`
- 权限：登录

### 8.6 回帖
- 方法：`POST`
- 路径：`/api/forum/topics/:id/replies`
- 权限：登录

请求体：

```json
{
  "content": "I think a direct simulation is enough."
}
```

### 8.7 编辑回帖
- 方法：`PUT`
- 路径：`/api/forum/replies/:id`
- 权限：登录

### 8.8 删除回帖
- 方法：`DELETE`
- 路径：`/api/forum/replies/:id`
- 权限：登录

## 9. 题单、班级、作业 `/api/playlists` `/api/classes` `/api/assignments`

### 9.1 公开题单列表
- 方法：`GET`
- 路径：`/api/playlists`
- 权限：公开
- 查询参数：
  - `page`
  - `page_size`
  - `keyword`

### 9.2 题单详情
- 方法：`GET`
- 路径：`/api/playlists/:id`
- 权限：公开

### 9.3 我的班级
- 方法：`GET`
- 路径：`/api/classes/my`
- 权限：登录

### 9.4 加入班级
- 方法：`POST`
- 路径：`/api/classes/join`
- 权限：登录

请求体：

```json
{
  "join_code": "ABCD1234"
}
```

### 9.5 班级详情
- 方法：`GET`
- 路径：`/api/classes/:id`
- 权限：登录

### 9.6 作业详情
- 方法：`GET`
- 路径：`/api/assignments/:id`
- 权限：登录

## 10. 教师接口 `/api/teacher`

权限：教师或管理员。

### 10.1 题单管理
- `GET /api/teacher/playlists`
- `GET /api/teacher/playlists/:id`
- `POST /api/teacher/playlists`
- `PUT /api/teacher/playlists/:id`

创建/更新题单请求体：

```json
{
  "title": "Week 1 Training",
  "description": "Basic loops and conditionals",
  "visibility": "class",
  "problems": [
    { "problem_id": 3, "display_order": 1 },
    { "problem_id": 4, "display_order": 2 }
  ]
}
```

### 10.2 题解管理
- `GET /api/teacher/problems/:id/solutions`
- `POST /api/teacher/problems/:id/solutions`
- `PUT /api/teacher/problems/:id/solutions/:solution_id`
- `DELETE /api/teacher/problems/:id/solutions/:solution_id`

请求体：

```json
{
  "title": "Official Solution",
  "content": "Use a simple linear scan.",
  "visibility": "public"
}
```

### 10.3 班级管理
- `GET /api/teacher/classes`
- `POST /api/teacher/classes`
- `GET /api/teacher/classes/:id`
- `GET /api/teacher/classes/:id/analytics`
- `PUT /api/teacher/classes/:id`
- `GET /api/teacher/classes/:id/members`
- `PUT /api/teacher/classes/:id/members/:user_id`

班级分析接口返回：
- `member_count`
- `assignment_count`
- `unique_problem_count`
- `overall_completion_rate`
- `assignments`
- `top_students`

更新班级成员请求体：

```json
{
  "role": "assistant",
  "status": "active"
}
```

### 10.4 作业管理
- `GET /api/teacher/classes/:id/assignments`
- `POST /api/teacher/classes/:id/assignments`
- `GET /api/teacher/assignments/:id`
- `PUT /api/teacher/assignments/:id`
- `DELETE /api/teacher/assignments/:id`

创建作业请求体：

```json
{
  "playlist_id": 1,
  "title": "Homework 1",
  "description": "Finish the first playlist.",
  "type": "homework",
  "start_at": "2026-04-10T08:00:00Z",
  "due_at": "2026-04-17T15:59:59Z"
}
```

## 11. 管理员接口 `/api/admin`

权限：管理员。

### 11.1 题目管理
- `GET /api/admin/problems`
- `POST /api/admin/problems`
- `GET /api/admin/problems/:id`
- `PUT /api/admin/problems/:id`
- `DELETE /api/admin/problems/:id`

创建/更新题目请求体：

```json
{
  "display_id": "1001",
  "title": "A + B Problem",
  "description": "输入两个整数，输出它们的和",
  "input_desc": "输入两个整数 a, b",
  "output_desc": "输出一个整数",
  "sample_input": "1 2",
  "sample_output": "3",
  "hint": "",
  "source": "SEU OJ",
  "judge_mode": "standard",
  "time_limit_ms": 1000,
  "memory_limit_mb": 128,
  "visible": true,
  "testcases": [
    {
      "case_type": "sample",
      "input_data": "1 2",
      "output_data": "3",
      "score": 0,
      "sort_order": 1,
      "is_active": true
    },
    {
      "case_type": "hidden",
      "input_data": "2 3",
      "output_data": "5",
      "score": 100,
      "sort_order": 2,
      "is_active": true
    }
  ]
}
```

### 11.2 比赛管理
- `GET /api/admin/contests`
- `POST /api/admin/contests`
- `GET /api/admin/contests/:id`
- `PUT /api/admin/contests/:id`
- `DELETE /api/admin/contests/:id`
- `GET /api/admin/contests/:id/ranklist`

创建/更新比赛请求体：

```json
{
  "title": "SEU Spring Warmup 2026",
  "description": "Warmup contest for SEU OJ.",
  "rule_type": "acm",
  "start_time": "2026-04-09T11:00:00Z",
  "end_time": "2026-04-09T14:00:00Z",
  "is_public": true,
  "allow_practice": true,
  "ranklist_freeze_at": "2026-04-09T13:00:00Z",
  "problems": [
    { "problem_id": 3, "problem_code": "A", "display_order": 1 },
    { "problem_id": 4, "problem_code": "B", "display_order": 2 }
  ]
}
```

### 11.3 比赛公告管理
- `GET /api/admin/contests/:id/announcements`
- `GET /api/admin/contests/:id/announcements/:announcement_id`
- `POST /api/admin/contests/:id/announcements`
- `PUT /api/admin/contests/:id/announcements/:announcement_id`
- `DELETE /api/admin/contests/:id/announcements/:announcement_id`

### 11.4 提交管理
- `GET /api/admin/submissions`
- `POST /api/admin/submissions/:id/rejudge`

常用查询参数：
- `page`
- `page_size`
- `user_id`
- `problem_id`
- `contest_id`
- `status`

### 11.5 用户管理
- `GET /api/admin/users`
- `PUT /api/admin/users/:id`

更新用户请求体：

```json
{
  "username": "alice",
  "userid": "213220001",
  "role": "teacher",
  "status": "active"
}
```

### 11.6 站内公告管理
- `POST /api/admin/announcements`
- `PUT /api/admin/announcements/:id`
- `DELETE /api/admin/announcements/:id`

请求体：

```json
{
  "title": "System Maintenance",
  "content": "The system will be unavailable tonight.",
  "is_pinned": true
}
```

## 12. 主要状态值约定

### 12.1 提交状态
- `Pending`
- `Running`
- `Accepted`
- `Wrong Answer`
- `Compile Error`
- `Runtime Error`
- `Time Limit Exceeded`
- `System Error`

### 12.2 比赛状态
- `upcoming`
- `running`
- `ended`

### 12.3 论坛范围
- `general`
- `problem`
- `contest`

### 12.4 题单可见性
- `public`
- `private`
- `class`

### 12.5 作业类型
- `homework`
- `exam`

## 13. 建议联调顺序

推荐前后端联调顺序：

1. `auth`：注册、登录、获取当前用户
2. `problems`：题目列表、题目详情、运行样例
3. `submissions`：创建提交、我的提交、提交详情
4. `contests`：比赛列表、详情、榜单、比赛内题目
5. `playlists/classes/assignments`：题单、班级、作业
6. `forum`：帖子列表、详情、发帖回帖
7. `teacher/admin`：教师端、管理端接口

## 14. 说明

- 文档基于当前代码实现，不是 Swagger 自动生成结果。
- 如果后续新增路由，建议同步维护本文件。
- 如果你们需要对外展示或自动测试，下一步建议补：
  - `OpenAPI / Swagger`
  - Postman Collection
  - `docs/example-requests.md`




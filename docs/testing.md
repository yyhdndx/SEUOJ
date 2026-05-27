# 测试文档

本文档说明 SEU OJ 后端测试的组织方式、运行命令和覆盖范围。用户提到的“继承测试”这里按软件工程常用的“集成测试”落地；仓库中目前也没有需要专门做继承层级验证的业务模型。

## 测试类型

### 单元测试

单元测试覆盖不依赖外部服务的纯逻辑和边界处理：

- 配置加载与环境变量覆盖：`internal/config`
- JWT 生成、解析、过期和签名错误：`internal/utils`
- Gin 中间件：JWT、可选 JWT、CORS、限流、角色鉴权、请求耗时、审计辅助函数
- 响应 envelope：`internal/response`
- 观测工具：`internal/observability`
- 沙箱配置、语言规格、Docker 参数拼装、输出截断缓冲区
- 业务纯函数：题目包解析、输出对比、竞赛状态、教学进度等

### 集成测试

集成测试使用 GORM SQLite 内存库，不依赖本机 MySQL、Redis 或 Docker：

- Repository 层：题目、测试点、提交、提交结果 CRUD 和过滤排序
- Service 层：认证、题目生命周期、题解、题目包导入导出、公告、竞赛、论坛、榜单、统计、用户管理
- Router 层：`/api/auth/*` 登录注册流程、管理员题目 CRUD、公开题目详情/统计、健康检查、审计日志写入

SQLite 只用于测试本地业务逻辑和 GORM 调用路径。生产数据库仍是 MySQL。

## 运行命令

在后端目录执行：

```powershell
cd "D:\desk\软件工程\SEUOJ\seu-oj-backend"
go test ./...
```

生成覆盖率：

```powershell
go test ./... -coverprofile=coverage
go tool cover -func=coverage
```

生成 HTML 覆盖率报告：

```powershell
go tool cover -html=coverage -o coverage.html
```

`coverage` 和 `coverage.html` 是本地生成文件，已加入 `.gitignore`。

## 当前覆盖率基线

2026-05-27 本地验证结果：

- 全量测试：`go test ./...` 通过
- 总语句覆盖率：`34.0%`
- 高覆盖核心包：`config 96.2%`、`middleware 83.0%`、`repository 88.3%`、`router 88.8%`、`response 100.0%`、`utils 90.9%`
- `service` 覆盖率：`44.1%`

覆盖率较低的主要是需要真实 Docker/Redis/MySQL 或更重业务数据编排的入口和长流程，例如 `main`、`database.New`、`queue`、`judge.Worker`、部分教学模块。

## 后续扩展建议

新增业务时优先按以下顺序补测试：

1. 先测纯函数和参数校验。
2. Repository/Service 使用 SQLite 内存库覆盖正常路径、权限错误、not found、重复数据。
3. API/Router 用 `httptest` 走真实 Gin 路由，验证鉴权、响应 envelope 和关键副作用。
4. Docker 判题、Redis 队列、MySQL 方言行为单独放到需要外部服务的端到端测试，不混入默认 `go test ./...`。

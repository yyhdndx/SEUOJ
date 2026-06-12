# 部署说明

本文档说明如何用 Docker Compose 在一台机器上启动 SEU OJ 的演示环境。

## 适用场景

该方案适合课程设计演示、组内联调和验收复现。它默认使用远程 MySQL，并启动：

- Redis：判题队列和缓存
- Docker DinD：供判题沙箱创建语言运行容器
- db-init：一次性数据库建表和导入演示数据
- Web：后端 API 与静态前端
- Worker：判题 Worker

仓库也保留了可选的本地 MySQL profile，便于没有远程数据库时完整本地复现。

## 前置要求

- Docker Desktop 或 Docker Engine
- Docker Compose v2
- 首次启动需要联网拉取基础镜像和判题语言镜像

如果 Docker Hub 访问慢或超时，可以在 Docker Desktop 的 daemon 配置中加入镜像源：

```json
{
  "registry-mirrors": [
    "https://docker.m.daocloud.io",
    "https://docker.1panel.live",
    "https://docker.1ms.run"
  ]
}
```

修改后需要重启 Docker Desktop。当前 `docker-compose.yml` 里的内部 Docker DinD 服务也已经配置了同样的 mirror。

## 快速启动

在仓库根目录执行：

```bash
cp .env.example .env
```

编辑 `.env`，至少填入远程 MySQL 密码：

```env
DB_HOST=8.140.36.87
DB_PORT=3306
DB_USER=remote_admin
DB_PASSWORD=your_remote_mysql_password
DB_NAME=seuoj
```

然后启动：

```bash
docker compose up -d --build
```

首次启动会拉取这些判题镜像，耗时可能较长：

- `gcc:13`
- `python:3`
- `eclipse-temurin:21`
- `golang:1`
- `rust:1`

启动完成后访问：

```text
http://127.0.0.1:8080/
```

查看服务状态：

```bash
docker compose ps
```

查看日志：

```bash
docker compose logs -f web worker
```

停止服务：

```bash
docker compose down
```

如果需要连数据库、语言镜像和演示数据一起清空：

```bash
docker compose down -v
```

## 可配置环境变量

可以在仓库根目录创建 `.env` 覆盖默认值：

```env
SEUOJ_WEB_PORT=8080
DB_HOST=8.140.36.87
DB_PORT=3306
DB_USER=remote_admin
DB_PASSWORD=your_remote_mysql_password
DB_NAME=seuoj
JWT_SECRET=change-me-before-demo
```

默认 Web 端口是 `8080`。如果本机端口冲突，可以改成：

```env
SEUOJ_WEB_PORT=18080
```

然后访问 `http://127.0.0.1:18080/`。

## 使用本地 MySQL

如果不使用远程 MySQL，而是希望 Compose 同时启动本地 MySQL，把 `.env` 改成：

```env
DB_HOST=mysql
DB_PORT=3306
DB_USER=root
DB_PASSWORD=seuoj_root
DB_NAME=seu_oj
MYSQL_ROOT_PASSWORD=seuoj_root
MYSQL_DATABASE=seu_oj
```

然后使用 profile 启动：

```bash
docker compose --profile local-db up -d --build
```

## 判题沙箱说明

Compose 中没有直接挂载宿主机 Docker socket，而是启动了一个 `docker:dind` 服务作为内部 Docker 引擎。Web 的 `Run` 样例测试和 Worker 的正式判题都会通过 `DOCKER_HOST=tcp://docker:2375` 调用这个内部 Docker 引擎。

`judge-work` 卷会同时挂载给 Web、Worker 和 Docker DinD，确保用户代码临时目录能被沙箱容器正确挂载。

## 常见问题

### Submit 一直 Pending

通常是 Worker 没有正常运行。检查：

```bash
docker compose ps worker
docker compose logs worker
```

如果日志提示缺少语言镜像，执行：

```bash
docker compose up sandbox-images
docker compose restart worker
```

### 首次启动很慢

首次启动需要构建应用镜像，并拉取 Redis、Docker DinD 和多种语言运行镜像。后续因为有 Docker 缓存和 `docker-data` 卷，会明显变快。只有使用 `local-db` profile 时才会额外拉取 MySQL 镜像。

### 想重新导入演示数据

最干净的方式是删除卷：

```bash
docker compose down -v
docker compose up -d --build
```

这会清空本地 Redis 数据、Docker DinD 镜像缓存和日志卷。如果使用了 `local-db` profile，也会清空本地 MySQL 数据；远程 MySQL 不会被 `docker compose down -v` 删除。

# Flowboard 部署示例

这是一个可直接运行的前后端分离项目，用于演示「推送代码后自动部署」与 Jenkins 集成。

- `frontend`：Vue 3 + TypeScript 任务看板。
- `backend`：Go REST API，使用 SQLite 持久化任务。
- `docker-compose.yml`：生产运行编排；前端通过 Nginx 反向代理 API。
- `Jenkinsfile`：Jenkins Pipeline 构建并部署。
- `.github/workflows/deploy.yml`：推送 `main` 后构建镜像并通过 SSH 在服务器执行部署。

## 本地启动

要求：Docker Desktop（推荐），或 Go 1.22+ 与 Node 20+。

```bash
cd deploy-1
docker compose up --build
```

打开 <http://localhost:8080>。SQLite 数据保存于 Docker volume `taskboard_data`，容器重建不会丢失。

开发时可分开运行：

```bash
# 终端 1
cd backend && go run .

# 终端 2
cd frontend && npm install && npm run dev
```

前端开发服务器会把 `/api` 代理到 `http://localhost:8081`。

## 自动部署前置条件

1. 将本目录作为独立仓库推送到 GitHub，默认分支为 `main`。
2. 在服务器安装 Docker 与 Docker Compose Plugin，并把仓库克隆到 `/opt/flowboard`。
3. 在 GitHub 仓库 Settings → Secrets and variables → Actions 添加：
   `DEPLOY_HOST`、`DEPLOY_USER`、`DEPLOY_SSH_KEY`、`DEPLOY_PORT`（可选）。
4. 服务器上的部署用户需要可运行 Docker。每次推送到 `main`，GitHub Actions 会通过 SSH 执行 `scripts/deploy.sh`。

## Jenkins 集成

在 Jenkins 新建 Pipeline，配置仓库与凭据，脚本路径填写 `Jenkinsfile`。在凭据中创建：

- `registry-credentials`：容器镜像仓库用户名/密码（若使用私有镜像仓库）。
- `deploy-ssh-key`：服务器 SSH 私钥。

Jenkins 环境变量 `DEPLOY_HOST`、`DEPLOY_USER`、`DEPLOY_PORT` 可在任务配置或全局配置中设置。流水线会先验证 Go 测试和前端构建，再将代码同步到服务器并执行部署脚本。

> GitHub Actions 与 Jenkins 两者择一启用即可，避免同一次提交触发重复部署。

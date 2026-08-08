#!/usr/bin/env sh
set -eu

# 用法：sh scripts/deploy.sh [all|frontend|backend]
# 决定要重建哪些服务；Jenkins 通过 PROJECT 参数传入，实现"指定项目部署"
PROJECT="${1:-all}"

case "$PROJECT" in
  all)
    # 前后端一起重建重启
    docker compose up -d --build --remove-orphans
    ;;
  frontend)
    # 只重建前端；--no-deps 避免连后端一起拉起/重建，后端容器保持不动
    docker compose up -d --build --no-deps frontend
    ;;
  backend)
    docker compose up -d --build --no-deps backend
    ;;
  *)
    echo "未知项目：$PROJECT（可选 all / frontend / backend）" >&2
    exit 1
    ;;
esac

# 清理旧的悬空镜像，避免磁盘被反复构建占满
docker image prune -f

# 基于 Jenkins 官方 LTS 镜像（自带 JDK 21）
# 叠加 Go 1.22 + Node 20，用于构建本项目的后端/前端
FROM jenkins/jenkins:lts-jdk21

USER root

# ---------- 安装 Go 1.22（验证/构建后端） ----------
ENV GO_VERSION=1.22.12
RUN curl -fsSL "https://golang.google.cn/dl/go${GO_VERSION}.linux-amd64.tar.gz" \
    | tar -C /usr/local -xzv \
    && /usr/local/go/bin/go version

# ---------- 安装 Node 20（构建前端） ----------
# Node 发布包是 .tar.xz，先安装 xz 解压工具
ENV NODE_VERSION=20.19.4
RUN apt-get update && apt-get install -y xz-utils && rm -rf /var/lib/apt/lists/*
RUN curl -fsSL "https://npmmirror.com/mirrors/node/v${NODE_VERSION}/node-v${NODE_VERSION}-linux-x64.tar.xz" \
    | tar -C /usr/local -xJ \
    && mv /usr/local/node-v${NODE_VERSION}-linux-x64 /usr/local/node \
    && /usr/local/node/bin/node --version

# ---------- 配置国内加速（依赖下载） ----------
# Go 模块走 goproxy.cn
ENV GOPROXY=https://goproxy.cn,direct
# npm 全局镜像源（写入全局 npmrc，jenkins 用户也能读到）
RUN mkdir -p /usr/local/node/etc \
    && echo "registry=https://registry.npmmirror.com" > /usr/local/node/etc/npmrc

# ---------- 把 Go/Node 加进 PATH ----------
ENV PATH=/usr/local/go/bin:/usr/local/node/bin:$PATH

# 切回官方镜像默认的 jenkins 用户（避免权限问题）
USER jenkins

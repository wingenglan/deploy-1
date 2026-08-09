# 基于 Jenkins 官方 LTS 镜像（自带 JDK 21）
# 该镜像只做"调度者"：安装 docker CLI，用于驱动宿主机 Docker 引擎拉起 agent 容器
# 具体的构建工具（Go/Node/Python 等）不再装进 Jenkins，改由每个阶段的 agent 容器提供
FROM jenkins/jenkins:lts-jdk21

USER root

# 安装 docker CLI（静态二进制，与宿主机 Docker Desktop 引擎版本一致）
# 用阿里云镜像加速；download.docker.com 在国内可能无法访问
ARG DOCKER_VERSION=27.3.1
RUN curl -fsSL "https://mirrors.aliyun.com/docker-ce/linux/static/stable/x86_64/docker-${DOCKER_VERSION}.tgz" \
    | tar -xz -C /tmp docker/docker \
    && mv /tmp/docker/docker /usr/local/bin/docker \
    && rm -rf /tmp/docker \
    && docker --version

# 切回官方镜像默认的 jenkins 用户（避免权限问题）
USER jenkins

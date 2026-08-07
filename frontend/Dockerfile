FROM node:20-alpine AS build
# 使用国内可访问的 npm 镜像（解决 npm 官方源慢/超时问题）
RUN npm config set registry https://registry.npmmirror.com
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM nginx:1.27-alpine
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /app/dist /usr/share/nginx/html
EXPOSE 80

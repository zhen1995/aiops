# AIOPS 前端镜像（多阶段构建：Vite 构建 + Nginx 托管，/api 反代到 backend 容器）
# 构建上下文为仓库根目录（compose 中 context: .），以便复用 deploy/nginx/default.conf
FROM node:20-alpine AS build
WORKDIR /src
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

FROM nginx:alpine
COPY --from=build /src/dist /usr/share/nginx/html
COPY deploy/nginx/default.conf /etc/nginx/conf.d/default.conf
EXPOSE 80

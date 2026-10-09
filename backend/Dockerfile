# AIOPS Go 后端镜像（多阶段构建）
# GOPROXY 可通过 --build-arg 覆盖（如海外服务器用 https://proxy.golang.org,direct）
ARG GOPROXY=https://goproxy.cn,direct
FROM golang:1.26-alpine AS build
ARG GOPROXY
WORKDIR /src
ENV CGO_ENABLED=0
ENV GOPROXY=${GOPROXY}
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -trimpath -o /out/center ./cmd/center

FROM alpine:3
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=build /out/center /app/center
# 不打包 config.yaml：容器内通过环境变量注入全部配置（见 docker-compose.yml），避免把开发环境 DSN 打进镜像
EXPOSE 8080
CMD ["/app/center"]

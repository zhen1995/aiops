# AIOPS Go 后端镜像（多阶段构建）
FROM golang:1.26-alpine AS build
WORKDIR /src
ENV CGO_ENABLED=0
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

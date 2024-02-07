# 前端构建
FROM node:20-alpine AS frontend-builder

WORKDIR /app
COPY webui/ ./
RUN npm install -g pnpm && \
    pnpm install && \
    pnpm build && \
    rm -rf node_modules

# 后端构建
FROM golang:1.23 AS backend-builder

WORKDIR /app
COPY backend/ ./
COPY --from=frontend-builder /app/dist/ ./web/

# 添加架构检查和交叉编译支持
RUN go version && \
    arch=$(uname -m) && \
    echo "Building for architecture: $arch" && \
    if [ "$arch" = "aarch64" ]; then \
        export GOARCH=arm64; \
    elif [ "$arch" = "x86_64" ]; then \
        export GOARCH=amd64; \
    else \
        export GOARCH=$arch; \
    fi && \
    echo "Using GOARCH=$GOARCH" && \
    GOOS=linux CGO_ENABLED=0 go build -o BangumiBuddy && \
    chmod +x BangumiBuddy && \
    ls -la

FROM alpine:latest

WORKDIR /app
RUN mkdir -p /data/log /config
COPY --from=backend-builder /app/BangumiBuddy /app/
COPY --from=backend-builder /app/web/ /app/web/

# 添加调试命令
RUN chmod +x /app/BangumiBuddy 

EXPOSE 6937

# 创建数据卷
VOLUME ["/config", "/data"]

CMD ["/app/BangumiBuddy"] 
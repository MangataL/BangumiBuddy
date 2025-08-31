# 前端构建
FROM node:20-alpine AS frontend-builder

WORKDIR /app
COPY webui/ ./
RUN npm install -g pnpm && \
    pnpm install && \
    pnpm build && \
    rm -rf node_modules

# 后端构建
FROM golang:1.24 AS backend-builder

# 安装 HarfBuzz 开发库和构建工具
RUN apt-get update && apt-get install -y \
    libharfbuzz-dev \
    pkg-config \
    gcc \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY backend/ ./
COPY --from=frontend-builder /app/dist/ ./web/

# 设置CGO环境变量并构建
ENV CGO_ENABLED=1
ENV GOOS=linux
RUN go version && \
    go build -o BangumiBuddy && \
    chmod +x BangumiBuddy && \
    ls -la BangumiBuddy

# 运行时镜像 - 使用Debian以保持与构建环境的一致性
FROM debian:bookworm-slim

# 安装运行时依赖
RUN apt-get update && apt-get install -y \
    ca-certificates \
    tzdata \
    libharfbuzz0b \
    libharfbuzz-subset0 \
    && rm -rf /var/lib/apt/lists/* \
    && ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone

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
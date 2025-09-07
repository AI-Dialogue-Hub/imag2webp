# 使用多阶段构建来减小镜像大小
FROM golang:1.25-alpine AS builder

# 安装必要的编译依赖
RUN apk add --no-cache \
    gcc \
    g++ \
    musl-dev \
    libwebp-dev \
    make

# 设置工作目录
WORKDIR /app

# 复制go.mod和go.sum文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用（使用构建参数）
ARG VERSION=dev
ARG BUILD_DATE=unknown
RUN CGO_ENABLED=1 GOOS=linux go build \
    -ldflags="-X 'main.version=${VERSION}' -X 'main.buildDate=${BUILD_DATE}'" \
    -o image2webp ./cmd/main.go

# 创建最终运行镜像
FROM alpine:latest

# 安装运行时依赖
RUN apk add --no-cache \
    libwebp \
    ca-certificates \
    curl

# 创建非root用户
RUN adduser -D -u 1000 webpuser

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder --chown=webpuser:webpuser /app/image2webp .

# 复制静态文件（如果有）
COPY --chown=webpuser:webpuser ./front ./front

# 切换到非root用户
USER webpuser

# 暴露端口（使用ARG支持动态端口）
ARG PORT=10080
EXPOSE ${PORT}

# 设置环境变量默认值
ENV PORT=${PORT}
ENV MAX_UPLOAD_SIZE=33554432
ENV DEFAULT_QUALITY=80
ENV DEFAULT_LOSSLESS=false
ENV LOG_LEVEL=info
ENV LOG_FORMAT=json

# 健康检查（使用环境变量端口）
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:${PORT}/v1/health || exit 1

# 启动应用
CMD ["./image2webp"]
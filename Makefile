# Makefile for image2webp project

# 加载环境变量
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

# Variables with defaults
APP_NAME ?= image2webp
DOCKER_IMAGE ?= image2webp
DOCKER_TAG ?= latest
PORT ?= 10080
VERSION ?= dev
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

.PHONY: help build run docker-build docker-run docker-build-env

help:
	@echo "Available commands:"
	@echo "  make build          - Build the Go binary"
	@echo "  make run            - Run the application locally"
	@echo "  make docker-build   - Build Docker image with default settings"
	@echo "  make docker-build-env - Build Docker image using .env variables"
	@echo "  make docker-run     - Run Docker container"

# Local development
build:
	@echo "Building Go binary..."
	CGO_ENABLED=1 go build -o $(APP_NAME) ./cmd/main.go

run: build
	@echo "Starting application on port $(PORT)..."
	PORT=$(PORT) ./$(APP_NAME)

# Docker builds
docker-build:
	@echo "Building Docker image $(DOCKER_IMAGE):$(DOCKER_TAG)..."
	docker build \
		--build-arg PORT=$(PORT) \
		--build-arg VERSION=$(VERSION) \
		--build-arg BUILD_DATE=$(BUILD_DATE) \
		-t $(DOCKER_IMAGE):$(DOCKER_TAG) .

docker-build-env:
	@echo "Building Docker image using .env variables..."
	docker build \
		--build-arg PORT=${PORT} \
		--build-arg VERSION=${VERSION} \
		--build-arg BUILD_DATE=${BUILD_DATE} \
		-t $(DOCKER_IMAGE):$(DOCKER_TAG) .

docker-run: docker-build-env
	@echo "Starting Docker container on port $(PORT)..."
	docker run -d \
		-p $(PORT):$(PORT) \
		--name $(APP_NAME) \
		-e PORT=$(PORT) \
		-e MAX_UPLOAD_SIZE=$(MAX_UPLOAD_SIZE) \
		-e DEFAULT_QUALITY=$(DEFAULT_QUALITY) \
		-e DEFAULT_LOSSLESS=$(DEFAULT_LOSSLESS) \
		-e LOG_LEVEL=$(LOG_LEVEL) \
		-e LOG_FORMAT=$(LOG_FORMAT) \
		$(DOCKER_IMAGE):$(DOCKER_TAG)

docker-run-env: docker-build-env
	@echo "Starting Docker container with all .env variables..."
	docker run -d \
		-p $(PORT):$(PORT) \
		--name $(APP_NAME) \
		--env-file .env \
		$(DOCKER_IMAGE):$(DOCKER_TAG)

# 其他命令保持不变...
docker-stop:
	docker stop $(APP_NAME) || true

docker-rm: docker-stop
	docker rm $(APP_NAME) || true

docker-logs:
	docker logs -f $(APP_NAME)

docker-clean: docker-rm
	docker rmi $(DOCKER_IMAGE):$(DOCKER_TAG) || true

health-check:
	curl -f http://localhost:$(PORT)/v1/health || echo "Health check failed"

# 显示当前环境变量配置
show-env:
	@echo "Current environment configuration:"
	@echo "  PORT: $(PORT)"
	@echo "  MAX_UPLOAD_SIZE: $(MAX_UPLOAD_SIZE)"
	@echo "  DEFAULT_QUALITY: $(DEFAULT_QUALITY)"
	@echo "  DEFAULT_LOSSLESS: $(DEFAULT_LOSSLESS)"
	@echo "  LOG_LEVEL: $(LOG_LEVEL)"
	@echo "  VERSION: $(VERSION)"
# 简单实用的文档问答RAG系统 Makefile

# 变量定义
BINARY_NAME=docs-qa
VERSION=0.1.0
BUILD_DIR=bin
DIST_DIR=dist
MAIN_PATH=./cmd

# 默认目标
.PHONY: all
all: build

# 构建（当前平台）
.PHONY: build
build:
	@echo "🔨 构建 $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "✅ 构建完成: $(BUILD_DIR)/$(BINARY_NAME)"

# 运行
.PHONY: run
run: build
	@echo "🚀 运行程序..."
	@$(BUILD_DIR)/$(BINARY_NAME) docs "退款流程是怎样的？"

# 清理
.PHONY: clean
clean:
	@echo "🧹 清理..."
	@rm -rf $(BUILD_DIR) $(DIST_DIR)
	@echo "✅ 清理完成"

# 测试
.PHONY: test
test:
	@echo "🧪 运行测试..."
	go test ./...
	@echo "✅ 测试完成"

# 格式化代码
.PHONY: fmt
fmt:
	@echo "🎨 格式化代码..."
	gofmt -w .
	@echo "✅ 格式化完成"

# 跨平台构建
.PHONY: cross-build
cross-build: build-linux build-mac build-windows

# 构建 Linux
.PHONY: build-linux
build-linux:
	@echo "🐧 构建 Linux 版本..."
	@mkdir -p $(BUILD_DIR)/linux
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/linux/$(BINARY_NAME) $(MAIN_PATH)
	@echo "✅ Linux 版本: $(BUILD_DIR)/linux/$(BINARY_NAME)"

# 构建 Mac
.PHONY: build-mac
build-mac:
	@echo "🍎 构建 Mac 版本..."
	@mkdir -p $(BUILD_DIR)/mac
	GOOS=darwin GOARCH=arm64 go build -o $(BUILD_DIR)/mac/$(BINARY_NAME) $(MAIN_PATH)
	@echo "✅ Mac 版本: $(BUILD_DIR)/mac/$(BINARY_NAME)"

# 构建 Windows
.PHONY: build-windows
build-windows:
	@echo "🪟 构建 Windows 版本..."
	@mkdir -p $(BUILD_DIR)/windows
	GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/windows/$(BINARY_NAME).exe $(MAIN_PATH)
	@echo "✅ Windows 版本: $(BUILD_DIR)/windows/$(BINARY_NAME).exe"

# 创建发布包
.PHONY: release
release: clean cross-build
	@echo "📦 创建发布包..."
	@mkdir -p $(DIST_DIR)
	@cd $(BUILD_DIR)/linux && tar -czf ../../$(DIST_DIR)/$(BINARY_NAME)-linux-amd64.tar.gz $(BINARY_NAME)
	@cd $(BUILD_DIR)/mac && tar -czf ../../$(DIST_DIR)/$(BINARY_NAME)-mac-arm64.tar.gz $(BINARY_NAME)
	@cd $(BUILD_DIR)/windows && zip ../../$(DIST_DIR)/$(BINARY_NAME)-windows-amd64.zip $(BINARY_NAME).exe
	@echo "✅ 发布包创建完成: $(DIST_DIR)/"

# Docker 构建
.PHONY: docker-build
docker-build:
	@echo "🐳 构建 Docker 镜像..."
	docker build -t $(BINARY_NAME):$(VERSION) -t $(BINARY_NAME):latest .
	@echo "✅ Docker 镜像构建完成"

# Docker 运行
.PHONY: docker-run
docker-run: docker-build
	@echo "🐳 运行 Docker 容器..."
	docker run -p 8080:8080 -v $(PWD)/docs:/app/docs $(BINARY_NAME):latest

# 安装到系统
.PHONY: install
install: build
	@echo "📦 安装到系统..."
	@sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "✅ 安装完成"

# 卸载
.PHONY: uninstall
uninstall:
	@echo "🗑️ 卸载..."
	@sudo rm -f /usr/local/bin/$(BINARY_NAME)
	@echo "✅ 卸载完成"

# 帮助信息
.PHONY: help
help:
	@echo "📚 可用命令:"
	@echo "  make build        构建当前平台"
	@echo "  make run          构建并运行"
	@echo "  make clean        清理构建文件"
	@echo "  make test         运行测试"
	@echo "  make fmt          格式化代码"
	@echo "  make cross-build  构建所有平台"
	@echo "  make build-linux  构建 Linux 版本"
	@echo "  make build-mac    构建 Mac 版本"
	@echo "  make build-windows 构建 Windows 版本"
	@echo "  make release      创建发布包"
	@echo "  make docker-build 构建 Docker 镜像"
	@echo "  make docker-run   运行 Docker 容器"
	@echo "  make install      安装到系统"
	@echo "  make uninstall    卸载"
	@echo "  make help         显示帮助"

# 显示版本
.PHONY: version
version:
	@echo "$(BINARY_NAME) v$(VERSION)"
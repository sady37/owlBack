# wisefido-qinglan Makefile

.PHONY: all build run test clean docker-build docker-run docker-compose-up docker-compose-down help

# 默认目标
all: build

# 构建应用
build:
	@echo "Building wisefido-qinglan..."
	go build -o wisefido-qinglan cmd/wisefido-qinglan/main.go
	@echo "Build completed: wisefido-qinglan"

# 运行应用
run: build
	@echo "Starting wisefido-qinglan..."
	./wisefido-qinglan

# 开发模式运行（自动重新加载）
dev:
	@echo "Starting wisefido-qinglan in development mode..."
	air

# 测试
test:
	@echo "Running tests..."
	go test ./... -v

# 代码格式化
fmt:
	@echo "Formatting code..."
	go fmt ./...

# 代码检查
lint:
	@echo "Running linter..."
	golangci-lint run

# 清理构建文件
clean:
	@echo "Cleaning up..."
	rm -f wisefido-qinglan
	rm -rf dist/
	@echo "Clean completed"

# Docker构建
docker-build:
	@echo "Building Docker image..."
	docker build -t wisefido-qinglan:latest .

# Docker运行
docker-run: docker-build
	@echo "Running Docker container..."
	docker run -p 8081:8081 -v $(PWD)/config.yaml:/app/config.yaml wisefido-qinglan:latest

# Docker Compose启动
docker-compose-up:
	@echo "Starting services with Docker Compose..."
	docker-compose up -d

# Docker Compose停止
docker-compose-down:
	@echo "Stopping services with Docker Compose..."
	docker-compose down

# Docker Compose查看日志
docker-compose-logs:
	@echo "Showing logs..."
	docker-compose logs -f

# 数据库迁移
migrate:
	@echo "Running database migrations..."
	# 这里可以添加数据库迁移命令

# 生成Swagger文档
swagger:
	@echo "Generating Swagger documentation..."
	# 这里可以添加Swagger生成命令

# 性能测试
bench:
	@echo "Running benchmarks..."
	go test -bench=. ./...

# 覆盖率测试
coverage:
	@echo "Running coverage analysis..."
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# 依赖更新
update-deps:
	@echo "Updating dependencies..."
	go get -u ./...
	go mod tidy

# 安全检查
security-check:
	@echo "Running security checks..."
	gosec ./...

# 帮助信息
help:
	@echo "wisefido-qinglan Makefile Commands:"
	@echo ""
	@echo "  build              - Build the application"
	@echo "  run                - Build and run the application"
	@echo "  dev                - Run in development mode with auto-reload"
	@echo "  test               - Run tests"
	@echo "  fmt                - Format code"
	@echo "  lint               - Run linter"
	@echo "  clean              - Clean build files"
	@echo "  docker-build       - Build Docker image"
	@echo "  docker-run         - Run Docker container"
	@echo "  docker-compose-up  - Start services with Docker Compose"
	@echo "  docker-compose-down - Stop services with Docker Compose"
	@echo "  docker-compose-logs - View Docker Compose logs"
	@echo "  migrate            - Run database migrations"
	@echo "  swagger            - Generate Swagger documentation"
	@echo "  bench              - Run benchmarks"
	@echo "  coverage           - Run coverage analysis"
	@echo "  update-deps        - Update dependencies"
	@echo "  security-check     - Run security checks"
	@echo "  help               - Show this help message"
	@echo ""
	@echo "Environment Variables:"
	@echo "  CONFIG_PATH        - Path to config file (default: config.yaml)"
	@echo "  LOG_LEVEL          - Log level (debug, info, warn, error)"
	@echo ""

# 安装开发工具
install-tools:
	@echo "Installing development tools..."
	go install github.com/cosmtrek/air@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/securego/gosec/v2/cmd/gosec@latest
	@echo "Development tools installed"

# 创建配置文件模板
create-config:
	@echo "Creating config.yaml from template..."
	cp config.yaml.example config.yaml
	@echo "Please edit config.yaml with your settings"

# 备份数据库
backup-db:
	@echo "Backing up database..."
	@read -p "Enter backup file name: " filename; \
	pg_dump -h localhost -U postgres -d owlrd > "backups/$$filename.sql"
	@echo "Database backup completed"

# 恢复数据库
restore-db:
	@echo "Restoring database..."
	@read -p "Enter backup file name: " filename; \
	psql -h localhost -U postgres -d owlrd < "backups/$$filename.sql"
	@echo "Database restore completed"

# 查看服务状态
status:
	@echo "Service Status:"
	@echo "---------------"
	@if [ -f wisefido-qinglan ]; then \
		echo "Application: Built"; \
	else \
		echo "Application: Not built"; \
	fi
	@echo ""
	@echo "Docker Containers:"
	@docker-compose ps
	@echo ""
	@echo "Network Ports:"
	@echo "  HTTP API:      http://localhost:8081"
	@echo "  MQTT Broker:   mqtt://localhost:1883"
	@echo "  PostgreSQL:    postgresql://localhost:5432"
	@echo "  Redis:         redis://localhost:6379"
	@echo "  PGAdmin:       http://localhost:5050"
	@echo "  RedisInsight:  http://localhost:5540"

# 性能监控
monitor:
	@echo "Starting performance monitoring..."
	@echo "Press Ctrl+C to stop"
	@echo ""
	@echo "CPU and Memory Usage:"
	@if [ -f wisefido-qinglan ]; then \
		pid=$$(pgrep -f wisefido-qinglan); \
		if [ -n "$$pid" ]; then \
			top -p $$pid; \
		else \
			echo "Application is not running"; \
		fi \
	else \
		echo "Application is not built"; \
	fi

# 压力测试
stress-test:
	@echo "Running stress test..."
	@echo "This will send 1000 requests to the health endpoint"
	@ab -n 1000 -c 10 http://localhost:8081/health

# 生成API文档
api-docs:
	@echo "Generating API documentation..."
	@echo "# wisefido-qinglan API Documentation" > API.md
	@echo "" >> API.md
	@echo "## Health Check" >> API.md
	@echo "" >> API.md
	@echo "### GET /health" >> API.md
	@echo "" >> API.md
	@echo "Returns service health status." >> API.md
	@echo "" >> API.md
	@echo "**Response:**" >> API.md
	@echo '```json' >> API.md
	@echo '{' >> API.md
	@echo '  "status": "healthy",' >> API.md
	@echo '  "service": "wisefido-qinglan",' >> API.md
	@echo '  "time": "2024-01-01T00:00:00Z"' >> API.md
	@echo '}' >> API.md
	@echo '```' >> API.md
	@echo "" >> API.md
	@echo "API documentation generated: API.md"
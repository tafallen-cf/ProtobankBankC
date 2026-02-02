# ProtobankBankC - Development Makefile
# Common commands for local development

.PHONY: help setup up down restart logs clean test build

# Default target
.DEFAULT_GOAL := help

# =============================================================================
# HELP
# =============================================================================

help: ## Show this help message
	@echo "ProtobankBankC - Development Commands"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

# =============================================================================
# SETUP
# =============================================================================

setup: ## Initial project setup (copy .env, make scripts executable)
	@echo "🚀 Setting up Protobank development environment..."
	@if [ ! -f .env ]; then \
		cp .env.example .env; \
		echo "✅ Created .env file from .env.example"; \
		echo "⚠️  Please review and update .env with your configuration"; \
	else \
		echo "ℹ️  .env file already exists"; \
	fi
	@chmod +x scripts/*.sh
	@echo "✅ Made scripts executable"
	@echo ""
	@echo "Next steps:"
	@echo "  1. Review .env file and update if needed"
	@echo "  2. Run 'make up' to start all services"
	@echo "  3. Run 'make logs' to view service logs"

init: setup ## Alias for setup

# =============================================================================
# DOCKER COMPOSE
# =============================================================================

up: ## Start all services
	@echo "🚀 Starting Protobank services..."
	docker-compose up -d
	@echo ""
	@echo "✅ All services started!"
	@echo ""
	@echo "📊 Service URLs:"
	@echo "  - Auth Service:         http://localhost:3001"
	@echo "  - User Service:         http://localhost:3002"
	@echo "  - Account Service:      http://localhost:3003"
	@echo "  - Transaction Service:  http://localhost:3004"
	@echo "  - Card Service:         http://localhost:3005"
	@echo "  - Payment Service:      http://localhost:3006"
	@echo "  - Notification Service: http://localhost:3007"
	@echo "  - Analytics Service:    http://localhost:3008"
	@echo ""
	@echo "🗄️  Infrastructure:"
	@echo "  - PostgreSQL:          localhost:5432"
	@echo "  - PgBouncer:           localhost:6432"
	@echo "  - Redis:               localhost:6379"
	@echo "  - RabbitMQ:            localhost:5672"
	@echo "  - RabbitMQ Management: http://localhost:15672 (admin/admin)"
	@echo ""
	@echo "Run 'make logs' to view logs, 'make down' to stop all services"

start: up ## Alias for up

down: ## Stop all services
	@echo "🛑 Stopping Protobank services..."
	docker-compose down
	@echo "✅ All services stopped"

stop: down ## Alias for down

restart: ## Restart all services
	@echo "🔄 Restarting Protobank services..."
	docker-compose restart
	@echo "✅ All services restarted"

rebuild: ## Rebuild and restart all services
	@echo "🔨 Rebuilding Protobank services..."
	docker-compose up -d --build
	@echo "✅ All services rebuilt and restarted"

# =============================================================================
# LOGS
# =============================================================================

logs: ## View logs from all services
	docker-compose logs -f

logs-auth: ## View logs from auth service
	docker-compose logs -f auth-service

logs-user: ## View logs from user service
	docker-compose logs -f user-service

logs-account: ## View logs from account service
	docker-compose logs -f account-service

logs-transaction: ## View logs from transaction service
	docker-compose logs -f transaction-service

logs-card: ## View logs from card service
	docker-compose logs -f card-service

logs-payment: ## View logs from payment service
	docker-compose logs -f payment-service

logs-notification: ## View logs from notification service
	docker-compose logs -f notification-service

logs-analytics: ## View logs from analytics service
	docker-compose logs -f analytics-service

logs-postgres: ## View logs from PostgreSQL
	docker-compose logs -f postgres

logs-redis: ## View logs from Redis
	docker-compose logs -f redis

logs-rabbitmq: ## View logs from RabbitMQ
	docker-compose logs -f rabbitmq

# =============================================================================
# DATABASE
# =============================================================================

db-connect: ## Connect to PostgreSQL database
	docker-compose exec postgres psql -U postgres -d protobank

db-reset: ## Reset database (WARNING: deletes all data!)
	@echo "⚠️  WARNING: This will delete ALL data in the database!"
	@read -p "Are you sure? (yes/no): " confirm && [ $$confirm = "yes" ] || exit 1
	@echo "🗑️  Dropping and recreating database..."
	docker-compose exec postgres psql -U postgres -c "DROP DATABASE IF EXISTS protobank;"
	docker-compose exec postgres psql -U postgres -c "CREATE DATABASE protobank;"
	docker-compose exec postgres psql -U postgres -d protobank -f /docker-entrypoint-initdb.d/01-schema.sql
	@echo "✅ Database reset complete"

db-backup: ## Backup database to backup.sql
	@echo "💾 Backing up database..."
	docker-compose exec -T postgres pg_dump -U postgres protobank > backup.sql
	@echo "✅ Database backed up to backup.sql"

db-restore: ## Restore database from backup.sql
	@echo "📥 Restoring database from backup.sql..."
	docker-compose exec -T postgres psql -U postgres protobank < backup.sql
	@echo "✅ Database restored from backup.sql"

# =============================================================================
# TESTING
# =============================================================================

test: ## Run all tests
	@echo "🧪 Running tests..."
	@cd backend/auth-service && go test -v -race ./...

test-unit: ## Run unit tests
	@echo "🧪 Running unit tests..."
	@cd backend/auth-service && go test -v -short ./...

test-integration: ## Run integration tests
	@echo "🧪 Running integration tests..."
	@cd backend/auth-service && go test -v -tags=integration ./tests/integration/...

test-coverage: ## Run tests with coverage
	@echo "📊 Running tests with coverage..."
	@cd backend/auth-service && go test -race -coverprofile=coverage.out -covermode=atomic ./...
	@cd backend/auth-service && go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report generated: backend/auth-service/coverage.html"

# =============================================================================
# SECURITY
# =============================================================================

security-scan: ## Run security scans
	@echo "🔒 Running security scans..."
	@cd backend/auth-service && ../../scripts/security-scan.sh

lint: ## Run linter
	@echo "🔍 Running linter..."
	@cd backend/auth-service && golangci-lint run --timeout=5m

fmt: ## Format code
	@echo "✨ Formatting code..."
	@cd backend/auth-service && go fmt ./...
	@cd backend/auth-service && goimports -w .

vet: ## Run go vet
	@echo "🔍 Running go vet..."
	@cd backend/auth-service && go vet ./...

# =============================================================================
# CLEANUP
# =============================================================================

clean: ## Remove containers, volumes, and orphans
	@echo "🧹 Cleaning up..."
	docker-compose down -v --remove-orphans
	@echo "✅ Cleanup complete"

clean-all: clean ## Remove everything including images
	@echo "🧹 Removing all images..."
	docker-compose down -v --rmi all --remove-orphans
	@echo "✅ Deep cleanup complete"

prune: ## Prune Docker system (careful!)
	@echo "⚠️  WARNING: This will remove all unused Docker resources system-wide!"
	@read -p "Are you sure? (yes/no): " confirm && [ $$confirm = "yes" ] || exit 1
	docker system prune -af --volumes
	@echo "✅ Docker system pruned"

# =============================================================================
# STATUS
# =============================================================================

status: ## Show status of all services
	@echo "📊 Service Status:"
	@docker-compose ps

ps: status ## Alias for status

health: ## Check health of all services
	@echo "🏥 Health Check:"
	@echo ""
	@echo "PostgreSQL:"
	@docker-compose exec postgres pg_isready -U postgres -d protobank && echo "  ✅ Healthy" || echo "  ❌ Unhealthy"
	@echo ""
	@echo "Redis:"
	@docker-compose exec redis redis-cli --pass redis ping >/dev/null 2>&1 && echo "  ✅ Healthy" || echo "  ❌ Unhealthy"
	@echo ""
	@echo "RabbitMQ:"
	@docker-compose exec rabbitmq rabbitmq-diagnostics ping >/dev/null 2>&1 && echo "  ✅ Healthy" || echo "  ❌ Unhealthy"

# =============================================================================
# DEVELOPMENT
# =============================================================================

shell-postgres: ## Open shell in PostgreSQL container
	docker-compose exec postgres sh

shell-redis: ## Open shell in Redis container
	docker-compose exec redis sh

shell-rabbitmq: ## Open shell in RabbitMQ container
	docker-compose exec rabbitmq sh

redis-cli: ## Connect to Redis CLI
	docker-compose exec redis redis-cli -a redis

# =============================================================================
# CODE GENERATION (Future)
# =============================================================================

generate: ## Generate code (protobuf, swagger, etc.)
	@echo "⚙️  Code generation not yet implemented"

swagger: ## Generate Swagger documentation
	@echo "📚 Swagger generation not yet implemented"

# =============================================================================
# GIT
# =============================================================================

git-status: ## Show git status
	@git status

git-log: ## Show git log
	@git log --oneline -10

# =============================================================================
# INFO
# =============================================================================

info: ## Show project information
	@echo "ProtobankBankC - Monzo Clone Banking Application"
	@echo ""
	@echo "Version: 1.0.0"
	@echo "Tech Stack: Go, PostgreSQL, Redis, RabbitMQ"
	@echo "Architecture: Microservices"
	@echo ""
	@echo "Documentation:"
	@echo "  - README.md"
	@echo "  - ARCHITECTURE.md"
	@echo "  - API_SPECIFICATION.md"
	@echo "  - TECH_STACK.md"
	@echo ""
	@echo "Run 'make help' for available commands"

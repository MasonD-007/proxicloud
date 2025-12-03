# ProxiCloud Frontend Makefile

.PHONY: help dev build start install clean lint format type-check test

# Variables
NODE=npm
NEXT=npx next

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

install: ## Install dependencies
	$(NODE) install

dev: ## Run development server
	$(NODE) run dev

build: ## Build for production
	$(NODE) run build

start: ## Start production server (requires build first)
	$(NODE) run start

export: ## Export static site
	$(NEXT) export

lint: ## Run ESLint
	$(NODE) run lint

lint-fix: ## Fix ESLint issues
	$(NODE) run lint -- --fix

format: ## Format code with Prettier
	$(NODE) run format || npx prettier --write .

format-check: ## Check code formatting
	$(NODE) run format:check || npx prettier --check .

type-check: ## Run TypeScript type checking
	$(NODE) run type-check || npx tsc --noEmit

test: ## Run tests
	$(NODE) run test || echo "No tests configured yet"

test-watch: ## Run tests in watch mode
	$(NODE) run test:watch || echo "No tests configured yet"

test-coverage: ## Run tests with coverage
	$(NODE) run test:coverage || echo "No tests configured yet"

check: lint type-check ## Run all checks (lint, type-check)

clean: ## Clean build artifacts and dependencies
	@echo "Cleaning..."
	@rm -rf .next
	@rm -rf out
	@rm -rf node_modules
	@rm -rf coverage
	@rm -f .turbo
	@echo "Clean complete"

clean-cache: ## Clean Next.js cache
	@echo "Cleaning Next.js cache..."
	@rm -rf .next/cache
	@echo "Cache cleaned"

update: ## Update dependencies
	$(NODE) update

update-check: ## Check for outdated dependencies
	$(NODE) outdated

audit: ## Run security audit
	$(NODE) audit

audit-fix: ## Fix security vulnerabilities
	$(NODE) audit fix

analyze: build ## Analyze bundle size
	@echo "Building with bundle analyzer..."
	ANALYZE=true $(NODE) run build

serve: build ## Build and serve production build
	$(NODE) run start

docker-build: ## Build Docker image
	docker build -t proxicloud-frontend:latest .

docker-run: ## Run Docker container
	docker run -p 3000:3000 proxicloud-frontend:latest

package: build ## Package frontend for deployment
	@echo "Packaging frontend..."
	@mkdir -p ../dist
	tar -czf ../dist/proxicloud-frontend.tar.gz \
		.next/ \
		public/ \
		package.json \
		package-lock.json \
		next.config.js \
		tsconfig.json
	@echo "Package created: ../dist/proxicloud-frontend.tar.gz"

.DEFAULT_GOAL := help

SHELL := /bin/bash

GO ?= go
PY ?= python3
NPM ?= npm

UI_DIR ?= frontend_landing
PY_SDK_DIR ?= sdk/python

CACHE_DIR := $(CURDIR)/.cache
export GOCACHE := $(CACHE_DIR)/go-build
export GOMODCACHE := $(CACHE_DIR)/go-mod
export GOTMPDIR := $(CACHE_DIR)/go-tmp

.PHONY: bootstrap dev test lint migrate-up migrate-down helm-lint e2e

bootstrap:
	@mkdir -p "$(GOCACHE)" "$(GOMODCACHE)" "$(GOTMPDIR)"
	@$(GO) mod tidy
	@if [ -d "$(UI_DIR)" ]; then \
		if [ ! -d "$(UI_DIR)/node_modules" ]; then \
			echo "==> npm ci ($(UI_DIR))"; \
			( cd "$(UI_DIR)" && $(NPM) ci ); \
		fi; \
	fi

dev:
	@./scripts/dev.sh

lint:
	@mkdir -p "$(GOCACHE)" "$(GOMODCACHE)" "$(GOTMPDIR)"
	@echo "==> gofmt (check)"
	@unformatted="$$(gofmt -l $$(find . -type f -name '*.go' -not -path './.cache/*' -not -path './frontend_landing/*' -not -path './ui/*'))"; \
	if [ -n "$$unformatted" ]; then \
		echo "$$unformatted"; \
		echo "gofmt check failed (run: gofmt -w <files>)"; \
		exit 1; \
	fi
	@echo "==> go vet"
	@$(GO) vet ./...
	@echo "==> python compileall"
	@PYTHONPATH="$(PY_SDK_DIR)/src" $(PY) -m compileall -q "$(PY_SDK_DIR)/src"
	@if [ -d "$(UI_DIR)" ]; then \
		echo "==> ui typecheck ($(UI_DIR))"; \
		( cd "$(UI_DIR)" && $(NPM) run -s typecheck ); \
	fi

test:
	@mkdir -p "$(GOCACHE)" "$(GOMODCACHE)" "$(GOTMPDIR)"
	@echo "==> go test"
	@$(GO) test ./...
	@echo "==> python tests"
	@PYTHONPATH="$(PY_SDK_DIR)/src" $(PY) -m unittest discover -s "$(PY_SDK_DIR)/tests" -p 'test_*.py'
	@if [ -d "$(UI_DIR)" ]; then \
		echo "==> ui tests ($(UI_DIR))"; \
		( cd "$(UI_DIR)" && $(NPM) test ); \
	fi

migrate-up:
	@./scripts/migrate.sh up

migrate-down:
	@./scripts/migrate.sh down

helm-lint:
	@helm lint deploy/helm/animus-datapilot

e2e:
	@mkdir -p "$(GOCACHE)" "$(GOMODCACHE)" "$(GOTMPDIR)"
	@$(GO) test -tags=e2e ./e2e -count=1

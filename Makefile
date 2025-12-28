SHELL := /bin/bash

GO ?= go
PY ?= python3

PY_SDK_DIR ?= open/sdk/python
GO_PACKAGES ?= ./open/...

CACHE_DIR := $(CURDIR)/.cache
export GOCACHE := $(CACHE_DIR)/go-build
export GOMODCACHE := $(CACHE_DIR)/go-mod
export GOTMPDIR := $(CACHE_DIR)/go-tmp

.PHONY: bootstrap test lint

bootstrap:
	@mkdir -p "$(GOCACHE)" "$(GOMODCACHE)" "$(GOTMPDIR)"
	@$(GO) mod tidy

lint:
	@mkdir -p "$(GOCACHE)" "$(GOMODCACHE)" "$(GOTMPDIR)"
	@echo "==> gofmt (check)"
	@unformatted="$$(gofmt -l $$(find open -type f -name '*.go' -not -path './.cache/*'))"; \
	if [ -n "$$unformatted" ]; then \
		echo "$$unformatted"; \
		echo "gofmt check failed (run: gofmt -w <files>)"; \
		exit 1; \
	fi
	@echo "==> go vet"
	@$(GO) vet $(GO_PACKAGES)
	@echo "==> python compileall"
	@PYTHONPATH="$(PY_SDK_DIR)/src" $(PY) -m compileall -q "$(PY_SDK_DIR)/src"

test:
	@mkdir -p "$(GOCACHE)" "$(GOMODCACHE)" "$(GOTMPDIR)"
	@echo "==> go test"
	@$(GO) test $(GO_PACKAGES)
	@echo "==> python tests"
	@PYTHONPATH="$(PY_SDK_DIR)/src" $(PY) -m unittest discover -s "$(PY_SDK_DIR)/tests" -p 'test_*.py'

# kamonu — build / run / test / package
#
# 跑 `make`（或 `make help`）列出所有指令。
# unix-first（macOS + Linux），CGO_ENABLED=0 靜態編譯。

BINARY   := kamonu
PKG      := .
DIST_DIR := dist

GOOS   := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w

.DEFAULT_GOAL := help

##@ 編譯（build）

.PHONY: build
build: ## 編譯本地執行檔 → ./kamonu（CGO_ENABLED=0 靜態、-trimpath、strip）
	CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -o $(BINARY) $(PKG)
	@echo "built ./$(BINARY)  ($(VERSION) $(GOOS)/$(GOARCH))"

.PHONY: install
install: ## go install → $$GOBIN/PATH
	CGO_ENABLED=0 go install -trimpath -ldflags "$(LDFLAGS)" $(PKG)

.PHONY: uninstall
uninstall: ## 移除已安裝的 kamonu（從 $$GOBIN，否則 $$GOPATH/bin）
	@dir="$$(go env GOBIN)"; [ -n "$$dir" ] || dir="$$(go env GOPATH)/bin"; \
	if [ -f "$$dir/$(BINARY)" ]; then rm -f "$$dir/$(BINARY)" && echo "removed $$dir/$(BINARY)"; else echo "not installed ($$dir/$(BINARY))"; fi

.PHONY: tidy
tidy: ## go mod tidy（關掉 go.work，照 go.mod 寫的 terminu-go tag）
	GOWORK=off go mod tidy

##@ 發布（release）

.PHONY: release-check
release-check: ## goreleaser check：驗證 .goreleaser.yaml
	goreleaser check

.PHONY: snapshot
snapshot: ## goreleaser 本機 snapshot 打包（不發布、不推 tap）→ dist/
	GOWORK=off goreleaser release --snapshot --clean

##@ 執行（run）

.PHONY: run
run: ## 本地跑（讀真正的 ~/.kamonu）
	go run $(PKG)

##@ 測試 / 檢查（test）

.PHONY: test
test: ## 跑所有測試（有 go.work 時用本機的 ../terminu-go）
	go test ./...

.PHONY: test-race
test-race: ## 帶 race detector 跑測試（check 也跑它）
	go test -race ./...

.PHONY: check-lib
check-lib: ## 關掉 go.work，用 go.mod 寫的 terminu-go tag 編譯、測試（過不了：用到了 lib 還沒發版的東西）
	GOWORK=off go build ./...
	GOWORK=off go test ./...

.PHONY: vet
vet: ## go vet ./...
	go vet ./...

.PHONY: fmt
fmt: ## gofmt -w（就地格式化）
	gofmt -w .

.PHONY: fmt-check
fmt-check: ## gofmt -l（列出未格式化的檔；有輸出即失敗，CI 用）
	@out=$$(gofmt -l .); if [ -n "$$out" ]; then echo "not gofmt'ed:"; echo "$$out"; exit 1; fi

.PHONY: check
check: fmt-check vet test-race check-lib ## fmt-check + vet + test -race + check-lib 一次跑（commit 前）

##@ 打包（package）

.PHONY: package
package: build ## 打包 → dist/kamonu_<ver>_<os>_<arch>.tar.gz
	@mkdir -p $(DIST_DIR)
	tar -czf $(DIST_DIR)/$(BINARY)_$(VERSION)_$(GOOS)_$(GOARCH).tar.gz $(BINARY)
	@echo "packaged $(DIST_DIR)/$(BINARY)_$(VERSION)_$(GOOS)_$(GOARCH).tar.gz"

##@ 其他

.PHONY: clean
clean: ## 移除 ./kamonu 與 dist/
	rm -f $(BINARY)
	rm -rf $(DIST_DIR)

.PHONY: help
help: ## 顯示這份說明
	@awk 'BEGIN {FS = ":.*?## "} \
		/^##@/ {printf "\n\033[1m%s\033[0m\n", substr($$0, 5); next} \
		/^[a-zA-Z0-9_-]+:.*?## / {printf "  \033[36m%-14s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

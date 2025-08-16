MODULE := $(shell go list -m)
MODULE_NAME := $(lastword $(subst /, ,$(MODULE)))
GO_BUILD := CGO_ENABLED=0 go build -trimpath

.PHONY: build-linux-amd
build-linux-amd: ## Build linux amd64 binary
	GOOS=linux GOARCH=amd64 $(GO_BUILD) -o ./build/linux_x86/ ./...

.PHONY: build-linux-arm
build-linux-arm: ## Build linux arm64 binary
	GOOS=linux GOARCH=arm64 $(GO_BUILD) -o ./build/linux_arm64/ ./...

.PHONY: release
release: build-linux-amd build-linux-arm ## Build release tarball
	tar -czf ./build/$(MODULE_NAME)_linux_x86.tar.gz -C ./build/linux_x86/ .
	tar -czf ./build/$(MODULE_NAME)_linux_arm64.tar.gz -C ./build/linux_arm64/ .

.PHONY: fmt
fmt:
	go test ./...
	go fmt ./...
	golangci-lint run --fix
	golangci-lint fmt
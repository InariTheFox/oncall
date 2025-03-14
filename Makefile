WIRE_TAGS = "oss"

GO = go
GO_VERSION = 1.24.1
GO_RACE  := $(shell [ -n "$(GO_RACE)" -o -e ".go-race-enabled-locally" ] && echo 1 )
GO_RACE_FLAG := $(if $(GO_RACE),-race)
GO_BUILD_FLAGS += $(if $(GO_BUILD_DEV),-dev)
GO_BUILD_FLAGS += $(if $(GO_BUILD_TAGS),-build-tags=$(GO_BUILD_TAGS))
GO_BUILD_FLAGS += $(GO_RACE_FLAG)

targets := $(shell echo '$(sources)' | tr "," " ")

.PHONY: all
all: deps build

.PHONY: deps
deps:

.PHONY: deps-go
deps-go:
	$(GO) run $(GO_RACE_FLAG) build.go setup

.PHONY: gen-go
gen-go:
	@echo "generate go files"
	$(GO) run $(GO_RACE_FLAG) wire gen -tags $(WIRE_TAGS) ./pkg/server

.PHONY: build-go
build-go: gen-go update-workspace ## Build all Go binaries.
	@echo "build go files with updated workspace"
	$(GO) run build.go $(GO_BUILD_FLAGS) build

.PHONY: build-server
build-server: ## Build OnCall server.
	@echo "build server"
	$(GO) run build.go $(GO_BUILD_FLAGS) build-server

.PHONY: build-cli
build-cli: ## Build OnCall CLI application.
	@echo "build oncall-cli"
	$(GO) run build.go $(GO_BUILD_FLAGS) build-cli

.PHONY: build
build: build-go
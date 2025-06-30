
# Default target
.PHONY: all
all: deps binary

# Version information
VERSION := $(shell cat VERSION)
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -X 'github.com/panyam/sdl/cmd/sdl/commands.Version=$(VERSION)' \
           -X 'github.com/panyam/sdl/cmd/sdl/commands.GitCommit=$(GIT_COMMIT)' \
           -X 'github.com/panyam/sdl/cmd/sdl/commands.BuildDate=$(BUILD_DATE)'

# Build targets
binary: 
	cd parser && make
	cd wasm && make
	go build -ldflags "$(LDFLAGS)" -o ${GOBIN}/sdl ./cmd/sdl/main.go

binlocal: 
	cd parser && make
	go build -ldflags "$(LDFLAGS)" -o /tmp/sdl ./cmd/sdl/main.go

# Installation targets
.PHONY: deps check-deps install-tools install

# Check prerequisites
check-deps:
	@echo "Checking prerequisites..."
	@command -v go >/dev/null 2>&1 || { echo "Error: Go is not installed. Visit https://golang.org/dl/"; exit 1; }
	@go version | grep -q "go1\.2[4-9]" || go version | grep -q "go1\.[3-9]" || { echo "Error: Go 1.24+ required"; exit 1; }
	@command -v node >/dev/null 2>&1 || { echo "Error: Node.js is not installed. Visit https://nodejs.org/"; exit 1; }
	@node --version | grep -q "v1[8-9]\|v[2-9]" || { echo "Error: Node.js 18+ required"; exit 1; }
	@command -v npm >/dev/null 2>&1 || { echo "Error: npm is not installed"; exit 1; }
	@command -v git >/dev/null 2>&1 || { echo "Error: git is not installed"; exit 1; }
	@command -v make >/dev/null 2>&1 || { echo "Error: make is not installed"; exit 1; }
	@echo "✓ All prerequisites found"

# Install required Go tools
install-tools:
	@echo "Installing required Go tools..."
	go install golang.org/x/tools/cmd/goyacc@latest
	go install github.com/bufbuild/buf/cmd/buf@latest
	@echo "✓ Go tools installed"

# Install Node dependencies
install-npm:
	@echo "Installing Node.js dependencies..."
	cd web && npm install
	@echo "✓ Node.js dependencies installed"

# Install everything needed for development
deps: check-deps install-tools install-npm
	@echo "✓ All dependencies installed successfully!"

# Full installation (deps + build)
install: deps all
	@echo "✓ SDL installed successfully!"
	@echo "Run 'sdl --help' to get started"

# Display version
.PHONY: version
version:
	@echo "SDL version: $(VERSION)"
	@echo "Git commit: $(GIT_COMMIT)"
	@echo "Build will use: $(LDFLAGS)"

# Release management
.PHONY: release release-patch release-minor release-major

# Helper function to bump version
define bump_version
	$(eval CURRENT_VERSION := $(shell cat VERSION))
	$(eval VERSION_PARTS := $(subst ., ,$(CURRENT_VERSION)))
	$(eval MAJOR := $(word 1,$(VERSION_PARTS)))
	$(eval MINOR := $(word 2,$(VERSION_PARTS)))
	$(eval PATCH := $(word 3,$(VERSION_PARTS)))
endef

release-patch:
	$(call bump_version)
	$(eval NEW_PATCH := $(shell echo $$(($(PATCH) + 1))))
	$(eval NEW_VERSION := $(MAJOR).$(MINOR).$(NEW_PATCH))
	@$(MAKE) release VERSION=$(NEW_VERSION)

release-minor:
	$(call bump_version)
	$(eval NEW_MINOR := $(shell echo $$(($(MINOR) + 1))))
	$(eval NEW_VERSION := $(MAJOR).$(NEW_MINOR).0)
	@$(MAKE) release VERSION=$(NEW_VERSION)

release-major:
	$(call bump_version)
	$(eval NEW_MAJOR := $(shell echo $$(($(MAJOR) + 1))))
	$(eval NEW_VERSION := $(NEW_MAJOR).0.0)
	@$(MAKE) release VERSION=$(NEW_VERSION)

release:
	@if [ -z "$(VERSION)" ]; then \
		echo "Error: VERSION not specified"; \
		echo "Usage: make release VERSION=1.0.0"; \
		echo "   or: make release-patch|release-minor|release-major"; \
		exit 1; \
	fi
	@echo "Preparing release for SDL v$(VERSION)"
	@echo "$(VERSION)" > VERSION
	@echo "✓ Updated VERSION file"
	@if [ -f CHANGELOG.md ]; then \
		DATE=$$(date +"%Y-%m-%d"); \
		echo "## Version $(VERSION) - $$DATE\n" > CHANGELOG.tmp; \
		cat CHANGELOG.md >> CHANGELOG.tmp; \
		mv CHANGELOG.tmp CHANGELOG.md; \
		echo "✓ Updated CHANGELOG.md"; \
	fi
	@git add VERSION CHANGELOG.md
	@git commit -m "Release v$(VERSION)" || true
	@git tag -a "v$(VERSION)" -m "Release version $(VERSION)"
	@echo "✓ Created git tag v$(VERSION)"
	@echo ""
	@echo "Release preparation complete!"
	@echo ""
	@echo "Next steps:"
	@echo "1. Push changes: git push origin main"
	@echo "2. Push tag: git push origin v$(VERSION)"
	@echo "3. Build release: make clean && make"

buf:
	buf generate

reload: buf

dash:
	cd web && npm run build

# Development workflow: build and test dashboard
dev-test: binary
	cd web && ./dev-test.sh

# Quick development validation
dev-quick: binary
	cd web && npm run dev-quick

dev-screenshot: binary
	cd web && npm run dev-screenshot

run:
	go test

test:
	go test ./...

bench:
	cd core && go test -bench=Benchmark -benchmem

watch:
	while true; do clear	; make run ; fswatch  -o ../ | echo "Files changed, re-testing..."; sleep 1 ; done

testall: test bench

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	rm -f ${GOBIN}/sdl
	rm -rf web/dist
	rm -rf web/node_modules
	rm -rf node_modules
	cd parser && make clean || true
	@echo "✓ Clean complete"

# Deep clean (including dependencies)
.PHONY: distclean
distclean: clean
	@echo "Removing all downloaded dependencies..."
	go clean -modcache
	@echo "✓ Deep clean complete"

sdlfiles:
	@find . | grep -v "\.git" | grep -v "\.sh" | grep -v "\..decl" | grep -v attic | grep -v prompt | grep -v vscode | grep -v dsl | grep -v _test.go | grep -v "\.bak" | grep -v debug | grep -v "\.output" | grep -v "\.svg" | grep -v "\.png" | grep -v parser.go | grep -v CLAUDE | grep -v node_modules | grep -v "\.\/sdl" | grep -v package | grep -v playwright-report | grep -v test-results | grep -v dist | grep -v sdl

sdlprompt:
	source ~/personal/.shhelpers && files_for_llm `make sdlfiles`

parserfiles:
	@find . |  grep -v "\.git" | grep -v "\.bak" | grep -e "decl" -e "parser" | grep -v ll | grep -v y.output | grep -v parser.go  | grep -v debug | grep -v "\.output"

parserprompt:
	source ~/personal/.shhelpers && files_for_llm `make parserfiles`

newchatprompt:
	cat prompts/STARTER.prompt && make sdlprompt

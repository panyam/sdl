# SystemDetailTool WASM Build
.PHONY: wasm clean test

# Build WASM binary for SystemDetailTool
wasm:
	@echo "Building SystemDetailTool WASM binary..."
	@mkdir -p ../../web/dist/wasm
	GOOS=js GOARCH=wasm go build -o ../../web/dist/wasm/systemdetail.wasm ./wasm/
	@echo "✅ SystemDetailTool WASM binary built: ../../web/dist/wasm/systemdetail.wasm"

# Clean build artifacts
clean:
	@echo "Cleaning SystemDetailTool build artifacts..."
	@rm -f ../../web/dist/wasm/systemdetail.wasm
	@echo "✅ Clean complete"

# Run tests
test:
	@echo "Running SystemDetailTool tests..."
	go test -v .
	go test -v ../shared/recipe
	@echo "✅ Tests complete"

# Build for development (includes tests)
dev: test wasm
	@echo "✅ Development build complete"

# Build for production
prod: wasm
	@echo "✅ Production build complete"
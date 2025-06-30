#!/bin/bash

# Build SDL WASM module

echo "Building SDL WASM..."

# Set WASM build environment
export GOOS=js
export GOARCH=wasm

# Build the WASM binary
mkdir -p ../web/dist
go build -o ../web/dist/sdl.wasm ./cmd/

if [ "$?" != "0" ]; then
  echo "Build failed..."
  exit 1
fi

# Copy the Go WASM support file
cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" ../web/dist/

echo "Build complete! Output:"
echo "  - web/sdl.wasm"
echo "  - web/wasm_exec.js"
echo ""
echo "To test locally:"
echo "  1. cd web"
echo "  2. python3 -m http.server 8080"
echo "  3. Open http://localhost:8080"

#!/bin/bash

# Development inner-loop script for SDL Dashboard testing
# Assumes server is running on port 8080

echo "🚀 Starting SDL Dashboard Development Test Loop"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Check if server is running
echo "📋 Checking if server is running on port 8080..."
if ! curl -s http://localhost:8080 > /dev/null; then
    echo "❌ Server not running on port 8080. Please start with: sdl serve --port 8080"
    exit 1
fi
echo "✅ Server is running"

# Run the development test
echo "🧪 Running development loop test..."
npx playwright test development-loop.spec.ts --reporter=line

# Show screenshots if they exist
echo ""
echo "📸 Screenshots generated:"
ls -la tests/screenshots/dev-loop-*.png 2>/dev/null || echo "No screenshots found"
ls -la tests/screenshots/dev-quick-*.png 2>/dev/null || echo "No quick screenshots found"

echo ""
echo "🎯 Quick test options:"
echo "  npm run dev-test         # Run full development test"
echo "  npm run dev-quick        # Run quick validation only"
echo "  npm run dev-screenshot   # Just take a screenshot"
echo ""
echo "🔧 For iterative development:"
echo "  1. Make changes to dashboard.ts or canvas-web.go"
echo "  2. Run: make && ./dev-test.sh"
echo "  3. Check screenshots in tests/screenshots/"
echo "  4. Repeat"
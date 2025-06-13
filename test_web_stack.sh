#!/bin/bash

echo "🧪 Testing SDL Web Stack"
echo "========================="

# Build the frontend
echo "📦 Building frontend..."
cd web && npm run build
if [ $? -ne 0 ]; then
    echo "❌ Frontend build failed"
    exit 1
fi
cd ..

# Build the backend  
echo "🔨 Building backend..."
go build -o sdl ./cmd/sdl
if [ $? -ne 0 ]; then
    echo "❌ Backend build failed"
    exit 1
fi

# Start server in background
echo "🚀 Starting server..."
./sdl serve --port 8080 &
SERVER_PID=$!

# Wait for server to start
sleep 2

# Test API endpoints
echo "🔍 Testing API endpoints..."

# Test Load endpoint
echo "Testing /api/load..."
LOAD_RESULT=$(curl -s -X POST http://localhost:8080/api/load \
  -H "Content-Type: application/json" \
  -d '{"filePath": "examples/contacts/contacts.sdl"}')

echo "Load result: $LOAD_RESULT"

if echo "$LOAD_RESULT" | grep -q '"success":true'; then
    echo "✅ Load API test passed"
else
    echo "❌ Load API test failed"
fi

# Test Use endpoint  
echo "Testing /api/use..."
USE_RESULT=$(curl -s -X POST http://localhost:8080/api/use \
  -H "Content-Type: application/json" \
  -d '{"systemName": "ContactsSystem"}')

echo "Use result: $USE_RESULT"

if echo "$USE_RESULT" | grep -q '"success":true'; then
    echo "✅ Use API test passed"
else
    echo "❌ Use API test failed"
fi

# Test Set endpoint
echo "Testing /api/set..."
SET_RESULT=$(curl -s -X POST http://localhost:8080/api/set \
  -H "Content-Type: application/json" \
  -d '{"path": "server.pool.ArrivalRate", "value": 8.0}')

echo "Set result: $SET_RESULT"

if echo "$SET_RESULT" | grep -q '"success":true'; then
    echo "✅ Set API test passed"
else
    echo "❌ Set API test failed"
fi

# Test Run endpoint
echo "Testing /api/run..."
RUN_RESULT=$(curl -s -X POST http://localhost:8080/api/run \
  -H "Content-Type: application/json" \
  -d '{"varName": "test", "target": "server.HandleLookup", "runs": 100}')

echo "Run result: $RUN_RESULT"

if echo "$RUN_RESULT" | grep -q '"success":true'; then
    echo "✅ Run API test passed"
else
    echo "❌ Run API test failed"
fi

# Test static file serving
echo "🌐 Testing static file serving..."
if curl -s http://localhost:8080/ | grep -q "SDL Canvas Dashboard"; then
    echo "✅ Static file serving test passed"
else
    echo "❌ Static file serving test failed"
fi

# Clean up
echo "🧹 Cleaning up..."
kill $SERVER_PID 2>/dev/null

echo ""
echo "🎉 Web stack testing complete!"
echo ""
echo "To start the dashboard:"
echo "  ./sdl serve --port 8080"
echo "  Open http://localhost:8080 in your browser"
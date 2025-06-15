# SDL Client-Server Architecture

## Overview
The SDL project uses a clean client-server architecture that separates the REPL console from server logs for the best user experience.

## Architecture

### Server (`sdl serve`)
- Hosts the Canvas simulation engine
- Runs web dashboard on port 8080
- Shows traffic generator and measurement logs
- Displays server statistics
- Provides REST API for all Canvas operations

### Console (`sdl console`)
- Pure REPL interface without server logs
- Connects to Canvas server via REST API
- Clean terminal experience
- Maintains command history and tab completion

## Testing Instructions

### 1. Start Server (Terminal 1)
```bash
# Start SDL server with default settings
./bin/sdl serve

# Or with custom options
./bin/sdl serve --port 9090 --no-logs --no-stats
```

Expected output:
```
🚀 SDL Canvas Server v1.0
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📊 Dashboard:    http://localhost:8080
🛠️  REST API:     http://localhost:8080/api/canvas
📡 WebSocket:    ws://localhost:8080/api/live
💻 Console:      sdl console --server http://localhost:8080
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

2025/06/15 09:45:00 ✅ Server started successfully on port 8080
2025/06/15 09:45:00 📝 Logging enabled (use --no-logs to disable)
2025/06/15 09:45:00 📈 Statistics display enabled (updates every 5s)
```

### 2. Connect Console (Terminal 2)
```bash
# Connect console client to server
./bin/sdl console
```

Expected output:
```
🔌 SDL Console Client v1.0
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
🎯 Server:       http://localhost:8080
📊 Dashboard:    http://localhost:8080 (open in browser)
💬 Type 'help' for available commands, Ctrl+D to quit
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

✅ Connected to SDL server

SDL> 
```

### 2b. What Happens Without Server (User Guidance)
If you run `./bin/sdl console` without starting the server first:

```
🔌 SDL Console Client v1.0
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
🎯 Server:       http://localhost:8080
❌ Cannot connect to SDL server at http://localhost:8080

To use SDL console, first start the server:

🚀 Terminal 1: Start SDL server
   sdl serve

🔌 Terminal 2: Connect console client
   sdl console

Or connect to a different server:
   sdl console --server http://other-host:8080

💡 The server hosts the Canvas engine, web dashboard, and logs.
   The console provides a clean REPL experience.
```

### 3. Test Commands
In the client terminal:
```bash
SDL> load ./examples/contacts/contacts.sdl
✅ Loaded: ./examples/contacts/contacts.sdl

SDL> use ContactsSystem
✅ System activated: ContactsSystem

SDL> set server.pool.ArrivalRate 15
✅ Set server.pool.ArrivalRate = 15

SDL> run test1 server.HandleLookup 100
✅ Simulation completed: 100 runs of server.HandleLookup
```

### 4. Observe Server Logs
In the server terminal, you should see logs like:
```
2025/06/15 09:45:05 📊 Stats: Files=1 Systems=1 Generators=0 Measurements=0 Runs=1
```

### 5. Remote Server Connection
You can connect to remote SDL servers:
```bash
# Connect to remote server
./bin/sdl console --server http://production-server:8080
```

## API Endpoints

### Console Commands
- `POST /api/console/load` - Load SDL file
- `POST /api/console/use` - Activate system
- `POST /api/console/set` - Set parameter value
- `POST /api/console/run` - Run simulation
- `GET /api/console/help` - Get help text
- `GET /api/console/state` - Get Canvas state

### Canvas Management
- `GET /api/canvas/state` - Canvas state management
- `POST /api/canvas/generators` - Traffic generation
- `POST /api/canvas/measurements` - Measurement management

## Benefits

1. **Clean REPL Experience**: No server logs cluttering the console
2. **Remote Access**: Console can connect to servers on different machines
3. **Multiple Clients**: Multiple console clients can connect to same server
4. **Server Monitoring**: Server shows dedicated logs and statistics
5. **Scalability**: Server can handle web dashboard and multiple API clients
6. **Simplified Architecture**: Only one way to use SDL - clean and consistent

## Troubleshooting

### Client Cannot Connect
```bash
❌ Cannot connect to server: Get "http://localhost:8080/api/console/help": dial tcp [::1]:8080: connect: connection refused
💡 Make sure the server is running: sdl serve
```

Solution: Start the server first with `sdl serve`

### API Errors
If commands fail in client mode, check the server logs for detailed error messages.
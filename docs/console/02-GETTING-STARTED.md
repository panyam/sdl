# SDL Console & Server Tutorial - Getting Started

This chapter walks you through launching your first SDL server and console session.

## Step 1: Start the SDL Server

Open your first terminal and start the SDL server:

```bash
cd /path/to/sdl
./bin/sdl serve
```

You should see output like:

```
🚀 SDL Server v1.0
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
🎯 Canvas Engine: Ready
📊 Web Dashboard: http://localhost:8080
🔌 REST API:      http://localhost:8080/api
📈 WebSocket:     ws://localhost:8080/ws
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

✅ SDL server started successfully
💡 Keep this terminal open - server logs will appear here
🌐 Open http://localhost:8080 in your browser for the dashboard

Canvas Statistics (updates every 30s):
🎯 Systems Loaded: 0
⚡ Active Generators: 0  
📊 Active Measurements: 0
```

**Important:** Keep this terminal open! The server will show logs, statistics, and traffic generator activity here.

## Step 2: Open the Web Dashboard (Optional)

Open your web browser and navigate to:
```
http://localhost:8080
```

You'll see the SDL dashboard with empty charts initially. As you work with the console, data will appear here in real-time.

## Step 3: Start the Console Client

Open a second terminal and start the console client:

```bash
cd /path/to/sdl
./bin/sdl console
```

You should see:

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

## Step 4: Test Basic Connectivity

Try a simple command to verify everything is working:

```
SDL> help
```

You should see the help output listing available commands.

## Step 5: First SDL File Load

Load an example SDL file:

```
SDL> load examples/contacts/contacts.sdl
✅ Loaded examples/contacts/contacts.sdl successfully
Available systems: ContactsSystem
```

Notice how:
- The console stays clean with just the response
- The server terminal shows detailed loading logs
- The web dashboard updates to show the loaded system

## Architecture in Action

Now you can see the architecture working:

1. **Console Terminal** - Clean REPL interface, no clutter
2. **Server Terminal** - Detailed logs and statistics  
3. **Web Dashboard** - Real-time visualization

## What's Next?

Now that you have a working SDL session, continue to **[Basic Commands](03-BASIC-COMMANDS.md)** to learn the core SDL operations.

## Quick Reference

```bash
# Terminal 1: Start server
./bin/sdl serve

# Terminal 2: Start console  
./bin/sdl console

# In console: Load and explore
SDL> load examples/contacts/contacts.sdl
SDL> help
SDL> exit
```

## Troubleshooting

**"Cannot connect to SDL server"**
- Make sure `sdl serve` is running in another terminal
- Check that port 8080 is not used by another process
- Verify the server URL with `--server` flag if needed

**"No such file or directory"**
- Make sure you're in the SDL project directory
- Verify the SDL binary exists: `ls -la bin/sdl`
- Build if needed: `go build -o bin/sdl cmd/sdl/main.go`
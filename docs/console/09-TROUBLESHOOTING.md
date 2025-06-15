# SDL Console & Server Tutorial - Troubleshooting

This chapter provides comprehensive troubleshooting guidance for common SDL console and server issues.

## Prerequisites

Familiarity with previous tutorial chapters and basic system administration concepts.

## Common Connection Issues

### "Cannot connect to SDL server"

#### Symptoms
```
❌ Cannot connect to SDL server at http://localhost:8080

To use SDL console, first start the server:

🚀 Terminal 1: Start SDL server
   sdl serve

🔌 Terminal 2: Connect console client  
   sdl console
```

#### Causes and Solutions

**1. Server Not Running**
```bash
# Check if server process is running
ps aux | grep sdl
ps aux | grep "sdl serve"

# If not running, start server
./bin/sdl serve
```

**2. Wrong Port or Host**
```bash
# Check server configuration
netstat -an | grep 8080
ss -tlnp | grep 8080

# If server is on different port
sdl console --server http://localhost:9090
```

**3. Firewall Blocking Connection**
```bash
# Check firewall status
sudo ufw status
sudo iptables -L

# Allow port through firewall
sudo ufw allow 8080
```

**4. Server Crashed or Failed to Start**
```bash
# Check server logs for errors
./bin/sdl serve 2>&1 | tee server.log

# Common startup issues:
# - Port already in use
# - Insufficient permissions
# - Missing dependencies
```

### "Connection refused" or "Connection timeout"

#### For Local Connections
```bash
# Verify server is listening on correct interface
netstat -an | grep 8080
# Should show: tcp 0 0 127.0.0.1:8080 or 0.0.0.0:8080

# If only showing 127.0.0.1, restart server with:
./bin/sdl serve --host 0.0.0.0
```

#### For Remote Connections
```bash
# Test basic connectivity
ping remote-host
telnet remote-host 8080

# Test HTTP connectivity
curl http://remote-host:8080/api/health

# Check if server allows external connections
ssh remote-host
./bin/sdl serve --host 0.0.0.0 --port 8080
```

## Server Issues

### Server Won't Start

#### "Port already in use"
```
Error: listen tcp :8080: bind: address already in use
```

**Solutions:**
```bash
# Find process using port 8080
lsof -i :8080
netstat -tlnp | grep 8080

# Kill conflicting process
kill <process-id>

# Or use different port
./bin/sdl serve --port 8081
```

#### "Permission denied"
```
Error: listen tcp :80: bind: permission denied
```

**Solutions:**
```bash
# Use unprivileged port (>1024)
./bin/sdl serve --port 8080

# Or run with elevated privileges (not recommended)
sudo ./bin/sdl serve --port 80
```

#### Missing Dependencies
```
Error: ./bin/sdl: No such file or directory
```

**Solutions:**
```bash
# Build SDL binary
go build -o bin/sdl cmd/sdl/main.go

# Verify binary exists
ls -la bin/sdl

# Check binary permissions
chmod +x bin/sdl
```

### Server Performance Issues

#### High CPU Usage
```bash
# Monitor server resource usage
top -p $(pgrep sdl)
htop -p $(pgrep sdl)

# Check for excessive generator load
SDL> gen status
SDL> gen stop  # Stop all generators temporarily
```

#### High Memory Usage
```bash
# Monitor memory usage
ps aux | grep sdl
free -h

# Check DuckDB database size
du -h *.db
```

#### Slow Response Times
```bash
# Check server logs for slow operations
tail -f server.log | grep "slow"

# Monitor network latency
ping server-host

# Check disk I/O for database operations
iostat -x 1
```

## Console Issues

### Console Won't Start

#### "Command not found"
```
bash: sdl: command not found
```

**Solutions:**
```bash
# Use full path to binary
./bin/sdl console

# Or add to PATH
export PATH=$PATH:/path/to/sdl/bin
echo 'export PATH=$PATH:/path/to/sdl/bin' >> ~/.bashrc
```

#### Binary Permission Issues
```
bash: ./bin/sdl: Permission denied
```

**Solutions:**
```bash
# Fix binary permissions
chmod +x bin/sdl

# Verify permissions
ls -la bin/sdl
```

### Console Connection Problems

#### WebSocket Connection Failed
```
⚠️  WebSocket connection failed - falling back to HTTP polling
```

**Solutions:**
```bash
# Check server WebSocket endpoint
curl -i -N -H "Connection: Upgrade" -H "Upgrade: websocket" \
  http://localhost:8080/ws

# Verify no proxy interfering
unset http_proxy https_proxy

# Try explicit WebSocket URL
sdl console --server http://localhost:8080 --ws-url ws://localhost:8080/ws
```

#### Tab Completion Not Working
```bash
# Verify go-prompt library is working
SDL> <TAB>
# Should show available commands

# If not working, check terminal compatibility
echo $TERM
# Should be xterm, xterm-256color, or similar

# Try different terminal
export TERM=xterm
```

## Command Execution Issues

### "System not found" Errors

#### Loading SDL Files
```
SDL> load examples/contacts/contacts.sdl
❌ Error loading file: no such file or directory
```

**Solutions:**
```bash
# Verify file exists
ls -la examples/contacts/contacts.sdl

# Use absolute path
SDL> load /full/path/to/contacts.sdl

# Check current working directory
SDL> !pwd
```

#### Using Systems
```
SDL> use ContactsSystem
❌ Error: system 'ContactsSystem' not found
```

**Solutions:**
```bash
# Check what systems are available
SDL> info
Available systems: ContactsSystem, UserSystem

# Verify system name spelling
SDL> use <TAB>  # Use tab completion

# Reload file if needed
SDL> load examples/contacts/contacts.sdl
SDL> use ContactsSystem
```

### Generator Issues

#### Generators Not Starting
```
SDL> gen start load1
❌ Error starting generator 'load1': target method not found
```

**Solutions:**
```bash
# Verify target method exists
SDL> info
Available methods:
  - server.HandleLookup
  - server.HandleCreate

# Check generator configuration
SDL> gen list

# Recreate generator with correct target
SDL> gen remove load1
SDL> gen add load1 server.HandleLookup 10
SDL> gen start load1
```

#### Generators Producing Errors
```
✅ Generator load1: executed 10 calls to server.HandleLookup
❌ Generator load1 error: connection refused
```

**Solutions:**
```bash
# Check system configuration
SDL> get

# Verify system dependencies are running
SDL> !ps aux | grep database
SDL> !netstat -an | grep database_port

# Reduce generator rate
SDL> gen set load1 rate 5
```

### Measurement Issues

#### No Measurement Data
```
SDL> measure data lat1
📊 No data points available for measurement 'lat1'
```

**Solutions:**
```bash
# Verify measurement exists
SDL> measure list

# Check if generators are running
SDL> gen status

# Run manual execution to generate data
SDL> run test server.HandleLookup 10

# Verify measurement target matches execution target
SDL> measure remove lat1
SDL> measure add lat1 server.HandleLookup latency
```

#### DuckDB Connection Issues
```
❌ Error: database connection failed
```

**Solutions:**
```bash
# Check DuckDB file permissions
ls -la *.db

# Verify disk space
df -h .

# Restart server to reset database connection
# Ctrl+C to stop server
./bin/sdl serve
```

## Web Dashboard Issues

### Dashboard Not Loading

#### Browser Connection Failed
```
This site can't be reached
ERR_CONNECTION_REFUSED
```

**Solutions:**
```bash
# Verify server is running with web interface
curl http://localhost:8080

# Check server startup logs for web server errors
./bin/sdl serve 2>&1 | grep -i "web\|http\|dashboard"

# Try different browser
# Clear browser cache and cookies
```

#### Dashboard Shows "Disconnected"
```
🔴 Disconnected from server
```

**Solutions:**
```bash
# Check WebSocket connection
curl -i -N -H "Connection: Upgrade" -H "Upgrade: websocket" \
  http://localhost:8080/ws

# Refresh browser page
# Check browser console for JavaScript errors (F12)

# Verify server hasn't crashed
ps aux | grep sdl
```

### Dashboard Display Issues

#### Charts Not Updating
```bash
# Check measurement data exists
SDL> measure data lat1

# Verify generators are running
SDL> gen status

# Check browser console for errors (F12)
# Look for WebSocket connection errors

# Try refreshing page
# Clear browser cache
```

#### Missing Data in Charts
```bash
# Verify measurement time range
SDL> measure data lat1 --last 5m

# Check chart time window in dashboard
# Adjust time range controls in dashboard

# Verify system clock synchronization
date
# On server: date
```

## Performance Troubleshooting

### Slow Console Response

#### High Network Latency
```bash
# Test network latency to server
ping -c 10 server-host

# Use traceroute to identify bottlenecks
traceroute server-host

# For remote servers, consider SSH tunneling
ssh -L 8080:localhost:8080 user@remote-host
sdl console --server http://localhost:8080
```

#### Server Overload
```bash
# Check server resource usage
ssh server-host
top
htop
iostat -x 1

# Reduce load on server
SDL> gen stop  # Stop all generators
SDL> gen set load1 rate 5  # Reduce rates
```

### Memory Issues

#### High Memory Usage
```bash
# Monitor memory usage over time
while true; do
  ps aux | grep sdl | grep -v grep
  sleep 5
done

# Check for memory leaks
valgrind --leak-check=full ./bin/sdl serve

# Reduce measurement data retention
SDL> sql DELETE FROM traces WHERE timestamp < NOW() - INTERVAL 1 DAY
```

#### Out of Memory Errors
```
fatal error: runtime: out of memory
```

**Solutions:**
```bash
# Increase system memory limits
ulimit -v unlimited

# Use smaller batch sizes
SDL> set measurement.batch_size 100

# Clean up old measurement data
SDL> sql DELETE FROM traces WHERE timestamp < NOW() - INTERVAL 1 HOUR
```

## Data Issues

### DuckDB Problems

#### Database Locked
```
❌ Error: database is locked
```

**Solutions:**
```bash
# Stop all processes accessing database
pkill -f sdl

# Remove lock file if exists
rm -f *.db-wal *.db-shm

# Restart server
./bin/sdl serve
```

#### Corrupted Database
```
❌ Error: database disk image is malformed
```

**Solutions:**
```bash
# Backup existing database
cp measurements.db measurements.db.backup

# Try to repair database
sqlite3 measurements.db ".recover" | sqlite3 measurements_recovered.db

# If repair fails, start fresh
rm measurements.db
./bin/sdl serve
```

### File System Issues

#### Disk Space Full
```
❌ Error: no space left on device
```

**Solutions:**
```bash
# Check disk usage
df -h .

# Clean up old measurement data
SQL> sql DELETE FROM traces WHERE timestamp < NOW() - INTERVAL 1 DAY

# Rotate log files
rm server.log.old
mv server.log server.log.old
```

## Diagnostic Tools

### Enable Debug Logging
```bash
# Start server with debug logging
./bin/sdl serve --debug

# Start console with debug logging
sdl console --debug

# Set log level
export SDL_LOG_LEVEL=debug
```

### Network Diagnostics
```bash
# Test API endpoints
curl http://localhost:8080/api/health
curl http://localhost:8080/api/info

# Test WebSocket
wscat -c ws://localhost:8080/ws

# Monitor network traffic
tcpdump -i any port 8080
```

### System Diagnostics
```bash
# Check system resources
free -h
df -h
iostat -x 1

# Monitor processes
ps aux | grep sdl
lsof -p $(pgrep sdl)

# Check system limits
ulimit -a
```

## Getting Help

### Log Analysis
```bash
# Save server logs for analysis
./bin/sdl serve 2>&1 | tee sdl-server.log

# Save console session for debugging
script -a console-session.log
sdl console
# ... reproduce issue ...
exit
```

### System Information
```bash
# Gather system information
echo "SDL Version:" && ./bin/sdl version
echo "OS:" && uname -a
echo "Go Version:" && go version
echo "System Resources:" && free -h && df -h
```

### Error Reporting
When reporting issues, include:
1. **SDL version** - `./bin/sdl version`
2. **Operating system** - `uname -a`
3. **Error messages** - Exact text from console/server
4. **Steps to reproduce** - Minimal example that triggers issue
5. **Configuration** - Server flags, environment variables
6. **Logs** - Server and console output

## Prevention Best Practices

1. **Monitor Resource Usage** - Regular checks prevent issues
2. **Clean Up Data Regularly** - Prevent database bloat
3. **Use Reasonable Generator Rates** - Avoid overloading systems
4. **Test Network Connectivity** - Verify remote access works
5. **Backup Important Data** - Save measurement data and configurations
6. **Update Regularly** - Keep SDL version current
7. **Document Configurations** - Track server settings and changes
8. **Monitor Disk Space** - Ensure adequate storage for measurements
9. **Use SSH Tunnels** - Secure and reliable remote connections
10. **Implement Health Checks** - Automated monitoring for production use

This completes the SDL Console & Server Tutorial series. You now have comprehensive knowledge to effectively use SDL for system modeling, load testing, and performance analysis.
# SDL Console & Server Tutorial - Remote Access

This chapter covers connecting to SDL servers on remote machines, enabling distributed testing and collaborative development workflows.

## Prerequisites

Complete [Advanced Features](07-ADVANCED-FEATURES.md) and understand:
- Local SDL server and console setup
- Basic console operations and commands
- Network connectivity concepts

## Remote Access Overview

SDL's client-server architecture enables powerful remote access scenarios:

### Use Cases
- **Load Testing from Multiple Machines** - Distribute load generation
- **Collaborative Development** - Multiple developers sharing one server
- **Production Monitoring** - Monitor live systems from anywhere
- **Resource Scaling** - Run simulations on powerful remote servers
- **CI/CD Integration** - Automated testing from build systems

### Architecture Benefits
```
┌─────────────────┐    Network    ┌─────────────────┐
│  Local Console  │ ◄────────────► │  Remote Server  │
│  (Your Laptop)  │      HTTPS     │ (Powerful VM)   │
│                 │                │                 │
│ • Clean REPL    │                │ • Canvas Engine │
│ • Tab Complete  │                │ • Web Dashboard │
│ • History       │                │ • Measurements  │
└─────────────────┘                └─────────────────┘
```

## Basic Remote Connection

### Connect to Remote Server
```bash
# Connect to remote SDL server
sdl console --server http://remote-host:8080
```

Console output:
```
🔌 SDL Console Client v1.0
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
🎯 Server:       http://remote-host:8080
📊 Dashboard:    http://remote-host:8080 (open in browser)
💬 Type 'help' for available commands, Ctrl+D to quit
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

✅ Connected to SDL server at remote-host:8080

SDL> 
```

### Test Remote Connection
```
SDL> info
Connected to: remote-host:8080
Server Version: SDL v1.0
Uptime: 2h 34m 15s
Active Systems: 0
Active Generators: 0
Active Measurements: 0
```

### Access Remote Dashboard
Open your browser to:
```
http://remote-host:8080
```
The dashboard works identically to local access.

## Setting Up Remote Servers

### Basic Remote Server Setup

#### 1. Start Server on Remote Machine
```bash
# On the remote machine
ssh user@remote-host
cd /path/to/sdl
./bin/sdl serve --host 0.0.0.0 --port 8080
```

Server output:
```
🚀 SDL Server v1.0
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
🎯 Canvas Engine: Ready
📊 Web Dashboard: http://0.0.0.0:8080 (accessible from any network interface)
🔌 REST API:      http://0.0.0.0:8080/api
📈 WebSocket:     ws://0.0.0.0:8080/ws
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

✅ SDL server started successfully
💡 Server accessible from: http://your-ip:8080
🌐 Network interfaces: all (0.0.0.0)
```

#### 2. Connect from Local Machine
```bash
# On your local machine
sdl console --server http://remote-host:8080
```

### Advanced Server Configuration

#### Custom Port Configuration
```bash
# Start server on custom port
./bin/sdl serve --host 0.0.0.0 --port 9090

# Connect to custom port
sdl console --server http://remote-host:9090
```

#### Multiple Server Management
```bash
# Development server
./bin/sdl serve --host 0.0.0.0 --port 8080

# Testing server  
./bin/sdl serve --host 0.0.0.0 --port 8081

# Production server
./bin/sdl serve --host 0.0.0.0 --port 8082
```

## Security Considerations

### Network Security

#### Firewall Configuration
```bash
# Allow SDL server port through firewall
sudo ufw allow 8080
sudo ufw allow from 192.168.1.0/24 to any port 8080  # Restrict to local network
```

#### SSH Tunneling (Recommended)
For secure connections, use SSH tunneling:

```bash
# Create SSH tunnel
ssh -L 8080:localhost:8080 user@remote-host

# In another terminal, connect to tunneled port
sdl console --server http://localhost:8080
```

This approach:
- Encrypts all traffic between console and server
- Doesn't require opening firewall ports
- Uses existing SSH authentication

### HTTPS Configuration

#### SSL Certificate Setup
```bash
# Generate self-signed certificate (development only)
openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -days 365 -nodes

# Start server with HTTPS
./bin/sdl serve --host 0.0.0.0 --port 8443 --tls --cert cert.pem --key key.pem
```

#### Connect to HTTPS Server
```bash
# Connect to HTTPS server
sdl console --server https://remote-host:8443
```

### Access Control

#### IP-based Restrictions
```bash
# Start server with IP restrictions
./bin/sdl serve --host 0.0.0.0 --port 8080 --allow-ips "192.168.1.0/24,10.0.0.0/8"
```

#### Authentication Integration
```bash
# Start server with API key authentication
./bin/sdl serve --host 0.0.0.0 --port 8080 --api-key "your-secret-key"

# Connect with API key
sdl console --server http://remote-host:8080 --api-key "your-secret-key"
```

## Multi-Client Scenarios

### Collaborative Development

#### Scenario: Multiple Developers
```
Developer 1 (Lead):
  - Sets up remote server with shared SDL files
  - Configures baseline measurements and generators
  
Developer 2 (Client):  
  - Connects to shared server
  - Runs complementary tests
  - Views real-time results in dashboard
  
Developer 3 (Monitor):
  - Connects read-only to observe testing
  - Monitors dashboard for performance trends
```

#### Implementation:
```bash
# Developer 1: Set up shared server
ssh dev-server
cd /shared/sdl-project
./bin/sdl serve --host 0.0.0.0 --port 8080

# Developer 2: Connect and contribute
sdl console --server http://dev-server:8080
SDL> load shared_system.sdl
SDL> use SharedSystem
SDL> gen add dev2_load server.HandleWrites 15

# Developer 3: Monitor activity
sdl console --server http://dev-server:8080
SDL> gen status
SDL> measure list
# Also monitor dashboard at http://dev-server:8080
```

### Load Distribution

#### Scenario: Multi-Machine Load Testing
Generate load from multiple client machines:

```bash
# Machine 1: High read load
sdl console --server http://test-server:8080
SDL> gen add reads_machine1 server.HandleLookup 100

# Machine 2: Write load
sdl console --server http://test-server:8080  
SDL> gen add writes_machine2 server.HandleCreate 25

# Machine 3: Mixed load
sdl console --server http://test-server:8080
SDL> gen add mixed_machine3 server.HandleLookup 50
SDL> gen add updates_machine3 server.HandleUpdate 10

# Coordinated start from all machines
SDL> gen start
```

The server aggregates all load and shows combined metrics.

## Cloud Deployment

### Docker Container Deployment

#### Dockerfile for SDL Server
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o bin/sdl cmd/sdl/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/bin/sdl .
COPY --from=builder /app/examples ./examples
EXPOSE 8080
CMD ["./sdl", "serve", "--host", "0.0.0.0", "--port", "8080"]
```

#### Deploy and Connect
```bash
# Build and run container
docker build -t sdl-server .
docker run -p 8080:8080 -d sdl-server

# Connect from local machine
sdl console --server http://docker-host:8080
```

### Kubernetes Deployment

#### SDL Server Deployment
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: sdl-server
spec:
  replicas: 1
  selector:
    matchLabels:
      app: sdl-server
  template:
    metadata:
      labels:
        app: sdl-server
    spec:
      containers:
      - name: sdl-server
        image: sdl-server:latest
        ports:
        - containerPort: 8080
        args: ["./sdl", "serve", "--host", "0.0.0.0", "--port", "8080"]
---
apiVersion: v1
kind: Service
metadata:
  name: sdl-service
spec:
  selector:
    app: sdl-server
  ports:
  - protocol: TCP
    port: 8080
    targetPort: 8080
  type: LoadBalancer
```

#### Connect to Kubernetes Service
```bash
# Get service external IP
kubectl get service sdl-service

# Connect to Kubernetes-hosted server
sdl console --server http://external-ip:8080
```

### Cloud Provider Examples

#### AWS EC2 Deployment
```bash
# Launch EC2 instance with SDL server
aws ec2 run-instances \
  --image-id ami-12345678 \
  --instance-type c5.large \
  --security-group-ids sg-12345678 \
  --user-data file://sdl-setup.sh

# Connect from anywhere
sdl console --server http://ec2-instance-ip:8080
```

#### Google Cloud Deployment  
```bash
# Create GCE instance with SDL server
gcloud compute instances create sdl-server \
  --machine-type=n1-standard-2 \
  --metadata-from-file startup-script=sdl-setup.sh

# Connect from local machine
sdl console --server http://gce-instance-ip:8080
```

## Remote Monitoring and Automation

### Remote Recipe Execution

#### Automated Testing Pipeline
```bash
# Upload test recipes to remote server
scp test_suite.recipe user@remote-host:/shared/recipes/

# Connect and execute remotely
sdl console --server http://remote-host:8080
SDL> execute /shared/recipes/test_suite.recipe
```

#### CI/CD Integration
```bash
#!/bin/bash
# ci-test.sh - Automated testing script

# Connect to remote test environment
sdl console --server http://test-env:8080 << EOF
load production_system.sdl
use ProductionSystem
execute automated_regression_test.recipe
measure export results csv test_results.csv
EOF

# Process results
if grep -q "FAILED" test_results.csv; then
  echo "Tests failed - blocking deployment"
  exit 1
else
  echo "Tests passed - proceeding with deployment"
fi
```

### Remote Monitoring Scripts

#### Health Check Script
```bash
#!/bin/bash
# health-check.sh

SERVER="http://production-server:8080"

# Check server connectivity
if sdl console --server $SERVER --command "info" > /dev/null 2>&1; then
  echo "✅ SDL server healthy"
else
  echo "❌ SDL server unavailable"
  exit 1
fi

# Check measurement data freshness
LAST_DATA=$(sdl console --server $SERVER --command "measure data main_latency --last 1m" | wc -l)
if [ $LAST_DATA -gt 0 ]; then
  echo "✅ Recent measurement data available"
else
  echo "⚠️ No recent measurement data"
fi
```

## Connection Management

### Server Discovery
```bash
# Scan network for SDL servers
nmap -p 8080 192.168.1.0/24

# Test specific servers
for host in server1 server2 server3; do
  echo "Testing $host..."
  sdl console --server http://$host:8080 --command "info" || echo "Failed"
done
```

### Connection Troubleshooting

#### Network Connectivity
```bash
# Test basic connectivity
ping remote-host
telnet remote-host 8080
curl http://remote-host:8080/api/health

# Test from SDL console
sdl console --server http://remote-host:8080 --debug
```

#### Firewall Issues
```bash
# Check if port is open
nmap -p 8080 remote-host

# Check server-side listening
ssh remote-host
netstat -an | grep 8080
ss -tlnp | grep 8080
```

## What's Next?

Continue to **[Troubleshooting](09-TROUBLESHOOTING.md)** to learn how to diagnose and resolve common issues with SDL console and server operations.

## Remote Access Best Practices

1. **Use SSH Tunneling** - Secure connections without exposing ports
2. **Monitor Network Latency** - High latency affects console responsiveness  
3. **Implement Access Control** - Restrict server access to authorized users
4. **Use Meaningful Server Names** - Easy identification in multi-server setups
5. **Document Server Configurations** - Track which servers run which tests
6. **Automate Server Setup** - Use scripts for consistent deployments
7. **Monitor Server Resources** - Ensure adequate CPU/memory for remote load
8. **Plan for Network Failures** - Design resilient testing workflows
9. **Use Container Orchestration** - Kubernetes/Docker for scalable deployments
10. **Implement Health Checks** - Automated monitoring of remote servers
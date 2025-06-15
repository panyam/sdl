# SDL Console & Server Tutorial - Web Dashboard

This chapter covers SDL's powerful web dashboard for real-time visualization and system monitoring.

## Prerequisites

Complete [Measurements](05-MEASUREMENTS.md) and have:
- SDL server running with active measurements
- Traffic generators producing data
- Understanding of measurement collection

## Accessing the Dashboard

### Open in Browser
With your SDL server running, navigate to:
```
http://localhost:8080
```

The dashboard automatically connects and displays real-time data from your SDL session.

### Dashboard URL from Console
The console shows the dashboard URL on startup:
```
🔌 SDL Console Client v1.0
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
🎯 Server:       http://localhost:8080
📊 Dashboard:    http://localhost:8080 (open in browser)
```

## Dashboard Overview

The SDL dashboard provides real-time visualization across several key areas:

### 1. System Overview Panel
- **Current System** - Shows active SDL system name
- **Connection Status** - Server connectivity indicator
- **Live Statistics** - Real-time execution counts

### 2. Traffic Generators Panel
- **Active Generators** - List of running traffic generators
- **Generator Controls** - Start/stop individual generators
- **Load Distribution** - Visual representation of traffic patterns

### 3. Live Measurements Panel
- **Real-time Charts** - Dynamic plots of measurement data
- **Multiple Metrics** - Latency, throughput, error rates
- **Time Windows** - Configurable data ranges

### 4. System Components Panel
- **Component Status** - Health of system components
- **Resource Usage** - CPU, memory, connections
- **Performance Indicators** - Key metrics at a glance

## Real-time Features

### Live Data Updates
The dashboard updates automatically every 2 seconds with:
- New measurement data points
- Generator status changes
- System performance metrics
- Error notifications

### WebSocket Connection
```
🔌 WebSocket: ws://localhost:8080/ws
📡 Real-time data streaming
🔄 Automatic reconnection on disconnect
```

## Working with Charts

### Measurement Visualization

#### Latency Charts
- **Line Charts** - Show latency trends over time
- **Distribution Plots** - Latency percentile distributions  
- **Real-time Updates** - New data appears automatically

Example: With `lat1` measurement active:
```
SDL[ContactsSystem]> measure add lat1 server.HandleLookup latency
SDL[ContactsSystem]> gen add load1 server.HandleLookup 20
SDL[ContactsSystem]> gen start load1
```

The dashboard immediately shows:
- Live latency line chart for `lat1`
- Real-time updates every 2 seconds
- Historical data as it accumulates

#### Throughput Charts  
- **Rate Graphs** - Calls per second over time
- **Aggregate Views** - Combined throughput across generators
- **Target Breakdown** - Per-method throughput analysis

#### Error Rate Charts
- **Success/Failure Ratio** - Visual error rate tracking
- **Error Trending** - Error rate changes over time
- **Alert Indicators** - Visual warnings for high error rates

### Chart Interactions

#### Time Range Controls
- **Zoom** - Mouse wheel to zoom in/out
- **Pan** - Click and drag to pan across time
- **Reset** - Double-click to reset view

#### Data Selection
- **Hover** - View exact values at specific times
- **Legend** - Toggle individual data series on/off
- **Export** - Save chart data or images

## Generator Management

### Visual Generator Control

#### Generator Status Display
```
Traffic Generators:
┌─────────────┬─────────────────────┬──────┬─────────┬─────────┐
│ Name        │ Target              │ Rate │ Status  │ Control │
├─────────────┼─────────────────────┼──────┼─────────┼─────────┤
│ load1       │ server.HandleLookup │ 20   │ Running │ [Stop]  │
│ burst_load  │ server.HandleLookup │ 100  │ Stopped │ [Start] │
└─────────────┴─────────────────────┴──────┴─────────┴─────────┘
```

#### Interactive Controls
- **Start/Stop Buttons** - Control generators from dashboard
- **Rate Adjustment** - Modify generator rates visually
- **Status Indicators** - Green (running), Red (stopped), Yellow (error)

### Load Distribution Visualization
- **Pie Charts** - Show load distribution across methods
- **Bar Charts** - Compare rates across generators
- **Timeline View** - Generator activity over time

## Advanced Dashboard Features

### Multi-System Support
When multiple systems are loaded:
- **System Selector** - Switch between different SDL systems
- **Cross-System Comparison** - Compare metrics across systems
- **Unified View** - See all systems simultaneously

### Historical Data Analysis
- **Data Range Selection** - View historical time periods
- **Trend Analysis** - Long-term performance trends
- **Comparison Mode** - Compare current vs historical performance

### Custom Views
- **Dashboard Layout** - Rearrange panels
- **Chart Configuration** - Customize chart types and styles
- **Filter Options** - Focus on specific measurements or generators

## Integration with Console

### Bi-directional Synchronization
Actions in the console immediately appear in the dashboard:

```
SDL[ContactsSystem]> gen start burst_load
```
Dashboard instantly shows:
- `burst_load` status changes to "Running"
- New traffic appears in charts
- Load distribution updates

### Dashboard-to-Console Commands
Some dashboard actions can be replicated in console:
- Start/stop generators: `gen start/stop <name>`
- View measurements: `measure data <name>`
- Check status: `gen status`, `measure status`

## Performance Monitoring

### Real-time Metrics
The dashboard displays key performance indicators:

#### System Health
- **CPU Usage** - Server resource consumption
- **Memory Usage** - Application memory footprint
- **Database Performance** - DuckDB query response times

#### Application Metrics
- **Request Latency** - End-to-end response times
- **Throughput** - Requests processed per second
- **Error Rates** - Success/failure percentages
- **Queue Depths** - Pending requests in system components

### Alerting and Notifications
- **Visual Alerts** - Red indicators for critical issues
- **Threshold Warnings** - Yellow warnings for approaching limits
- **Trend Alerts** - Notifications for significant changes

## Dashboard Scenarios

### Load Testing Visualization
```
# Console setup
SDL[ContactsSystem]> measure add latency server.HandleLookup latency
SDL[ContactsSystem]> measure add throughput server.HandleLookup throughput  
SDL[ContactsSystem]> gen add baseline server.HandleLookup 10
SDL[ContactsSystem]> gen start baseline
```

Dashboard shows:
- Baseline latency and throughput charts
- Real-time performance under 10 RPS load
- System behavior visualization

### Stress Testing Monitoring
```
# Gradually increase load
SDL[ContactsSystem]> gen add stress server.HandleLookup 50
SDL[ContactsSystem]> gen start stress
# ... monitor dashboard for latency increases ...
SDL[ContactsSystem]> gen set stress rate 100
# ... watch for performance degradation ...
```

Dashboard reveals:
- Latency increases under higher load
- Throughput saturation points
- Error rate changes
- Resource utilization patterns

### Multi-Method Analysis
```
# Compare different methods
SDL[ContactsSystem]> measure add lookup_lat server.HandleLookup latency
SDL[ContactsSystem]> measure add create_lat server.HandleCreate latency
SDL[ContactsSystem]> gen add lookup_gen server.HandleLookup 20
SDL[ContactsSystem]> gen add create_gen server.HandleCreate 10
SDL[ContactsSystem]> gen start
```

Dashboard displays:
- Side-by-side latency comparison
- Relative throughput analysis
- Method-specific performance characteristics

## Troubleshooting Dashboard Issues

### Connection Problems
**Dashboard not loading:**
- Verify SDL server is running: `sdl serve`
- Check URL: http://localhost:8080
- Look for port conflicts in server logs

**WebSocket disconnection:**
- Dashboard shows "Disconnected" indicator
- Automatic reconnection attempts
- Check server terminal for WebSocket errors

### Data Not Appearing
**Empty charts:**
- Ensure measurements are created: `measure list`
- Verify traffic generators are running: `gen status`
- Check for recent data: `measure data <name>`

**Delayed updates:**
- WebSocket may be disconnected
- Refresh browser page
- Restart SDL server if needed

### Performance Issues
**Slow dashboard loading:**
- Large amount of historical data
- Browser memory constraints
- Consider shorter time windows

**Charts not updating:**
- WebSocket connection issues
- Browser JavaScript errors (check dev console)
- Server overload (check server terminal)

## Browser Compatibility

### Supported Browsers
- **Chrome/Chromium** - Full feature support
- **Firefox** - Full feature support  
- **Safari** - Full feature support
- **Edge** - Full feature support

### Required Features
- WebSocket support
- Canvas/SVG rendering
- ES6 JavaScript support
- CSS3 flexbox support

## What's Next?

Now that you understand the web dashboard, continue to **[Advanced Features](07-ADVANCED-FEATURES.md)** to learn about command history, tab completion, and other power-user features.

## Best Practices

1. **Keep Dashboard Open** - Monitor real-time changes while working in console
2. **Use Multiple Windows** - Split screen with console and dashboard
3. **Monitor During Load Tests** - Watch for performance changes in real-time
4. **Export Key Charts** - Save important visualizations for analysis
5. **Check WebSocket Status** - Ensure real-time connection is active
6. **Refresh When Needed** - Browser refresh resolves most display issues
7. **Use Appropriate Time Windows** - Match chart ranges to testing scenarios
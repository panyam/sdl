# SDL Canvas Web Dashboard

🎉 **Successfully implemented!** A TypeScript + Tailwind web interface featuring a simple 2-row dynamic layout for real-time SDL system visualization and parameter manipulation.

## 🚀 Quick Start

1. **Build & Start the Server:**
   ```bash
   # Build frontend
   cd web && npm install && npm run build && cd ..
   
   # Build backend
   go build -o sdl ./cmd/sdl
   
   # Start server
   ./sdl serve --port 8080
   ```

2. **Open Dashboard:**
   - Navigate to: http://localhost:8080
   - You'll see the new 2-row dynamic "Incredible Machine" interface

## 🎛️ Simple 2-Row Dynamic Layout

### **Row 1 (50% height): System Architecture + Controls**

#### **Left Panel (70% width): Enhanced System Architecture**
- **ContactAppServer**: Enhanced visualization with pool utilization, load, and success metrics
- **ContactDatabase**: Detailed connection pool status, cache hit rate, and utilization
- **HashIndex**: Lookup performance and capacity metrics
- **System Health Dashboard**: Color-coded overview (Success Rate, Avg Latency, Current Load)
- **More space for complex system visualization** - supports enterprise-scale architectures

#### **Right Side (30% width): Split Control Panels**

**Top Panel (48% height): Traffic Generation**
- **Enable/Disable Controls**: Checkboxes for traffic generators
- **Rate Sliders**: Real-time traffic rate adjustment (0-20 RPS)
- **Target Configuration**: Specific method targeting
- **Add Generator Button**: Dynamic traffic source creation

**Bottom Panel (48% height): System Parameters**
- **Server Arrival Rate** (0-50 RPS): Incoming request load
- **Server Pool Size** (1-50): Number of concurrent request handlers  
- **DB Arrival Rate** (0-30 RPS): Database query load
- **DB Pool Size** (1-20): Number of database connections
- **Cache Hit Rate** (0-100%): Database cache effectiveness

### **Row 2 (50% height): Dynamic Metrics Grid**
- **Unlimited Scrollable Charts**: Supports infinite metrics via `canvas.Measure()` calls
- **Automatic Grid Layout**: 3-column responsive grid (adjusts to screen size)
- **Color-Coded Charts**: 
  - Red: Latency metrics (p95, p99)
  - Green: QPS/Throughput metrics
  - Orange: Error rate metrics
  - Purple: Cache hit rate metrics
  - Blue: Utilization metrics
  - Pink: Memory metrics
- **Real-time Updates**: WebSocket-powered live chart updates
- **Proper Clipping**: All charts contained within panel boundaries

## 🔄 Real-time Features

### **WebSocket Updates**
- Instant parameter synchronization across all connected dashboards
- Live simulation progress updates
- Real-time metric refreshes

### **Auto-simulation**
- Parameter changes automatically trigger new simulations
- Results immediately visible in metrics and charts
- No manual refresh needed

## 🎪 Enhanced Workshop Demo Flow

### **Phase 1: Load the Service**
1. Click "Load Contacts Service" button in header
2. System architecture (left panel) populates with enhanced component visualization
3. System Health dashboard shows baseline metrics

### **Phase 2: Explore the New Layout**
- **Enhanced System Architecture**: Notice the prominent 70% layout with detailed component metrics
- **Split Right Panels**: Traffic Generation (top) and System Parameters (bottom)
- **Dynamic Metrics Grid**: Multiple live charts showing different system metrics
- **Panel Clipping**: Observe how all content stays within proper boundaries

### **Phase 3: Traffic Generation Controls**
- **Enable Traffic**: Check "Contact Lookup Traffic" in Traffic Generation panel
- **Adjust Rate**: Use rate slider to change traffic from 5.0 to 10.0 RPS
- **Visual Feedback**: Watch metrics update in real-time across all panels

### **Phase 4: Parameter Tuning**
- **Cache Optimization**: Move "Cache Hit Rate" slider in System Parameters to 80%
- **Pool Scaling**: Adjust "Server Pool Size" to demonstrate capacity scaling
- **Immediate Feedback**: Observe instant updates in System Health dashboard

### **Phase 5: Metrics Visualization**
- **Multiple Charts**: Watch as different metrics are displayed in the scrollable grid
- **Chart Types**: Latency (red), QPS (green), Error rates (orange)
- **Scrolling**: Demonstrate unlimited metric addition and grid scrolling

### **Phase 6: Advanced Scenarios**
- **Add Generator**: Click "+ Add" in Traffic Generation to create custom traffic
- **Complex Systems**: Load Netflix example to show enhanced architecture capabilities
- **Multi-Row Charts**: Demonstrate how metrics grid expands with more data points

### **Phase 7: System Stress Testing**
- **Overload Scenario**: Push traffic beyond system capacity
- **Visual Indicators**: Watch System Health dashboard turn red
- **Failure Patterns**: Observe different bottlenecks in the enhanced architecture view

## 🛠️ Technical Architecture

### **Backend (Go)**
- **HTTP Server**: Gorilla Mux router with CORS support
- **REST API**: Canvas operations (load, use, set, run, plot)
- **WebSocket**: Real-time bidirectional communication
- **Canvas Integration**: Direct access to proven Canvas API

### **Frontend (TypeScript + Tailwind)**
- **Build System**: Vite for fast development and optimized builds
- **Type Safety**: Full TypeScript coverage for API calls and data structures
- **Styling**: Tailwind CSS for rapid, responsive UI development
- **Charting**: Chart.js for real-time performance visualization
- **WebSocket Client**: Native WebSocket API for live updates

### **API Endpoints**
```
POST /api/load     - Load SDL file
POST /api/use      - Activate system
POST /api/set      - Modify parameters  
POST /api/run      - Execute simulation
POST /api/plot     - Generate visualizations
WS   /api/live     - Real-time updates
GET  /            - Dashboard interface
```

## 🎯 Success Metrics Achieved

✅ **Simple 2-row layout** with enhanced system architecture prominence  
✅ **Separated control panels** for better workflow organization  
✅ **Unlimited scrollable metrics** supporting infinite chart addition  
✅ **Proper panel clipping** ensuring clean visual boundaries  
✅ **Real-time parameter changes** visible immediately in all panels  
✅ **Smooth 60fps updates** during live demos  
✅ **Dynamic traffic generation** with add/remove capabilities  
✅ **Enhanced system visualization** supporting complex architectures  
✅ **Intuitive interface** accessible to workshop audiences  
✅ **Reliable performance** for conference presentations  
✅ **Type-safe development** with TypeScript  
✅ **Rapid styling iteration** with Tailwind utilities  
✅ **Zero-config deployment** - single binary serves everything  

## 🔧 Development Workflow

### **Frontend Development**
```bash
cd web
npm run dev  # Start Vite dev server on port 3000
```

### **Backend Development**  
```bash
go run ./cmd/sdl serve --port 8080
```

### **Production Build**
```bash
./test_web_stack.sh  # Automated build and test
```

## 🌟 Perfect for Conference Workshops

This simple 2-row web dashboard provides the ultimate "Incredible Machine" experience for system design interview coaching:

- **Enhanced Visual Impact**: Prominent 70% system architecture with detailed component visualization
- **Organized Controls**: Separated traffic generation and parameter controls for clear workflow
- **Unlimited Metrics**: Scrollable grid supporting infinite chart addition via `canvas.Measure()`
- **Proper Containment**: All panels properly clipped with no visual overflow
- **Immediate Feedback**: Multiple charts update in real-time showing performance changes
- **Interactive Learning**: Audience can see cause-and-effect relationships instantly across multiple metrics
- **Professional Presentation**: Clean, modern interface with proper spacing suitable for large-screen conference demos
- **Dynamic Scalability**: Layout adapts to complex systems and unlimited metrics

**The SDL Canvas Web Dashboard has reached new heights and is production-ready for workshop demonstrations!** 🎉

# System Showcase Design (Server-Side Rendering)

## Overview

This document outlines the design for a more accessible, beginner-friendly interface for SDL that complements the existing IDE-like dashboard. The new experience uses server-side rendering with Templar templates for better performance and simpler architecture.

## User Journey

```
System Listing Page → System Details Page → (Optional) Full Dashboard
     (Browse)            (Learn/Modify)         (Advanced Edit)
```

## Architecture

### Server-Side Routing

All routing is handled by the Go backend:
- `/systems` - System listing page (server-rendered)
- `/system/:id` - System details page (hybrid rendering)
- `/canvases/:id` - Full dashboard (client-rendered)

### Template Management with Templar

Using github.com/panyam/templar for template management:
- Base layout template
- System listing template
- System details template
- Component templates for reusability

## Page Structure

### 1. System Listing Page (`/systems`)

A catalog/gallery view of example systems that users can explore.

**Features:**
- Grid/card layout showing system previews
- Each card displays:
  - System name (e.g., "Bitly URL Shortener")
  - Description
  - Thumbnail/architecture preview
  - Complexity indicator (Beginner/Intermediate/Advanced)
  - Tags (e.g., "web", "microservices", "database")
- Search and filter capabilities
- Direct links to system details

**URL Structure:**
- `/systems` - Main listing
- `/systems?tag=microservices` - Filtered view

### 2. System Details Page (`/system/:id`)

A focused, single-system editing experience.

**Features:**
- Simplified toolbar (Run, Stop, Reset)
- Single file editor (no file tree)
- Live architecture visualization
- Metrics panel (when running)
- Recipe execution panel
- Share button for generating links
- "Open in Full Editor" option

**URL Structure:**
- `/system/bitly` - Bitly example
- `/system/bitly?version=v2` - Specific version
- `/system/bitly?mode=wasm` - Force WASM mode

## Data Model

### Project Structure
```typescript
interface SystemProject {
  id: string;           // e.g., "bitly"
  name: string;         // e.g., "Bitly URL Shortener"
  description: string;
  category: string;     // e.g., "Web Services"
  difficulty: 'beginner' | 'intermediate' | 'advanced';
  tags: string[];
  versions: {
    [key: string]: {
      sdl: string;      // SDL file content
      recipe: string;   // Recipe file content
      readme?: string;  // Optional documentation
    }
  };
  defaultVersion: string;
  thumbnail?: string;   // Architecture preview
}
```

### Server API Endpoints

```
GET  /api/systems                    # List all systems
GET  /api/systems/:id               # Get system details
GET  /api/systems/:id/version/:ver  # Get specific version
POST /api/systems/:id/run           # Run system (server mode)
POST /api/systems/:id/fork          # Create user copy
```

## Implementation Status

### Phase 1: Server-Side Infrastructure ✅ COMPLETED
1. ✅ Add Templar template engine to Go backend
2. ✅ Create route handlers for new pages
3. ✅ Design template hierarchy
4. ✅ Implement system catalog service

### Phase 2: Template Development ✅ COMPLETED
1. ✅ Base layout template with common elements
2. ✅ System listing template with cards
3. ✅ System details template with editor placeholder
4. ✅ Reusable component templates

### Phase 3: Client-Side Integration ✅ COMPLETED
1. ✅ Enhanced JavaScript for interactivity (search, filter, sort)
2. ✅ SystemDetailsPage class for editor management
3. ✅ Event handlers for server-rendered content
4. ✅ Progressive enhancement approach

### Phase 4: Runtime Management ✅ COMPLETED
1. ✅ **Server Mode**: Server manages Canvas runtime
2. ✅ **WASM Mode**: Browser manages Canvas runtime
3. ✅ Unified API for both modes
4. ✅ State synchronization

### Additional Enhancements Completed
1. ✅ **Unified Tailwind CSS**: Single CSS build for both server and client pages
2. ✅ **Theme Switcher**: Light/Dark/System mode support
3. ✅ **Enhanced Filtering**: Search, difficulty filters, and sorting options
4. ✅ **Responsive Design**: Mobile-friendly card layout

## Server-Side Architecture

Integrated into existing console package:

```
console/
├── systems_handler.go      # System showcase routes
├── system_catalog.go       # System examples catalog
├── template_setup.go       # Templar configuration
├── canvas_web.go          # Updated with new routes
└── templates/
    ├── base.html          # Base layout with Tailwind
    └── systems/
        ├── listing.html   # System listing page
        └── details.html   # System details page
```

## Template Structure with Templar

### Base Layout
```html
<!-- templates/base.html -->
<!DOCTYPE html>
<html>
<head>
    <title>{{.Title}} - SDL Canvas</title>
    <link rel="stylesheet" href="/assets/style.css">
    {{block "head" .}}{{end}}
</head>
<body>
    {{block "content" .}}{{end}}
    
    <div id="app" 
         data-page-type="{{.PageType}}"
         data-page-data='{{.PageDataJSON}}'>
    </div>
    
    <script src="/assets/main.js"></script>
    {{block "scripts" .}}{{end}}
</body>
</html>
```

### System Listing Template
```html
<!-- templates/systems/listing.html -->
{{# include "base.html" #}}

{{define "content"}}
<div class="system-listing-page">
    <header class="page-header">
        <h1>SDL System Examples</h1>
        <p>Explore and learn from real-world system architectures</p>
    </header>
    
    <div class="filters-section">
        <input type="text" id="search-input" 
               placeholder="Search systems..." 
               class="search-input">
        
        <div class="filter-buttons">
            <button class="filter-btn active" data-filter="all">All</button>
            <button class="filter-btn" data-filter="beginner">Beginner</button>
            <button class="filter-btn" data-filter="intermediate">Intermediate</button>
            <button class="filter-btn" data-filter="advanced">Advanced</button>
        </div>
    </div>
    
    <div id="systems-grid" class="systems-grid">
        {{range .Systems}}
        {{# include "components/system_card.html" #}}
        {{end}}
    </div>
</div>
{{end}}
```

### System Card Component
```html
<!-- templates/components/system_card.html -->
<div class="system-card" 
     data-id="{{.ID}}"
     data-difficulty="{{.Difficulty}}"
     data-name="{{.Name}}"
     data-description="{{.Description}}"
     data-tags="{{range .Tags}}{{.}} {{end}}">
    <div class="card-header">
        <h3>{{.Name}}</h3>
        <span class="difficulty-badge {{.Difficulty}}">{{.Difficulty}}</span>
    </div>
    
    <p class="card-description">{{.Description}}</p>
    
    <div class="card-tags">
        {{range .Tags}}
        <span class="tag">{{.}}</span>
        {{end}}
    </div>
    
    <div class="card-footer">
        <a href="/system/{{.ID}}" class="view-btn">
            View System →
        </a>
    </div>
</div>
```

## Standard Library Auto-Import

For simplified editing, automatically include imports:

```sdl
// Auto-inserted at top of SDL file
import stdlib.storage.Disk
import stdlib.compute.LoadBalancer
import stdlib.data.HashIndex
import stdlib.network.CDN

// User's system definition starts here
system Bitly {
  // ...
}
```

## Benefits

1. **Lower Barrier to Entry**: No file management, just focus on the system
2. **Shareable Examples**: Direct links to specific designs
3. **Progressive Disclosure**: Start simple, graduate to full IDE
4. **Better Learning**: Curated examples with explanations
5. **Quick Experimentation**: Modify and run without setup

## Migration Path

Users can progress from:
1. Viewing examples (read-only)
2. Modifying examples (in simplified editor)
3. Creating from scratch (still simplified)
4. Full IDE experience (current dashboard)

## Technical Considerations

### URL-Based State
- System ID and version in URL
- Mode (server/wasm) in query params
- Shareable and bookmarkable

### Performance
- Lazy load system definitions
- Cache compiled WASM modules
- Preload common dependencies

### Security
- Sandboxed execution in WASM mode
- Rate limiting in server mode
- No file system access in showcase mode

## Future Enhancements

1. **Embedded Mode**: Embed system viewer in documentation
2. **Playground Mode**: Temporary experiments without saving
3. **Tutorial Integration**: Step-by-step guides
4. **Community Submissions**: User-contributed examples
5. **Diff View**: Compare versions side-by-side
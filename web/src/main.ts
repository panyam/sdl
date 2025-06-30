import './style.css';
import { Dashboard } from './dashboard.js';
import { WASMDashboard } from './wasm-dashboard.js';

// Extract canvas ID from URL path
function getCanvasIdFromUrl(): string {
  const path = window.location.pathname;
  const match = path.match(/^\/canvases\/([^\/]+)/);
  
  if (match && match[1]) {
    return match[1];
  }
  
  // Default to 'default' if no canvas ID found
  return 'default';
}

// Check if Server mode is requested
function isServerMode(): boolean {
  const urlParams = new URLSearchParams(window.location.search);
  return urlParams.get('server') === 'true';
}

// Initialize the dashboard when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
  setTimeout(() => {
    const canvasId = getCanvasIdFromUrl();
    const useServerCanvas = isServerMode();
    
    console.log(`🚀 SDL Canvas Dashboard starting for canvas: ${canvasId} (Mode: ${useServerCanvas ? "ServerRuntime" : "WASMRuntime"})`);
    
    if (useServerCanvas ) {
      const dashboard = new Dashboard(canvasId);
      dashboard.initialize();
    } else {
      const dashboard = new WASMDashboard(canvasId);
      dashboard.initialize();
    }
  }, 100);
});

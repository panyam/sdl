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

// Check if WASM mode is requested
function isWASMMode(): boolean {
  return true;
  /*
  const urlParams = new URLSearchParams(window.location.search);
  return urlParams.get('wasm') === 'true';
 */
}

// Initialize the dashboard when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
  setTimeout(() => {
    const canvasId = getCanvasIdFromUrl();
    const useWASM = isWASMMode();
    
    console.log(`🚀 SDL Canvas Dashboard starting for canvas: ${canvasId} (WASM: ${useWASM})`);
    
    if (useWASM) {
      const dashboard = new WASMDashboard(canvasId, true);
      dashboard.initialize();
    } else {
      const dashboard = new Dashboard(canvasId);
      dashboard.initialize();
    }
  }, 100);
});

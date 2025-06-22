import './style.css';
import { Dashboard } from './dashboard.js';

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

// Initialize the dashboard when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
  const canvasId = getCanvasIdFromUrl();
  console.log(`🚀 SDL Canvas Dashboard starting for canvas: ${canvasId}`);
  new Dashboard(canvasId);
});
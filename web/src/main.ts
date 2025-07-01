import './style.css';
import { Dashboard } from './dashboard.js';
import { WASMDashboard } from './wasm-dashboard.js';
import { SystemDetailsPage } from './system-details-page.js';
import { initializeSystemListing } from './system-listing-handlers.js';

// Initialize the application when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
  // Get page type from data attribute set by server
  const appElement = document.getElementById('app');
  if (!appElement) return;
  
  const pageType = appElement.dataset.pageType;
  const pageData = appElement.dataset.pageData ? JSON.parse(appElement.dataset.pageData) : {};
  
  console.log(`🚀 SDL Canvas loading page type: ${pageType}`);
  
  switch (pageType) {
    case 'canvas-dashboard':
      // Traditional dashboard for full IDE experience
      const canvasId = pageData.canvasId || 'default';
      const useServerCanvas = pageData.mode !== 'wasm';
      
      console.log(`Canvas: ${canvasId} (Mode: ${useServerCanvas ? "ServerRuntime" : "WASMRuntime"})`);
      
      if (useServerCanvas) {
        const dashboard = new Dashboard(canvasId);
        dashboard.initialize();
      } else {
        const dashboard = new WASMDashboard(canvasId);
        dashboard.initialize();
      }
      break;
      
    case 'system-details':
      // System details page for focused editing
      const detailsPage = new SystemDetailsPage(pageData);
      detailsPage.initialize();
      break;
      
    case 'system-listing':
      // System listing is server-rendered, just attach event handlers
      initializeSystemListing();
      break;
      
    default:
      console.error(`Unknown page type: ${pageType}`);
  }
});

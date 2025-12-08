import './style.css';
import { ThemeManager } from '@panyam/tsappkit';
import { CanvasViewerPageDockView } from './pages/CanvasViewerPage';
import { initializeSystemListing } from './system-listing-handlers.js';

// Initialize theme from tsappkit (supplements inline script in BasePage.html)
ThemeManager.init();

// Get page type from window.sdlPageData (set by server template)
const pageData = (window as any).sdlPageData || {};
const pageType = pageData.pageType;

console.log(`[SDL] Page type: ${pageType}`, pageData);

// Initialize based on page type
switch (pageType) {
  case 'canvas-dashboard':
    // Use LCMComponent lifecycle via loadAfterPageLoaded
    CanvasViewerPageDockView.loadAfterPageLoaded('canvasViewerPage', CanvasViewerPageDockView, 'CanvasViewerPageDockView');
    break;

  case 'system-listing':
    // System listing is server-rendered, just attach event handlers
    document.addEventListener('DOMContentLoaded', () => {
      initializeSystemListing();
    });
    break;

  case 'canvas-listing':
    // Canvas listing is server-rendered, no JS initialization needed
    break;

  default:
    // Unknown or no page type - likely a server-rendered page
    console.log(`[SDL] No JS initialization for page type: ${pageType || 'none'}`);
}

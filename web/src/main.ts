import './style.css';
import { ThemeManager } from '@panyam/tsappkit';
import { CanvasViewerPageDockView } from './pages/CanvasViewerPage';

// Initialize theme from tsappkit (supplements inline script in BasePage.html)
ThemeManager.init();

// Get page type from window.sdlPageData (set by server template)
const pageData = (window as any).sdlPageData || {};
const pageType = pageData.pageType;

console.log(`[SDL] Page type: ${pageType}`, pageData);

// Initialize based on page type
switch (pageType) {
  case 'canvas-dashboard':
    // Workspace IDE — uses LCMComponent lifecycle from tsappkit
    CanvasViewerPageDockView.loadAfterPageLoaded('canvasViewerPage', CanvasViewerPageDockView, 'CanvasViewerPageDockView');
    break;

  case 'workspace-listing':
  case 'workspace-create':
    // Server-rendered pages — no JS initialization needed
    break;

  default:
    console.log(`[SDL] No JS initialization for page type: ${pageType || 'none'}`);
}

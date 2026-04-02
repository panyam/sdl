import './style.css';
import { ThemeManager } from '@panyam/tsappkit';
import { WorkspaceViewerPageDockView } from './pages/WorkspaceViewerPage';

// Initialize theme from tsappkit (supplements inline script in BasePage.html)
ThemeManager.init();

// Get page type from window.sdlPageData (set by server template)
const pageData = (window as any).sdlPageData || {};
const pageType = pageData.pageType;

console.log(`[SDL] Page type: ${pageType}`, pageData);

// Initialize based on page type
switch (pageType) {
  case 'workspace-dashboard':
    // Workspace IDE — uses LCMComponent lifecycle from tsappkit
    WorkspaceViewerPageDockView.loadAfterPageLoaded('workspaceViewerPage', WorkspaceViewerPageDockView, 'WorkspaceViewerPageDockView');
    break;

  case 'workspace-listing':
  case 'workspace-create':
    // Server-rendered pages — no JS initialization needed
    break;

  default:
    console.log(`[SDL] No JS initialization for page type: ${pageType || 'none'}`);
}

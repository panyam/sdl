import './style.css';
import { Dashboard } from './dashboard.js';
import { WASMDashboard } from './wasm-dashboard.js';
import { SystemDetailsPage } from './system-details-page.js';

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
      attachSystemListingHandlers();
      break;
      
    default:
      console.error(`Unknown page type: ${pageType}`);
  }
});

// Attach event handlers for server-rendered system listing page
function attachSystemListingHandlers(): void {
  // Search functionality
  const searchInput = document.getElementById('search-input') as HTMLInputElement;
  if (searchInput) {
    searchInput.addEventListener('input', (e) => {
      const query = (e.target as HTMLInputElement).value.toLowerCase();
      filterSystems(query);
    });
  }

  // Filter buttons
  document.querySelectorAll('.filter-btn').forEach(btn => {
    btn.addEventListener('click', (e) => {
      const filter = (e.target as HTMLElement).dataset.filter || 'all';
      setActiveFilter(e.target as HTMLElement);
      filterByDifficulty(filter);
    });
  });
}

function filterSystems(query: string): void {
  const cards = document.querySelectorAll('.system-card') as NodeListOf<HTMLElement>;
  
  cards.forEach(card => {
    const name = card.dataset.name?.toLowerCase() || '';
    const description = card.dataset.description?.toLowerCase() || '';
    const tags = card.dataset.tags?.toLowerCase() || '';
    
    const matchesSearch = 
      name.includes(query) ||
      description.includes(query) ||
      tags.includes(query);
    
    card.style.display = matchesSearch ? 'block' : 'none';
  });
}

function filterByDifficulty(difficulty: string): void {
  const cards = document.querySelectorAll('.system-card') as NodeListOf<HTMLElement>;
  
  cards.forEach(card => {
    if (difficulty === 'all') {
      card.style.display = 'block';
    } else {
      card.style.display = card.dataset.difficulty === difficulty ? 'block' : 'none';
    }
  });
}

function setActiveFilter(activeBtn: HTMLElement): void {
  document.querySelectorAll('.filter-btn').forEach(btn => {
    btn.classList.remove('active');
  });
  activeBtn.classList.add('active');
}

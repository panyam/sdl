/**
 * Enhanced handlers for the system listing page
 */

interface SystemCardData {
  element: HTMLElement;
  id: string;
  name: string;
  description: string;
  tags: string[];
  difficulty: string;
  category: string;
  lastUpdated?: Date;
}

class SystemListingHandlers {
  private systems: SystemCardData[] = [];
  private currentFilter: string = 'all';
  private currentSearch: string = '';
  private currentSort: 'name' | 'difficulty' | 'updated' = 'name';
  private currentTheme: 'light' | 'dark' | 'system' = 'system';

  constructor() {
    this.initializeSystems();
    this.attachEventHandlers();
    this.initializeTheme();
  }

  private initializeSystems(): void {
    const cards = document.querySelectorAll('.system-card') as NodeListOf<HTMLElement>;
    
    this.systems = Array.from(cards).map(card => ({
      element: card,
      id: card.dataset.id || '',
      name: card.dataset.name || '',
      description: card.dataset.description || '',
      tags: (card.dataset.tags || '').trim().split(' ').filter(t => t),
      difficulty: card.dataset.difficulty || '',
      category: card.dataset.category || '',
      lastUpdated: card.dataset.lastUpdated ? new Date(card.dataset.lastUpdated) : undefined
    }));
  }

  private attachEventHandlers(): void {
    // Search functionality
    const searchInput = document.getElementById('search-input') as HTMLInputElement;
    if (searchInput) {
      searchInput.addEventListener('input', (e) => {
        this.currentSearch = (e.target as HTMLInputElement).value.toLowerCase();
        this.applyFilters();
      });
    }

    // Filter buttons
    document.querySelectorAll('.filter-btn').forEach(btn => {
      btn.addEventListener('click', (e) => {
        const filter = (e.target as HTMLElement).dataset.filter || 'all';
        this.currentFilter = filter;
        this.setActiveFilter(e.target as HTMLElement);
        this.applyFilters();
      });
    });

    // Sort dropdown
    const sortSelect = document.getElementById('sort-select') as HTMLSelectElement;
    if (sortSelect) {
      sortSelect.addEventListener('change', (e) => {
        this.currentSort = (e.target as HTMLSelectElement).value as any;
        this.applySorting();
      });
    }

    // Theme switcher
    const themeButtons = document.querySelectorAll('[data-theme]');
    themeButtons.forEach(btn => {
      btn.addEventListener('click', (e) => {
        const theme = (e.target as HTMLElement).dataset.theme as any;
        this.setTheme(theme);
      });
    });
  }

  private applyFilters(): void {
    this.systems.forEach(system => {
      const matchesSearch = this.matchesSearch(system);
      const matchesFilter = this.matchesFilter(system);
      
      system.element.style.display = matchesSearch && matchesFilter ? '' : 'none';
    });

    // Update visible count
    const visibleCount = this.systems.filter(s => s.element.style.display !== 'none').length;
    this.updateResultsCount(visibleCount);
  }

  private matchesSearch(system: SystemCardData): boolean {
    if (!this.currentSearch) return true;
    
    return (
      system.name.toLowerCase().includes(this.currentSearch) ||
      system.description.toLowerCase().includes(this.currentSearch) ||
      system.tags.some(tag => tag.toLowerCase().includes(this.currentSearch)) ||
      system.category.toLowerCase().includes(this.currentSearch)
    );
  }

  private matchesFilter(system: SystemCardData): boolean {
    if (this.currentFilter === 'all') return true;
    return system.difficulty === this.currentFilter;
  }

  private applySorting(): void {
    const container = document.getElementById('systems-grid');
    if (!container) return;

    const sorted = [...this.systems].sort((a, b) => {
      switch (this.currentSort) {
        case 'name':
          return a.name.localeCompare(b.name);
        case 'difficulty':
          const diffOrder: Record<string, number> = { 'beginner': 0, 'intermediate': 1, 'advanced': 2 };
          return (diffOrder[a.difficulty] || 0) - (diffOrder[b.difficulty] || 0);
        case 'updated':
          if (!a.lastUpdated || !b.lastUpdated) return 0;
          return b.lastUpdated.getTime() - a.lastUpdated.getTime();
        default:
          return 0;
      }
    });

    // Reorder DOM elements
    sorted.forEach(system => {
      container.appendChild(system.element);
    });
  }

  private setActiveFilter(activeBtn: HTMLElement): void {
    document.querySelectorAll('.filter-btn').forEach(btn => {
      btn.classList.remove('active');
    });
    activeBtn.classList.add('active');
  }

  private updateResultsCount(count: number): void {
    const countElement = document.getElementById('results-count');
    if (countElement) {
      countElement.textContent = `${count} system${count !== 1 ? 's' : ''} found`;
    }
  }

  // Theme management
  private initializeTheme(): void {
    // Check localStorage for saved theme
    const savedTheme = localStorage.getItem('sdl-theme') as any;
    if (savedTheme && ['light', 'dark', 'system'].includes(savedTheme)) {
      this.currentTheme = savedTheme;
    }

    // Check system preference
    if (this.currentTheme === 'system') {
      const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
      this.applyTheme(prefersDark ? 'dark' : 'light');
    } else {
      this.applyTheme(this.currentTheme);
    }

    // Listen for system theme changes
    window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', (e) => {
      if (this.currentTheme === 'system') {
        this.applyTheme(e.matches ? 'dark' : 'light');
      }
    });
  }

  private setTheme(theme: 'light' | 'dark' | 'system'): void {
    this.currentTheme = theme;
    localStorage.setItem('sdl-theme', theme);

    if (theme === 'system') {
      const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
      this.applyTheme(prefersDark ? 'dark' : 'light');
    } else {
      this.applyTheme(theme);
    }

    // Update active theme button
    document.querySelectorAll('[data-theme]').forEach(btn => {
      btn.classList.toggle('active', btn.getAttribute('data-theme') === theme);
    });
  }

  private applyTheme(theme: 'light' | 'dark'): void {
    if (theme === 'dark') {
      document.documentElement.classList.add('dark');
    } else {
      document.documentElement.classList.remove('dark');
    }
  }
}

// Export for use in main.ts
export function initializeSystemListing(): void {
  new SystemListingHandlers();
}
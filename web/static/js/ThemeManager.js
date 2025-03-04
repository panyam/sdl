class ThemeManager {
  constructor() {
    this.themeToggleBtn = document.getElementById('theme-toggle');
    this.darkIcon = document.getElementById('theme-toggle-dark-icon');
    this.lightIcon = document.getElementById('theme-toggle-light-icon');
    this.systemIcon = document.getElementById('theme-toggle-system-icon');

    // initialize
    this.setup()
  }

  setup() {
    // Theme handling
    const currentTheme = localStorage.theme || 'system';

    this.setTheme(currentTheme);
    this.themeToggleBtn.addEventListener('click', () => {
        const current = localStorage.theme || 'system';
        const nextTheme = current === 'light' ? 'dark' : 
                        current === 'dark' ? 'system' : 'light';
        this.setTheme(nextTheme);
    });

    window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', () => {
        if (!localStorage.theme) {
            setTheme('system');
        }
    });

    // Mobile menu handling
    const mobileMenuBtn = document.getElementById('mobile-menu-button');
    const mobileMenu = document.getElementById('mobile-menu');

    mobileMenuBtn.addEventListener('click', () => {
        mobileMenu.classList.toggle('hidden');
    });

    document.addEventListener('click', (event) => {
        if (!mobileMenuBtn.contains(event.target) && !mobileMenu.contains(event.target)) {
            mobileMenu.classList.add('hidden');
        }
    });

    window.addEventListener('resize', () => {
        if (window.innerWidth >= 768) {
            mobileMenu.classList.add('hidden');
        }
    });
  }

  updateThemeIcons(theme) {
    this.darkIcon.classList.add('hidden');
    this.lightIcon.classList.add('hidden');
    this.systemIcon.classList.add('hidden');

    switch(theme) {
        case 'dark':
            this.lightIcon.classList.remove('hidden');
            break;
        case 'light':
            this.darkIcon.classList.remove('hidden');
            break;
        default:
            this.systemIcon.classList.remove('hidden');
    }
  }

  setTheme(theme) {
    if (theme === 'dark') {
        document.documentElement.classList.add('dark');
        localStorage.theme = 'dark';
    } else if (theme === 'light') {
        document.documentElement.classList.remove('dark');
        localStorage.theme = 'light';
    } else {
        localStorage.removeItem('theme');
        if (window.matchMedia('(prefers-color-scheme: dark)').matches) {
            document.documentElement.classList.add('dark');
        } else {
            document.documentElement.classList.remove('dark');
        }
    }
    this.updateThemeIcons(theme);
  }
}

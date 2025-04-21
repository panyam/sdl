// components/theme.ts

/**
 * Theme management for LeetCoach application
 */
export class ThemeManager {
    // Theme options
    static LIGHT = 'light';
    static DARK = 'dark';
    static SYSTEM = 'system';

    // Icons for each theme state
    static readonly LIGHT_ICON_SVG = `<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-full h-full"><path stroke-linecap="round" stroke-linejoin="round" d="M12 3v2.25m6.364.386l-1.591 1.591M21 12h-2.25m-.386 6.364l-1.591-1.591M12 18.75V21m-4.773-4.227l-1.591 1.591M5.25 12H3m4.227-4.773L5.636 5.636M15.75 12a3.75 3.75 0 11-7.5 0 3.75 3.75 0 017.5 0z" /></svg>`; // Sun icon
    static readonly DARK_ICON_SVG = `<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-full h-full"><path stroke-linecap="round" stroke-linejoin="round" d="M21.752 15.002A9.718 9.718 0 0118 15.75c-5.385 0-9.75-4.365-9.75-9.75 0-1.33.266-2.597.748-3.752A9.753 9.753 0 003 11.25C3 16.635 7.365 21 12.75 21a9.753 9.753 0 009.002-5.998z" /></svg>`; // Moon icon
    static readonly SYSTEM_ICON_SVG = `<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-full h-full"><path stroke-linecap="round" stroke-linejoin="round" d="M9 17.25v1.007a3 3 0 01-.879 2.122L7.5 21h9l-.621-.621A3 3 0 0115 18.257V17.25m6-12V15a2.25 2.25 0 01-2.25 2.25H5.25A2.25 2.25 0 013 15V5.25m18 0A2.25 2.25 0 0018.75 3H5.25A2.25 2.25 0 003 5.25m18 0V12a2.25 2.25 0 01-2.25 2.25H5.25A2.25 2.25 0 013 12V5.25" /></svg>`; // Computer icon

    /**
     * Initialize theme based on saved preference or system default
     */
    static initialize(): void {
        const savedTheme = localStorage.getItem('theme');

        if (savedTheme === ThemeManager.DARK ||
            (!savedTheme && window.matchMedia('(prefers-color-scheme: dark)').matches)) {
            document.documentElement.classList.add('dark');
        } else {
            document.documentElement.classList.remove('dark');
        }
        // No need to update icon here, DetailPage will handle it on init
    }

    /**
     * Set theme and save preference
     */
    static setTheme(theme: string): void {
        if (theme === ThemeManager.SYSTEM) {
            localStorage.removeItem('theme');
            if (window.matchMedia('(prefers-color-scheme: dark)').matches) {
                document.documentElement.classList.add('dark');
            } else {
                document.documentElement.classList.remove('dark');
            }
        } else if (theme === ThemeManager.DARK) {
            localStorage.setItem('theme', ThemeManager.DARK);
            document.documentElement.classList.add('dark');
        } else { // LIGHT
            localStorage.setItem('theme', ThemeManager.LIGHT);
            document.documentElement.classList.remove('dark');
        }
        // Inform about potential TinyMCE redraw needed (optional, depends on how sensitive TinyMCE is)
        // console.log("Theme set, TinyMCE might need skin update if active editors exist.");
    }

    /**
     * Get current theme setting (light, dark, or system)
     */
    static getCurrentThemeSetting(): string {
        return localStorage.getItem('theme') || ThemeManager.SYSTEM;
    }

    /**
     * Gets the *next* theme in the cycle: Light -> Dark -> System -> Light ...
     */
    static getNextTheme(currentSetting: string): string {
        if (currentSetting === ThemeManager.LIGHT) {
            return ThemeManager.DARK;
        } else if (currentSetting === ThemeManager.DARK) {
            return ThemeManager.SYSTEM;
        } else { // SYSTEM or unknown goes to LIGHT
            return ThemeManager.LIGHT;
        }
    }

    /**
     * Gets the appropriate SVG icon string for a given theme setting.
     */
    static getIconSVG(themeSetting: string): string {
        switch (themeSetting) {
            case ThemeManager.LIGHT: return ThemeManager.LIGHT_ICON_SVG;
            case ThemeManager.DARK: return ThemeManager.DARK_ICON_SVG;
            case ThemeManager.SYSTEM:
            default: return ThemeManager.SYSTEM_ICON_SVG;
        }
    }

    /**
     * Gets a user-friendly label for the theme setting.
     */
    static getThemeLabel(themeSetting: string): string {
        switch (themeSetting) {
            case ThemeManager.LIGHT: return "Light Mode";
            case ThemeManager.DARK: return "Dark Mode";
            case ThemeManager.SYSTEM:
            default: return "System Default";
        }
    }


    /**
     * Initialize the ThemeManager (no instance needed for static methods)
     * We keep init() if other instance logic might be added later,
     * but it doesn't return anything useful right now.
     */
    public static init(): void {
        // Static initialization happens directly or via initialize()
        // No instance is created or needed for these static methods.
        ThemeManager.initialize();
    }
}

// No need to call initialize here, DetailPage will handle it
// document.addEventListener('DOMContentLoaded', () => {
//     ThemeManager.initialize();
// });

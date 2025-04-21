// components/DetailPage.ts

import { ThemeManager } from './ThemeManager';
import { DocumentTitle } from './DocumentTitle';
import { Modal } from './Modal';
import { SectionManager } from './SectionManager';
import { ToastManager } from './ToastManager';
import { TableOfContents } from './TableOfContents';
import { LeetCoachDocument } from './types';
import { DOCUMENT } from "./samples";

/**
 * Main application initialization
 */
class LeetCoachApp {
    // Keep other component properties...
    private themeManager: typeof ThemeManager | null = null; // Use typeof for static class
    private documentTitle: DocumentTitle | null = null;
    private modal: Modal | null = null;
    private sectionManager: SectionManager | null = null;
    private toastManager: ToastManager | null = null;
    private tableOfContents: TableOfContents | null = null;

    // New elements for theme toggle
    private themeToggleButton: HTMLButtonElement | null = null;
    private themeToggleIcon: HTMLElement | null = null;

    constructor() {
        this.initializeComponents();
        this.bindEvents();
        this.loadInitialState(); // Load document and set initial theme icon
    }

    /** Initialize all application components */
    private initializeComponents(): void {
        ThemeManager.init(); // Static init call
        this.modal = Modal.init();
        this.toastManager = ToastManager.init();
        this.documentTitle = DocumentTitle.init();
        this.sectionManager = SectionManager.init();
        this.tableOfContents = TableOfContents.init({
            onAddSectionClick: () => {
                this.sectionManager?.openSectionTypeSelector(null);
            }
        });

        if (this.sectionManager && this.tableOfContents) {
            this.sectionManager.setTocComponent(this.tableOfContents);
        }

        // Find new theme toggle elements
        this.themeToggleButton = document.getElementById('theme-toggle-button') as HTMLButtonElement;
        this.themeToggleIcon = document.getElementById('theme-toggle-icon');

        if (!this.themeToggleButton || !this.themeToggleIcon) {
            console.warn("Theme toggle button or icon element not found in Header.");
        }

        console.log('LeetCoach application initialized');
    }

    /** Bind application-level events */
    private bindEvents(): void {
        // --- Add new theme toggle button event ---
        if (this.themeToggleButton) {
            this.themeToggleButton.addEventListener('click', this.handleThemeToggleClick.bind(this));
        }

        // --- Mobile menu toggle ---
        const mobileMenuButton = document.getElementById('mobile-menu-button');
        if (mobileMenuButton) {
            mobileMenuButton.addEventListener('click', () => {
                this.tableOfContents?.toggleDrawer();
            });
        }

        // --- Save button ---
        const saveButton = document.querySelector('header button.bg-blue-600');
        if (saveButton) {
            saveButton.addEventListener('click', this.saveDocument.bind(this));
        }

        // --- Export button ---
        const exportButton = document.querySelector('header button.bg-gray-200');
        if (exportButton) {
            exportButton.addEventListener('click', this.exportDocument.bind(this));
        }
    }

    /** Load document data and set initial UI states */
    private loadInitialState(): void {
        this.loadDocument(DOCUMENT); // Load the document data
        this.updateThemeButtonState(); // Set the initial theme icon/label
    }


    /** Handles click on the new theme toggle button */
    private handleThemeToggleClick(): void {
        const currentSetting = ThemeManager.getCurrentThemeSetting();
        const nextSetting = ThemeManager.getNextTheme(currentSetting);
        ThemeManager.setTheme(nextSetting);
        this.updateThemeButtonState(nextSetting); // Update icon to reflect the *new* state

        // Optional: Show toast feedback
        // this.toastManager?.showToast('Theme Changed', `Switched to ${ThemeManager.getThemeLabel(nextSetting)}`, 'info', 2000);

        // Notify SectionManager which will, in turn, notify all sections
        this.sectionManager?.notifySectionsOfThemeChange();
    }

    /** Updates the theme toggle button's icon and aria-label */
    private updateThemeButtonState(currentTheme?: string): void {
        if (!this.themeToggleButton || !this.themeToggleIcon) return;

        const themeToDisplay = currentTheme || ThemeManager.getCurrentThemeSetting();
        const iconSVG = ThemeManager.getIconSVG(themeToDisplay);
        const label = `Toggle theme (currently: ${ThemeManager.getThemeLabel(themeToDisplay)})`;

        this.themeToggleIcon.innerHTML = iconSVG;
        this.themeToggleButton.setAttribute('aria-label', label);
        this.themeToggleButton.setAttribute('title', label); // Add tooltip
    }


    /** Load document data into the components */
    public loadDocument(doc: LeetCoachDocument): void {
        console.log("Loading document:", doc.metadata.id);
        if (this.documentTitle) {
            this.documentTitle.setTitle(doc.title);
        }
        if (this.sectionManager) {
            // loadSections will now also trigger the TOC update via the connected component
            this.sectionManager.loadSections(doc.sections);
        } else {
            console.error("SectionManager not initialized, cannot load document sections.");
        }
    }

    /** Save document */
    private saveDocument(): void {
        console.log("Attempting to save document...");
        // ... (save logic remains the same) ...
        if (!this.documentTitle || !this.sectionManager || !this.toastManager) {
            console.error("Cannot save: Core components not initialized.");
            this.toastManager?.showToast('Save Failed', 'Could not save document.', 'error');
            return;
        }

        const currentTimestamp = new Date().toISOString();
        const documentData: LeetCoachDocument = {
            metadata: {
                id: "doc-placeholder-uuid", // Replace with actual ID
                schemaVersion: "1.0",
                lastSavedAt: currentTimestamp
            },
            title: this.documentTitle.getTitle(),
            sections: this.sectionManager.getDocumentSections() // Get data from manager
        };

        try {
            const jsonString = JSON.stringify(documentData, null, 2);
            console.log("--- LeetCoach Document State (JSON) ---");
            console.log(jsonString);
            console.log("---------------------------------------");
            this.toastManager.showToast('Document Saved', 'Content logged to console.', 'success');
            // Update last saved time display in DocumentTitle component
            this.documentTitle.updateLastSavedTime();

        } catch (error) {
            console.error("Error serializing document data:", error);
            this.toastManager.showToast('Save Error', 'Could not prepare data for saving.', 'error');
        }
    }

    /** Export document (Placeholder) */
    private exportDocument(): void {
        if (this.toastManager) {
            this.toastManager.showToast('Export started', 'Your document is being prepared for export.', 'info');
            setTimeout(() => {
                this.toastManager?.showToast('Export complete', 'Document export simulation finished.', 'success');
            }, 1500);
        }
    }

    /** Initialize the application */
    public static init(): LeetCoachApp {
        return new LeetCoachApp();
    }
}

// Initialize the application when the DOM is fully loaded
document.addEventListener('DOMContentLoaded', () => {
    const lc = LeetCoachApp.init();
    // Sample document loading is now handled within LeetCoachApp constructor/init process
    // lc.loadDocument(DOCUMENT);
});

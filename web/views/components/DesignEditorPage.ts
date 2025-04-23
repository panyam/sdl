// components/DesignEditorPage.ts

import { ThemeManager } from './ThemeManager';
import { DocumentTitle } from './DocumentTitle';
import { Modal } from './Modal';
import { SectionManager } from './SectionManager';
import { ToastManager } from './ToastManager';
import { TableOfContents } from './TableOfContents';
import { LeetCoachDocument, DocumentSection } from './types';
import { DesignApi } from './Api';
import { V1GetDesignResponse } from './apiclient';

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

    // Store the current design ID
    private currentDesignId: string | null = null;

    constructor() {
        this.initializeComponents();
        this.bindEvents();
        this.loadInitialState(); // Load document and set initial theme icon
    }

    /** Initialize all application components */
    private initializeComponents(): void {
        const designId = (document.getElementById("designIdInput") as HTMLInputElement).value.trim();
        ThemeManager.init(); // Static init call
        this.modal = Modal.init();
        this.toastManager = ToastManager.init();
        this.documentTitle = new DocumentTitle(designId);
        this.sectionManager = SectionManager.init();
        this.tableOfContents = TableOfContents.init({
            onAddSectionClick: () => {
                // We need the last section ID to add after it.
                // Get sorted section data from SectionManager.
                const sections = this.sectionManager?.getDocumentSections() || []; // Get sorted data
                const lastSectionId = sections.length > 0 ? sections[sections.length - 1].id : null;
                this.sectionManager?.openSectionTypeSelector(lastSectionId, 'after'); // Use 'after' last section
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
            saveButton.addEventListener('click', this.saveDocument.bind(this)); // Save button is now a placeholder
        }

        // --- Export button ---
        const exportButton = document.querySelector('header button.bg-gray-200');
        if (exportButton) {
            exportButton.addEventListener('click', this.exportDocument.bind(this)); // Export remains placeholder
        }
    }

    /** Load document data and set initial UI states */
    private loadInitialState(): void {
        this.updateThemeButtonState(); // Set the initial theme icon/label

        // --- NEW: Get Design ID and Load Data ---
        // Assume the design ID is embedded in the body's data attribute by the Go template
        const designId = (document.getElementById("designIdInput") as HTMLInputElement).value.trim();
        if (designId.length > 0) {
            this.currentDesignId = designId;
            console.log(`Found Design ID: ${this.currentDesignId}. Loading data...`);
            this.loadDesignData(this.currentDesignId);
        } else {
            console.error("Design ID not found in body data attribute (data-design-id). Cannot load document.");
            this.toastManager?.showToast("Error", "Could not load document: Design ID missing.", "error");
            // Optionally, load empty state or redirect
            this.sectionManager?.handleEmptyState(); // Show empty state if no ID
        }
        // --- End NEW ---
    }

    /**
     * NEW: Fetches design metadata (title, section IDs/metadata) from the API.
     */
    private async loadDesignData(designId: string): Promise<void> {
        if (!this.documentTitle || !this.sectionManager || !this.toastManager) {
            console.error("Cannot load design data: Core components not initialized.");
            return;
        }

        // Optional: Show loading state
        // this.showLoadingIndicator(true);

        try {
            const response: V1GetDesignResponse = await DesignApi.designServiceGetDesign({
                id: designId,
                includeSectionMetadata: true // Request metadata for potential use later
            });

            console.log("API Response (GetDesign):", response);

            if (response.design) {
                // Update Document Title
                this.documentTitle.setTitle(response.design.name || "Untitled Design");

                // Pass Section Info to SectionManager
                const sectionIds = response.design.sectionIds || [];
                const sectionsMetadata = response.sectionsMetadata || null; // API might return undefined
                this.sectionManager.setInitialSectionInfo(sectionIds, sectionsMetadata);

                // --- Trigger Step 1.2 (to be implemented next) ---
                // At this point, we would trigger the loading of individual sections
                // based on the sectionIds stored in SectionManager.
                // For now, the page will remain empty section-wise.
                 console.log("Step 1.1 Complete: Design metadata loaded. Section content loading deferred.");
                 // If sectionIds is empty, handle empty state now. If not, handleEmptyState
                 this.loadSectionsContent(designId, sectionIds); // <-- NEW: Trigger section loading

                 // will be called later by SectionManager after loading sections.
                 if (sectionIds.length === 0) {
                     this.sectionManager.handleEmptyState();
                 }

            } else {
                throw new Error("API response missing design object.");
            }

        } catch (error: any) {
            console.error("Error loading design data:", error);
            const errorMsg = error.message || "Failed to fetch design details from the server.";
            this.toastManager.showToast("Load Failed", errorMsg, "error");
            // Show empty state on error
             this.sectionManager.handleEmptyState();

        } finally {
            // Optional: Hide loading state
            // this.showLoadingIndicator(false);
        }
    }

    /**
     * NEW: Fetches content for each section ID and loads them into SectionManager.
     */
    private async loadSectionsContent(designId: string, sectionIds: string[]): Promise<void> {
        if (!this.sectionManager || !this.toastManager) return;
        if (sectionIds.length === 0) {
            console.log("No section IDs found, skipping content loading.");
            // Empty state already handled by loadDesignData
            return;
        }

        console.log(`Step 1.2: Loading content for ${sectionIds.length} sections...`);
        // Optional: Show loading state for sections
        // this.showSectionLoadingIndicator(true);

        try {
            // Pass the IDs to SectionManager to handle the fetching and loading
            await this.sectionManager.loadSectionContentsByIds(designId, sectionIds);
            console.log("Step 1.2 Complete: Section content loading finished.");

        } catch (error: any) {
            console.error("Error during section content loading process:", error);
            this.toastManager.showToast("Section Load Failed", "Could not load content for some sections.", "error");
            // Ensure empty state is handled if loading completely fails or results in zero sections
            this.sectionManager.handleEmptyState();
        } finally {
            // Optional: Hide loading state for sections
            // this.showSectionLoadingIndicator(false);
        }
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
    // public loadDocument(doc: LeetCoachDocument): void { // Keep signature if needed? No, API drives load now.
    //     console.log("Loading document:", doc.metadata.id);
    //     if (this.documentTitle) {
    //         this.documentTitle.setTitle(doc.title);
    //     }
    //     if (this.sectionManager) {
    //         // loadSections will now also trigger the TOC update via the connected component
    //         // THIS IS REPLACED BY API LOADING
    //         // this.sectionManager.loadSections(doc.sections);
    //     } else {
    //         console.error("SectionManager not initialized, cannot load document sections.");
    //     }
    // }

    /** Save document (Placeholder - needs full implementation later) */
    private saveDocument(): void {
        console.log("Save button clicked (Placeholder - Requires API integration for full save)");
        if (!this.currentDesignId || !this.documentTitle || !this.sectionManager || !this.toastManager) {
            console.error("Cannot save: Core components not initialized or design ID missing.");
            this.toastManager?.showToast('Save Failed', 'Could not save document.', 'error');
            return;
        }
        // This full save logic will be replaced by incremental saves triggered by component callbacks
        this.toastManager.showToast('Save Action', 'Incremental saves handle updates. Full save TBD.', 'info');

        /*
        // --- Example of how full save *might* look (but prefer incremental) ---
        const currentTimestamp = new Date().toISOString();
        const documentData: LeetCoachDocument = {
            metadata: {
                id: this.currentDesignId, // Use the actual ID
                schemaVersion: "1.0",
                lastSavedAt: currentTimestamp
            },
            title: this.documentTitle.getTitle(),
            sections: this.sectionManager.getDocumentSections() // Get data from manager
        };
        // Call API to save the full documentData (less ideal than incremental)
        */
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
    // Sample document loading is REMOVED. Loading is triggered by API call in constructor/init.
    // lc.loadDocument(DOCUMENT);
});

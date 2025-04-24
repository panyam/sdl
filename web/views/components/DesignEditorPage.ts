// ./web/views/components/DesignEditorPage.ts

import { ThemeManager } from './ThemeManager';
import { DocumentTitle } from './DocumentTitle';
import { Modal } from './Modal';
import { SectionManager } from './SectionManager';
import { ToastManager } from './ToastManager';
import { TableOfContents } from './TableOfContents';
import { LeetCoachDocument, DocumentSection, SectionType, SectionData } from './types'; // Added SectionData
import { DesignApi } from './Api';
import { V1GetDesignResponse, V1Section } from './apiclient'; // Added V1Section
import { mapApiSectionTypeToFrontend } from './converters'; // Added converter
import { BaseSection } from './BaseSection'; // Added BaseSection

/**
 * Main application initialization
 */
class DesignEditorPage {
    private themeManager: typeof ThemeManager | null = null;
    private documentTitle: DocumentTitle | null = null;
    private modal: Modal | null = null;
    private sectionManager: SectionManager | null = null;
    private toastManager: ToastManager | null = null;
    private tableOfContents: TableOfContents | null = null;

    private themeToggleButton: HTMLButtonElement | null = null;
    private themeToggleIcon: HTMLElement | null = null;

    private currentDesignId: string | null = null;
    private isLoadingDesign: boolean = false; // Loading state

    constructor() {
        this.initializeComponents();
        this.bindEvents();
        this.loadInitialState();
    }

    private initializeComponents(): void {
        const designIdInput = document.getElementById("designIdInput") as HTMLInputElement | null;
        const designId = designIdInput?.value.trim() || null; // Allow null if input not found/empty

        ThemeManager.init();
        this.modal = Modal.init();
        this.toastManager = ToastManager.init();
        this.documentTitle = new DocumentTitle(designId); // Pass initial ID
        this.sectionManager = new SectionManager(designId); // Pass initial ID
        this.tableOfContents = TableOfContents.init({
            onAddSectionClick: () => {
                this.sectionManager?.openSectionTypeSelector(this.getLastSectionId(), 'after');
            }
        });

        if (this.sectionManager && this.tableOfContents) {
            this.sectionManager.setTocComponent(this.tableOfContents);
        }

        this.themeToggleButton = document.getElementById('theme-toggle-button') as HTMLButtonElement;
        this.themeToggleIcon = document.getElementById('theme-toggle-icon');

        if (!this.themeToggleButton || !this.themeToggleIcon) {
            console.warn("Theme toggle button or icon element not found in Header.");
        }

        console.log('LeetCoach application initialized');
    }

    private bindEvents(): void {
        if (this.themeToggleButton) {
            this.themeToggleButton.addEventListener('click', this.handleThemeToggleClick.bind(this));
        }

        const mobileMenuButton = document.getElementById('mobile-menu-button');
        if (mobileMenuButton) {
            mobileMenuButton.addEventListener('click', () => {
                this.tableOfContents?.toggleDrawer();
            });
        }

        const saveButton = document.querySelector('header button.bg-blue-600');
        if (saveButton) {
            saveButton.addEventListener('click', this.saveDocument.bind(this));
        }

       document.addEventListener('click', (e: MouseEvent) => {
            const target = e.target as HTMLElement;
            const sectionTypeOption = target.closest('.section-type-option, button.section-type-option');

            if (sectionTypeOption && this.modal?.getCurrentTemplate() === 'section-type-selector') {
                this.handleSectionTypeSelectionFromModal(sectionTypeOption);
            }
        });

        const exportButton = document.querySelector('header button.bg-gray-200');
        if (exportButton) {
            exportButton.addEventListener('click', this.exportDocument.bind(this));
        }
    }

     /**
      * Handles the click on a section type button within the modal.
      * Calls the SectionManager to initiate the API call and subsequent creation.
      */
     private handleSectionTypeSelectionFromModal(buttonElement: Element): void {
         if (!this.sectionManager || !this.modal) return;

         let sectionType: SectionType = 'text';
         const typeText = buttonElement.querySelector('span')?.textContent?.trim().toLowerCase() || '';

         if (typeText === 'drawing') sectionType = 'drawing';
         else if (typeText === 'plot') sectionType = 'plot';

         const modalData = this.modal.getCurrentData();
         const relativeToId = modalData?.relativeToId || null;
         const position = modalData?.position || 'after';

         // SectionManager handles API call and creation, including calling setInitialContentAndRender
         this.sectionManager.handleSectionTypeSelection(sectionType, relativeToId, position);

         this.modal.hide();
     }

    /** Load document data and set initial UI states */
    private loadInitialState(): void {
        this.updateThemeButtonState();

        const designIdInput = document.getElementById("designIdInput") as HTMLInputElement | null;
        const designId = designIdInput?.value.trim() || null;

        if (designId) {
            this.currentDesignId = designId;
            console.log(`Found Design ID: ${this.currentDesignId}. Loading data...`);
            this.loadDesignData(this.currentDesignId);
        } else {
            console.error("Design ID input element not found or has no value. Cannot load document.");
            this.toastManager?.showToast("Error", "Could not load document: Design ID missing.", "error");
            this.sectionManager?.handleEmptyState();
        }
    }

    /**
     * Fetches design metadata, initializes section shells, and triggers content loading for each section.
     */
    private async loadDesignData(designId: string): Promise<void> {
        if (!this.documentTitle || !this.sectionManager || !this.toastManager || this.isLoadingDesign) {
            console.warn("Cannot load design data: Core components not initialized or already loading.");
            return;
        }

        this.isLoadingDesign = true;
        // Optional: Show global loading state
        document.body.classList.add('opacity-50', 'pointer-events-none'); // Example loading state
        console.log(`Step 1.1: Loading design metadata for ${designId}...`);

        try {
            const response: V1GetDesignResponse = await DesignApi.designServiceGetDesign({
                id: designId,
                includeSectionMetadata: true
            });
            console.log("API Response (GetDesign):", response);

            if (response.design) {
                this.documentTitle.setTitle(response.design.name || "Untitled Design");

                let sectionsMeta: SectionData[] = [];
                if (response.sectionsMetadata && response.sectionsMetadata.length > 0) {
                     console.log("Using sectionsMetadata from API response.");
                     sectionsMeta = response.sectionsMetadata.map(apiMeta => ({
                         id: apiMeta.id || '',
                         designId: designId,
                         type: mapApiSectionTypeToFrontend(apiMeta.type),
                         title: apiMeta.title || '',
                         order: apiMeta.order || 0,
                         content: null // Start with null content
                     }));
                } else if (response.design.sectionIds && response.design.sectionIds.length > 0) {
                     console.warn("sectionsMetadata not returned, initializing shells with IDs only.");
                     sectionsMeta = response.design.sectionIds.map((id, index) => ({
                         id: id, designId: designId, type: 'text', title: `Section ${index + 1}`, order: index + 1, content: null
                     }));
                 } else {
                    console.log("Design has no sections.");
                 }

                console.log("Step 1.2: Initializing section shells...");
                // InitializeSections creates the instances and renders the basic structure
                const sectionInstances = this.sectionManager.initializeSections(sectionsMeta);

                console.log(`Step 1.3: Triggering content loading for ${sectionInstances.length} sections...`);
                if (sectionInstances.length > 0) {
                    const loadPromises = sectionInstances.map(instance => {
                        // Wrap loadContent in a try/catch just in case the promise itself rejects unexpectedly
                        // although loadContent already has internal error handling.
                        return instance.loadContent().catch(err => {
                             console.error(`Unhandled error during loadContent for section ${instance.sectionId}:`, err);
                             // Potentially update UI to show a permanent error for this section if needed
                        });
                    });
                    // Wait for all sections to attempt loading. Individual errors are handled within loadContent.
                    await Promise.all(loadPromises);
                    console.log("Step 1.3 Complete: All sections finished loading attempt.");
                } else {
                    console.log("No sections to load content for.");
                }

                // Update empty state *after* attempting to load all sections
                this.sectionManager.handleEmptyState();

            } else {
                throw new Error("API response missing design object.");
            }

        } catch (error: any) {
            console.error("Error loading design data:", error);
            const errorMsg = error.message || "Failed to fetch design details from the server.";
            this.toastManager.showToast("Load Failed", errorMsg, "error");
            this.sectionManager?.handleEmptyState();
        } finally {
            this.isLoadingDesign = false;
            // Optional: Hide global loading state
            document.body.classList.remove('opacity-50', 'pointer-events-none');
            console.log("Design loading process finished.");
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
 
    /** Helper to get the ID of the last section currently managed */
    private getLastSectionId(): string | null {
        const sections = this.sectionManager?.getDocumentSections() || []; // Get sorted data
        return sections.length > 0 ? sections[sections.length - 1].id : null;
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
}

document.addEventListener('DOMContentLoaded', () => {
    const lc = new DesignEditorPage();
});

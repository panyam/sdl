// ./web/views/components/DesignEditorPage.ts

import { ThemeManager } from './ThemeManager';
import { DocumentTitle } from './DocumentTitle';
import { Modal } from './Modal';
import { SectionManager } from './SectionManager';
import { ToastManager } from './ToastManager';
import { TableOfContents } from './TableOfContents';
import { LeetCoachDocument, DocumentSection, SectionType, SectionData } from './types';
import { DesignApi } from './Api';
import { V1GetDesignResponse, V1Section } from './apiclient';
import { convertApiSectionToSectionData } from './converters';
import { BaseSection } from './BaseSection';

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
        this.documentTitle = new DocumentTitle(designId);
        this.sectionManager = new SectionManager(designId);
        this.tableOfContents = TableOfContents.init({
            onAddSectionClick: () => {
                this.sectionManager?.openSectionTypeSelector(this.sectionManager.getLastSectionId(), 'after');
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
        if (!this.documentTitle || !this.sectionManager || !this.toastManager) {
            console.error("Cannot load design data: Core components not initialized.");
            return;
        }

        // TODO: Show global loading indicator
        console.log(`DesignEditorPage: Loading design ${designId}...`);

        try {
            // Step 1: Fetch Design Metadata (including section metadata)
            const response: V1GetDesignResponse = await DesignApi.designServiceGetDesign({
                id: designId,
                includeSectionMetadata: true // Request metadata (id, type, title, order)
            });
            console.log("API Response (GetDesign):", response);

            if (!response.design) {
                throw new Error("API response missing design object.");
            }

            // Update Document Title
            this.documentTitle.setTitle(response.design.name || "Untitled Design");
            this.documentTitle.updateLastSavedTime(new Date(response.design.updatedAt || Date.now())); // Use timestamp from API

            // Step 2: Prepare Metadata for SectionManager
            const sectionsMetadataRaw = response.sectionsMetadata || [];
            const sectionsMetadata: SectionData[] = sectionsMetadataRaw.map(apiSec => ({
                // Convert API metadata to the SectionData structure needed by initializeSections
                // Explicitly set content to null - it will be loaded by the section itself
                id: apiSec.id || '', // Ensure ID exists
                designId: designId, // Add designId here
                type: convertApiSectionToSectionData(apiSec).type, // Use converter for type mapping
                title: apiSec.title || 'Untitled Section',
                order: apiSec.order || 0,
                content: null // Mark content as not loaded yet
            }));

            // Step 3: Initialize Section Shells using SectionManager
            console.log(`DesignEditorPage: Initializing ${sectionsMetadata.length} section shells...`);
            // This creates the instances and DOM, but doesn't load content yet
            const sectionInstances: BaseSection[] = this.sectionManager.initializeSections(sectionsMetadata);

            // Step 4: Trigger Content Loading for Each Section Instance
            if (sectionInstances.length > 0) {
                console.log(`DesignEditorPage: Triggering loadContent() for ${sectionInstances.length} sections...`);
                const loadPromises = sectionInstances.map(instance =>
                    instance.loadContent().catch(err => {
                        // Catch individual load errors here so Promise.allSettled sees them as 'rejected'
                        console.error(`Error caught during loadContent for section ${instance.sectionId}:`, err);
                        return Promise.reject(err); // Ensure it's still treated as a rejection
                    })
                );

                // Wait for all load attempts to complete (successfully or with errors)
                const results = await Promise.allSettled(loadPromises);
                console.log("DesignEditorPage: All section load attempts finished.", results);

                // Optionally: Check results for overall success/failure logging
                const failedLoads = results.filter(r => r.status === 'rejected').length;
                if (failedLoads > 0) {
                     this.toastManager.showToast("Load Warning", `${failedLoads} section(s) failed to load content.`, "warning");
                }
            } else {
                 console.log("DesignEditorPage: No sections found to load content for.");
            }

            // Step 5: Update empty state *after* initialization and load attempts
            this.sectionManager.handleEmptyState(); // Reflects loaded sections (or lack thereof)


        } catch (error: any) {
            console.error("Error loading design data:", error);
            const errorMsg = error.message || "Failed to fetch design details from the server.";
            this.toastManager.showToast("Load Failed", errorMsg, "error");
            // Ensure empty state is shown on a major load failure
            this.sectionManager.handleEmptyState();

        } finally {
            // TODO: Hide global loading indicator
             console.log(`DesignEditorPage: Design loading process complete for ${designId}.`);
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

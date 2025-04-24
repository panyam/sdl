// ./web/views/components/SectionManager.ts

import { Modal } from './Modal';
import { BaseSection } from './BaseSection';
import { TextSection } from './TextSection';
import { DrawingSection } from './DrawingSection';
import { PlotSection } from './PlotSection';
import { DocumentSection, SectionType, TextContent, DrawingContent, PlotContent, SectionData, SectionCallbacks } from './types';
import { TocItemInfo, TableOfContents } from './TableOfContents';
import { V1Section, V1PositionType, /* other models if needed */ } from './apiclient'; // Import V1Section and V1PositionType
import { DesignApi } from './Api'; // Import API client
import { createApiSectionUpdateObject, convertApiSectionToSectionData , mapFrontendContentToApiUpdate, mapFrontendSectionTypeToApi } from './converters'; // Import the update object creator


/**
 * Manages document sections using the BaseSection hierarchy.
 * Collaborates with TableOfContents component for UI updates.
 */
export class SectionManager {
    private designId: string = "";
    private sections: Map<string, BaseSection> = new Map();
    // sectionData now primarily stores metadata managed by SectionManager (title, order).
    // Content is primarily managed within the BaseSection instance's data property after load.
    private sectionData: SectionData[] = [];
    private nextSectionId: number = 1; // Still potentially useful for client-side ID generation fallback? Maybe remove later.
    private sectionsContainer: HTMLElement | null;
    private emptyStateEl: HTMLElement | null;
    private fabAddSectionBtn: HTMLElement | null;
    private sectionTemplate: HTMLElement | null;
    private modal: Modal;

    // Reference to the external TOC component - set via method/constructor
    private tocComponent: TableOfContents | null = null;

    // Store section IDs and metadata fetched from the API before initializing shells
    private initialSectionIds: string[] = []; // Keep this if GetDesign returns only IDs
    private initialSectionsMetadata: V1Section[] | null = null; // Keep this

    // Predefined title suggestions (keep as before)
    private static readonly TEXT_TITLES: string[] = [
      "Requirements", "Functional Requirements", "Non-Functional Requirements",
      "API Design", "Data Model", "High-Level Design", "Detailed Design",
      "Assumptions", "Security Considerations", "Deployment Strategy",
      "Monitoring & Alerting", "Future Considerations", "Capacity Planning",
      "System Interfaces", "User Scenarios"
    ];
    private static readonly DRAWING_TITLES: string[] = [
        "Architecture Overview", "System Components", "Data Flow Diagram",
        "Sequence Diagram", "Network Topology", "Component Interactions",
        "Deployment View", "API Interactions", "Database Schema",
        "High-Level Architecture", "User Flow"
    ];
    private static readonly PLOT_TITLES: string[] = [
        "Scalability Analysis", "Latency vs Throughput", "QPS Estimates",
        "Storage Projections", "Cost Analysis", "Performance Metrics",
        "Resource Utilization", "Traffic Estimation", "Data Growth",
        "Benchmark Results"
    ];


    private static getRandomTitle(type: SectionType): string {
      const titles = type === 'drawing' ? this.DRAWING_TITLES : type === 'plot' ? this.PLOT_TITLES : this.TEXT_TITLES;
      const randomIndex = Math.floor(Math.random() * titles.length);
      return titles[randomIndex] || `New ${type.charAt(0).toUpperCase() + type.slice(1)} Section`; // Fallback
    }


    constructor(public currentDesignId: string | null) {
        this.sectionsContainer = document.getElementById('sections-container');
        this.emptyStateEl = document.getElementById('empty-state');
        this.fabAddSectionBtn = document.getElementById('fab-add-section-btn');
        this.sectionTemplate = document.getElementById('section-template');
        this.modal = Modal.getInstance();

        if (!this.sectionsContainer || !this.emptyStateEl || !this.fabAddSectionBtn || !this.sectionTemplate) {
            console.error("SectionManager: Could not find all required DOM elements. Check IDs.");
        }
        if (this.sectionTemplate) {
           this.sectionTemplate.classList.add('hidden');
        }

        this.bindEvents();
        // Initial handleEmptyState check might show empty briefly before load starts
        this.handleEmptyState();
    }

    /** Sets the TableOfContents component instance for communication */
    public setTocComponent(toc: TableOfContents): void {
        this.tocComponent = toc;
        // Update TOC immediately when it's set, in case sections loaded before TOC was ready
        this.triggerTocUpdate();
    }

    /**
     * NEW: Sets the design ID for API calls.
     */
    public setDesignId(designId: string): void {
        this.currentDesignId = designId;
        console.log(`SectionManager: Design ID set to ${designId}`);
    }

    /**
     * Stores the initial section IDs and metadata received from the API (e.g., from GetDesign).
     * This data will be used by initializeSections to create the section shells.
     */
    public setInitialSectionInfo(ids: string[], metadata?: V1Section[] | null): void {
        console.log("SectionManager: Received initial section info", { count: ids.length, hasMetadata: !!metadata });
        this.initialSectionIds = ids || [];
        this.initialSectionsMetadata = metadata || null;
        // Clear any existing sections before initializing new ones
        this.clearAllSections();
    }

    /** Bind event listeners (only for section creation/type selection now) */
    private bindEvents(): void {
        // Bind add section buttons originating from OUTSIDE the TOC
        // Note: The TOC's own 'Add Section' button is handled by TableOfContents component
        document.addEventListener('click', (e: MouseEvent) => {
            const target = e.target as HTMLElement;
            // Only listen for buttons NOT inside the TOC sidebar
            const addSectionBtn = target.closest('#create-first-section'); // Only listen for the empty state button here

            // Add buttons within sections are now handled by BaseSection -> onAddSectionRequest callback
            if (addSectionBtn) {
                e.preventDefault();
                let insertAfterId: string | null = null;
                // The #create-first-section button always adds to the end
                this.openSectionTypeSelector(null, 'after');
            }
        });


         // Bind the FAB "Add Section" button
         if (this.fabAddSectionBtn) { // Use the class property directly
            this.fabAddSectionBtn.addEventListener('click', () => {
                // Find the ID of the *current* last section
                const lastSection = this.sectionData.length > 0
                    ? this.sectionData.sort((a, b) => a.order - b.order)[this.sectionData.length - 1]
                    : null;
                const lastSectionId = lastSection ? lastSection.id : null;
                // Open selector to add AFTER the last section (or to the end if list is empty, though button should be hidden then)
                this.openSectionTypeSelector(lastSectionId, 'after');
            });
         }
    }

    /** Open section type selector modal */
    public openSectionTypeSelector(relativeToId: string | null, position: 'before' | 'after' = 'after'): void {
        this.modal.show('section-type-selector', {
            relativeToId: relativeToId || '',
            position: position
        });
        console.log('SectionManager opening section type selector modal', { relativeToId, position });
    }

    /**
     * Handles the user's choice from the section type selector modal.
     * Calls the API first, then calls createSectionInternal with the API response,
     * and finally tells the new section instance to render its initial content.
     */
    public async handleSectionTypeSelection(
        type: SectionType,
        relativeToId: string | null = null,
        position: 'before' | 'after' = 'after'
    ): Promise<void> {
         if (!this.currentDesignId) {
            console.error("Cannot add section: Design ID not set in SectionManager.");
            // toastManager.showToast("Error", "Cannot add section: Design ID missing.", "error");
            return;
        }
        console.log(`User requested to add a '${type}' section ${position} ${relativeToId || 'the end'}`);

        let apiPosition: V1PositionType = V1PositionType.PositionTypeEnd;
        if (relativeToId) {
            apiPosition = position === 'before' ? V1PositionType.PositionTypeBefore : V1PositionType.PositionTypeAfter;
        }

        const apiSectionPayload = {
             type: mapFrontendSectionTypeToApi(type),
             title: SectionManager.getRandomTitle(type),
             ...mapFrontendContentToApiUpdate(type, this.getDefaultContent(type)),
        };

        try {
            console.log("Calling designServiceAddSection API...");
            const createdApiSection = await DesignApi.designServiceAddSection({
                sectionDesignId: this.currentDesignId,
                body: {
                    section: apiSectionPayload,
                    relativeSectionId: relativeToId || undefined,
                    position: apiPosition,
                }
            });
            console.log("API Success: designServiceAddSection returned:", createdApiSection);

            // Convert API response to frontend SectionData
            const newSectionData = convertApiSectionToSectionData(createdApiSection);
            newSectionData.designId = this.currentDesignId; // Ensure designId is set

            // Create the section instance and DOM shell (constructor shows loading initially)
            const newSectionInstance = this.createSectionInternal(newSectionData, true); // Pass true for user action

            if (newSectionInstance) {
                // --- Use the new method to render the initial content ---
                // This bypasses the loadContent fetch for the newly created section,
                // using the content we just received from the API.
                console.log(`Calling setInitialContentAndRender for new section ${newSectionInstance.sectionId}`);
                newSectionInstance.setInitialContentAndRender(newSectionData.content);
                // --- End new method call ---

                // Scroll the new section into view smoothly
                const sectionElement = document.getElementById(newSectionInstance.sectionId);
                if (sectionElement) {
                    sectionElement.scrollIntoView({ behavior: 'smooth', block: 'center' });
                }

            } else {
                 console.error("Failed to create section instance locally after API success.");
                 // toastManager.showToast(...)
            }

        } catch (error: any) { // Catch specific error type if known
            console.error("API Error adding section:", error);
            // Consider checking error status code (e.g., 4xx vs 5xx) for different messages
            const errorMsg = error.message || (error.response ? await error.response.text() : 'Server Error');
            // toastManager.showToast("Error Adding Section", `Failed to add section: ${errorMsg}`, "error");
        }
    }

    /**
     * Internal method to create the section instance, DOM elements, and update state.
     * Called during initial load (with metadata only) or after a successful addSection API call (with full data).
     * @param sectionData - Data for the section. Crucially, `content` might be empty/null during initial load.
     * @param isUserAction - Flag for reordering/TOC updates.
     * @returns The created BaseSection instance or null on failure.
     */
    private createSectionInternal(
        sectionData: SectionData,
        isUserAction: boolean = false
    ): BaseSection | null {
        console.log("createSectionInternal called with data:", sectionData, "isUserAction:", isUserAction);

        const sectionId = sectionData.id;
        const type = sectionData.type;
        const order = sectionData.order;

        if (!sectionId || !type) {
            console.error("Cannot create section internal: Missing ID or Type in data", sectionData);
            return null;
        }

        // --- DOM Element Creation ---
        if (!this.sectionTemplate || !this.sectionsContainer) { 
          console.error("sectionTemplate and sectionContainer are both missing")
          return null
        }
        const sectionEl = this.sectionTemplate.cloneNode(true) as HTMLElement;
        sectionEl.classList.remove('hidden');
        sectionEl.id = sectionId;
        sectionEl.dataset.sectionId = sectionId;
        sectionEl.dataset.sectionType = type;

        // --- DOM Insertion ---
        // Append to the end for now, reorderSectionsInDOM will fix the visual order later.
        this.sectionsContainer.appendChild(sectionEl);

        // --- Section Instance Creation ---
        let sectionInstance: BaseSection;
        const callbacks: SectionCallbacks = {
            onDelete: this.deleteSection.bind(this),
            onMoveUp: this.moveSectionUp.bind(this),
            onMoveDown: this.moveSectionDown.bind(this),
            onTitleChange: this.updateSectionTitle.bind(this),
            // onContentChange: REMOVED - Content saving handled by BaseSection
            onAddSectionRequest: this.openSectionTypeSelector.bind(this)
        };

        // Create instance with the provided data (content might be missing initially)
        // The BaseSection constructor now handles the initial loading state display.
        switch (type) {
            case 'text':    sectionInstance = new TextSection(sectionData, sectionEl, callbacks); break;
            case 'drawing': sectionInstance = new DrawingSection(sectionData, sectionEl, callbacks); break;
            case 'plot':    sectionInstance = new PlotSection(sectionData, sectionEl, callbacks); break;
            default:
                console.warn(`Unhandled section type "${type}" during creation. Defaulting to Text.`);
                sectionData.type = 'text'; // Correct the data type if defaulting
                sectionInstance = new TextSection(sectionData, sectionEl, callbacks);
        }

        // Add instance to the element for easy access (e.g., retry button)
        (sectionEl as any).componentInstance = sectionInstance;

        // --- Update State and UI ---
        this.sections.set(sectionId, sectionInstance);
        // Store the metadata part in sectionData
        this.sectionData.push({
             id: sectionData.id,
             designId: sectionData.designId,
             type: sectionData.type,
             title: sectionData.title,
             order: sectionData.order,
             content: null // SectionManager doesn't store content here anymore
         });

        // Renumber and update TOC only if it's a user action causing immediate UI change.
        // For bulk initialization, this will be called once at the end.
        if (isUserAction) {
             this.normalizeOrdersAndRenumber();
             this.reorderSectionsInDOM();
             this.triggerTocUpdate();
             this.handleEmptyState();
        }

        return sectionInstance;
    }

    /** Returns appropriate default content for a section type */
    private getDefaultContent(type: SectionType): SectionData['content'] {
        switch (type) {
            case 'text': return '';
            case 'drawing': return { format: 'placeholder_drawing', data: {} };
            case 'plot': return { format: 'placeholder_plot', data: {} };
            default: return '';
        }
    }

    /**
     * REMOVED: Fetches content for multiple section IDs and then loads them.
     * This is now handled by BaseSection.loadContent().
     */
    // public async loadSectionContentsByIds(designId: string, sectionIds: string[]): Promise<void> { /* ... removed ... */ }


    /**
     * Initializes section shells based on provided metadata (typically from API).
     * Creates instances and renders basic structure, but does *not* load content.
     * Content loading is triggered separately by calling loadContent() on the returned instances.
     * @param sectionsMetadata - Array of section metadata objects.
     * @returns An array of the created BaseSection instances.
     */
    public initializeSections(sectionsMetadata: SectionData[]): BaseSection[] {
        if (!this.sectionsContainer) return [];
        console.log("SectionManager: Initializing sections from metadata:", sectionsMetadata.length);
        this.clearAllSections(); // Clear DOM, data, map

        const createdInstances: BaseSection[] = [];

        // Ensure metadata is sorted by order before creating instances
        sectionsMetadata.sort((a, b) => a.order - b.order);

        sectionsMetadata.forEach(meta => {
            // Ensure content is null/empty for initial shell creation
            const initialData: SectionData = {
                ...meta,
                content: null // Explicitly null/empty content
            };

            // Create the section shell
            const sectionInstance = this.createSectionInternal(initialData, false); // isUserAction = false
            if (sectionInstance) {
                createdInstances.push(sectionInstance);
            } else {
                 console.error(`Failed to create section instance for ID: ${meta.id}`);
            }
        });

        this.normalizeOrdersAndRenumber(); // Final renumbering after creating all shells
        this.triggerTocUpdate();           // Update TOC once after initializing all
        this.handleEmptyState();           // Update empty state (might be empty initially until content loads)

        console.log(`SectionManager: ${createdInstances.length} section shells initialized.`);
        return createdInstances;
    }

    /** Clears all sections and resets state */
     private clearAllSections(): void {
         // if (this.sectionsContainer) this.sectionsContainer.innerHTML = '';
         if (this.sectionsContainer) {
             // Remove only the actual section elements, not the empty state container
             const sectionElements = this.sectionsContainer.querySelectorAll('[data-section-id]');
             sectionElements.forEach(el => el.remove());
             console.log("SectionManager: Cleared section elements.");
         }

         this.sections.clear();
         this.sectionData = [];
         this.nextSectionId = 1;
         this.triggerTocUpdate(); // Update TOC to reflect emptiness
         this.handleEmptyState();
     }

    /** Normalize order values and update numbering */
     private normalizeOrdersAndRenumber(): void {
         this.sectionData.sort((a, b) => a.order - b.order);
         this.sectionData.forEach((sectionDataItem, index) => {
             sectionDataItem.order = index + 1;
             this.sections.get(sectionDataItem.id)?.updateNumber(index + 1);
         });
     }


    /** Delete a section */
    private deleteSection(sectionId: string): void {
        if (!this.currentDesignId) {
            console.error("Cannot delete section: Design ID not set.");
            // toastManager.showToast("Error", "Cannot delete section: Design ID missing.", "error");
            return;
        }
        if (!confirm('Are you sure you want to delete this section?')) return;

        console.log(`Attempting to delete section ${sectionId} from design ${this.currentDesignId}`);
        const sectionInstance = this.sections.get(sectionId);
        if (!sectionInstance) {
            console.warn(`Section instance ${sectionId} not found during delete.`);
            return; // Should not happen if UI is consistent
        }

        DesignApi.designServiceDeleteSection({
            designId: this.currentDesignId,
            sectionId: sectionId
        }).then(() => {
            console.log(`API: Successfully deleted section ${sectionId}`);
            // --- Local Removal on API Success ---
            sectionInstance.removeElement(); // Remove from DOM
            this.sectionData = this.sectionData.filter(s => s.id !== sectionId); // Remove metadata
            this.sections.delete(sectionId); // Remove instance reference
            this.normalizeOrdersAndRenumber();
            this.reorderSectionsInDOM();
            this.triggerTocUpdate();
            this.handleEmptyState();
            // toastManager.showToast("Section Deleted", "Section removed successfully.", "success", 2000);
        }).catch(error => {
            console.error(`API Error deleting section ${sectionId}:`, error);
            // toastManager.showToast("Delete Error", `Failed to delete section: ${error.message || 'Server Error'}`, "error");
            // Do NOT remove locally if API failed
        }).finally(() => {
            // Hide loading indicator?
        });
    }

    /** Common logic for moving sections via API */
    private async moveSectionApi(sectionId: string, moveUp: boolean): Promise<void> {
        // This relies on this.sectionData having correct order info, which it should
        if (!this.currentDesignId) {
             console.error("Cannot move section: Design ID not set.");
            // toastManager.showToast("Error", "Cannot move section: Design ID missing.", "error");
             return;
         }
        const currentIndex = this.sectionData.findIndex(s => s.id === sectionId);
        if (moveUp && currentIndex <= 0) return; // Cannot move first item up
        if (!moveUp && currentIndex >= this.sectionData.length - 1) return; // Cannot move last item down

        const targetIndex = moveUp ? currentIndex - 1 : currentIndex + 1;
        if (targetIndex < 0 || targetIndex >= this.sectionData.length) {
            console.warn("Calculated invalid target index for move:", { sectionId, moveUp, currentIndex, targetIndex });
            return; // Should be caught by above checks, but safeguard
        }
         const relativeSectionId = this.sectionData[targetIndex].id;
         const position = moveUp ? V1PositionType.PositionTypeBefore : V1PositionType.PositionTypeAfter;

         console.log(`Attempting to move section ${sectionId} ${position} section ${relativeSectionId} in design ${this.currentDesignId}`);

        try {
            await DesignApi.designServiceMoveSection({
                designId: this.currentDesignId,
                sectionId: sectionId, // Section being moved
                body: {
                    relativeSectionId: relativeSectionId,
                    position: position,
                }
            });
            console.log(`API: Successfully moved section ${sectionId}`);
            // --- Local Reorder on API Success ---
            const currentOrder = this.sectionData[currentIndex].order;
            const targetOrder = this.sectionData[targetIndex].order;
            this.sectionData[currentIndex].order = moveUp ? targetOrder - 0.5 : targetOrder + 0.5; // Temp order for sort
            this.reorderAndRenumber(); // Normalizes orders, updates DOM numbers, reorders DOM, triggers TOC
            // toastManager.showToast("Section Moved", "Section reordered successfully.", "success", 1500);
        } catch (error) {
            console.error(`API Error moving section ${sectionId}:`, error);
            // toastManager.showToast("Move Error", `Failed to move section: ${error.message || 'Server Error'}`, "error");
        } finally {
          // Hide loading indicator?
        }
    }

    /** Move a section up */
    private moveSectionUp(sectionId: string): void { this.moveSectionApi(sectionId, true); }

    /** Move a section down */
    private moveSectionDown(sectionId: string): void { this.moveSectionApi(sectionId, false); }

    /** Reorders DOM, renumbers, and updates TOC */
    private reorderAndRenumber(): void {
         this.normalizeOrdersAndRenumber();
         this.reorderSectionsInDOM();
         this.triggerTocUpdate(); // Update TOC after reordering
    }


    /** Reorder section elements in the DOM based on the sorted sectionData */
    private reorderSectionsInDOM(): void {
        if (!this.sectionsContainer) return;
        this.sectionData.forEach(data => {
            const sectionEl = document.getElementById(data.id);
            if (sectionEl) this.sectionsContainer!.appendChild(sectionEl);
        });
    }


    /** Update section title in data and trigger TOC update */
    private updateSectionTitle(sectionId: string, newTitle: string): void {
        if (!this.currentDesignId) {
            console.error("Cannot update section title: Design ID not set.");
            // Optionally show toast error - but this shouldn't happen if page loaded correctly
            return;
        }
        const sectionData = this.sectionData.find(s => s.id === sectionId);
        if (sectionData) {
             // Check if title actually changed before calling API
             if (sectionData.title === newTitle) {
                console.log(`Section ${sectionId}: Title unchanged, skipping API call.`);
                this.triggerTocUpdate(); // Still update TOC in case of display issues
                return;
            }
            sectionData.title = newTitle;
            this.triggerTocUpdate(); // Update TOC when title changes

            console.log(`Calling API to update title for section ${sectionId}`);
            // API call ONLY for title, mask is just "title"
            DesignApi.designServiceUpdateSection({
                sectionDesignId: this.currentDesignId,
                sectionId: sectionId,
                body: {
                     // The section object inside the body should ONLY contain the fields
                     // specified in the updateMask. The API client model might require
                     // the full structure, but only the 'title' field matters here.
                     // Let's rely on the backend correctly interpreting the mask.
                     // We can send a minimal section object.
                    section: { title: newTitle }, // Minimal payload
                    updateMask: "title", // Specify only title
                },
            }).then(updatedSection => {
                console.log(`API: Successfully updated title for section ${sectionId}`, updatedSection);
                // Optional: Update local sectionData.title again from response if needed
                // sectionData.title = updatedSection.title || newTitle;
                // toastManager.showToast("Title Saved", `Section "${newTitle}" title saved.`, "success", 1500);
            }).catch(error => {
                console.error(`API Error updating title for section ${sectionId}:`, error);
                // toastManager.showToast("Save Error", `Failed to save title for section "${newTitle}".`, "error");
                // TODO: Consider reverting UI or marking section as unsaved? For now, just log/toast.
            });
        }
    }

    /**
     * REMOVED: Update section content in data.
     * This is now handled by BaseSection.saveContent().
     */
    // private updateSectionContent(sectionId: string, newContent: SectionData['content']): void { /* ... removed ... */ }


    /** Creates the simplified data structure needed by the TOC component */
    private getTocItemsInfo(): TocItemInfo[] {
        return this.sectionData
            .sort((a, b) => a.order - b.order) // Ensure sorted before mapping
            .map(section => ({
                id: section.id,
                title: section.title,
                order: section.order
            }));
    }


    /** Calls the update method on the TOC component, if available */
    private triggerTocUpdate(): void {
        if (this.tocComponent) {
            this.tocComponent.update(this.getTocItemsInfo());
            console.log("TOC update triggered.");
        } else {
            console.warn("Cannot trigger TOC update: TOC component not set.");
        }
    }


    /** Handle empty state visibility */
    public handleEmptyState(): void {
        if (!this.emptyStateEl || !this.fabAddSectionBtn) return;
        if (this.sectionData.length === 0) {
            this.emptyStateEl.classList.remove('hidden');
            this.fabAddSectionBtn.classList.add('hidden'); // Hide final button when empty
        } else {
            this.emptyStateEl.classList.add('hidden');
            this.fabAddSectionBtn.classList.remove('hidden'); // Show final button when not empty
        }
    }


    /** Get all section data formatted for the document model */
    public getDocumentSections(): DocumentSection[] {
        return this.sectionData
            .sort((a, b) => a.order - b.order)
            .map(sectionDataItem => {
                const sectionInstance = this.sections.get(sectionDataItem.id);
                if (sectionInstance) {
                    // Get the data *including potentially unsaved content* from the instance
                    return sectionInstance.getDocumentData();
                }
                // Fallback: Return metadata only if instance is missing
                console.warn(`Section instance not found for ID: ${sectionDataItem.id} during getDocumentSections.`);
                // Return a structure matching DocumentSection but potentially without content
                return {
                    id: sectionDataItem.id,
                    title: sectionDataItem.title,
                    order: sectionDataItem.order,
                    type: sectionDataItem.type,
                    content: null // Indicate content might be missing
                } as any; // Cast needed as content is missing
            });
    }


    /** Notifies all managed sections that the application theme has changed. */
    public notifySectionsOfThemeChange(): void {
        console.log("SectionManager: Notifying all sections of theme change...");
        this.sections.forEach(section => {
            try {
                section.handleThemeChange();
            } catch (error) {
                 console.error(`Error handling theme change for section ${section.getId()}:`, error);
            }
        });
    }
}

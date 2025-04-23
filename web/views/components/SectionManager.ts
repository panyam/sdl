// components/SectionManager.ts

import { Modal } from './Modal';
import { BaseSection } from './BaseSection';
import { TextSection } from './TextSection';
import { DrawingSection } from './DrawingSection';
import { PlotSection } from './PlotSection';
import { DocumentSection, SectionType, TextContent, DrawingContent, PlotContent, SectionData, SectionCallbacks } from './types';
import { TocItemInfo, TableOfContents } from './TableOfContents';
import { V1Section, V1PositionType, /* other models if needed */ } from './apiclient'; // Import V1Section and V1PositionType
import { DesignApi } from './Api'; // Import API client
import { convertApiSectionToSectionData } from './converters'; // Import the converter
import { createApiSectionUpdateObject, mapFrontendContentToApiUpdate, mapFrontendSectionTypeToApi } from './converters'; // Import the update object creator


/**
 * Manages document sections using the BaseSection hierarchy.
 * Collaborates with TableOfContents component for UI updates.
 */
export class SectionManager {
    private designId: string = "";
    private sections: Map<string, BaseSection> = new Map();
    private sectionData: SectionData[] = [];
    private nextSectionId: number = 1;
    private sectionsContainer: HTMLElement | null;
    private emptyStateEl: HTMLElement | null;
    private fabAddSectionBtn: HTMLElement | null;
    private sectionTemplate: HTMLElement | null;
    private modal: Modal;

    // Reference to the external TOC component - set via method/constructor
    private tocComponent: TableOfContents | null = null;

    // Store section IDs and metadata fetched from the API before loading content
    private initialSectionIds: string[] = [];
    private initialSectionsMetadata: V1Section[] | null = null; // Optional metadata

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
        this.fabAddSectionBtn = document.getElementById('fab-add-section-btn'); // Find the new area
        this.sectionTemplate = document.getElementById('section-template');
        this.modal = Modal.getInstance();

        if (!this.sectionsContainer || !this.emptyStateEl || !this.fabAddSectionBtn || !this.sectionTemplate) {
            console.error("SectionManager: Could not find all required DOM elements. Check IDs.");
        }
        if (this.sectionTemplate) {
           this.sectionTemplate.classList.add('hidden');
        }

        this.bindEvents();
        this.handleEmptyState(); // Initial check
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
     * NEW: Stores the initial section IDs and metadata received from the API.
     * This data will be used later to load the actual section content.
     */
    public setInitialSectionInfo(ids: string[], metadata?: V1Section[] | null): void {
        console.log("SectionManager: Received initial section info", { count: ids.length, hasMetadata: !!metadata });
        this.initialSectionIds = ids || [];
        this.initialSectionsMetadata = metadata || null;
        // Do NOT load sections here yet. That's Step 1.2.
        // Do NOT call handleEmptyState here yet, as sectionData is still empty.
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
     * This is the entry point for USER-TRIGGERED section creation.
     * It calls the API first, then calls createSectionInternal with the API response.
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

        // Map frontend position to API position type
        let apiPosition: V1PositionType;
        if (!relativeToId) {
            apiPosition = V1PositionType.PositionTypeEnd;
        } else {
            apiPosition = position === 'before' ? V1PositionType.PositionTypeBefore : V1PositionType.PositionTypeAfter;
        }

        // Prepare minimal section data for the API request
        const apiSectionPayload = {
             // No ID - server generates it
             type: mapFrontendSectionTypeToApi(type), // Use existing converter
             title: SectionManager.getRandomTitle(type),
             // Use default content based on type for the API request
             // The converters should handle mapping this to textContent/drawingContent/plotContent
             ...mapFrontendContentToApiUpdate(type, this.getDefaultContent(type)), // Spread the oneof field
        };

        try {
            console.log("Calling designServiceAddSection API...");
            const createdApiSection = await DesignApi.designServiceAddSection({
                sectionDesignId: this.currentDesignId,
                body: {
                    section: apiSectionPayload,
                    relativeSectionId: relativeToId || undefined, // API expects undefined if null
                    position: apiPosition,
                }
            });
            console.log("API Success: designServiceAddSection returned:", createdApiSection);

            // Convert API response to frontend SectionData
            const newSectionData = convertApiSectionToSectionData(createdApiSection);

            // Now create the section locally using the data from the API response
            this.createSectionInternal(newSectionData, true); // Pass true to indicate it's a user action

        } catch (error) {
            console.error("API Error adding section:", error);
            // toastManager.showToast("Error Adding Section", `Failed to add section: ${error.message || 'Server Error'}`, "error");
        }
    }

    /**
     * Internal method to create the section instance, DOM elements, and update state.
     * Called either during initial load (with data from getSection) or after a successful addSection API call.
     * @param sectionData - The complete data for the section (potentially from API).
     * @param isUserAction - Flag to differentiate between load and user add for reordering/TOC updates.
     */
    private createSectionInternal(
        sectionData: SectionData,
        isUserAction: boolean = false // Default to false (load)
    ): BaseSection | null {
        console.log("createSectionInternal called with data:", sectionData, "isUserAction:", isUserAction);

        // Data is now passed in fully formed
        const sectionId = sectionData.id;
        const type = sectionData.type;
        const order = sectionData.order; // Order comes from API response or load

        // --- DOM Element Creation ---
        if (!this.sectionTemplate || !this.sectionsContainer) {
            console.error("Missing section template or container in createSectionInternal.");
            return null;
        }
        const sectionEl = this.sectionTemplate.cloneNode(true) as HTMLElement;
        sectionEl.classList.remove('hidden');
        sectionEl.id = sectionId;
        sectionEl.dataset.sectionId = sectionId;
        sectionEl.dataset.sectionType = type;

        // --- DOM Insertion ---
        // For both load and user-action adds after API call, we rely on reordering later.
        // Append to the end for now, normalizeOrdersAndRenumber will fix the visual order.
        this.sectionsContainer.appendChild(sectionEl);

        /* // --- Old DOM insertion logic based on relative position (No longer needed here) ---
        let insertBeforeEl: HTMLElement | null = null;
        if (isUserAction && relativeToId) { // Only calculate DOM position for user adds
             const relativeEl = this.sectionsContainer.querySelector(`[data-section-id="${relativeToId}"]`) as HTMLElement;
             if (relativeEl) {
                 insertBeforeEl = (position === 'before') ? relativeEl : relativeEl.nextElementSibling as HTMLElement | null;
             }
        } else {
            this.sectionsContainer.appendChild(sectionEl);
        }
        if (insertBeforeEl) { // Insert before a specific element (handles 'before' and 'after')
            this.sectionsContainer.insertBefore(sectionEl, insertBeforeEl);
        } else { // Append to the end (handles adding to empty list, loading, or fallback)
             this.sectionsContainer.appendChild(sectionEl); // Append to end if loading or no insert point found
        }
        */

        // --- Section Instance Creation ---
        let sectionInstance: BaseSection;
        const callbacks: SectionCallbacks = {
            onDelete: this.deleteSection.bind(this),
            onMoveUp: this.moveSectionUp.bind(this),
            onMoveDown: this.moveSectionDown.bind(this),
            onTitleChange: this.updateSectionTitle.bind(this),
            onContentChange: this.updateSectionContent.bind(this),
            onAddSectionRequest: this.openSectionTypeSelector.bind(this) // New callback handler
        };

        switch (type) {
            case 'text':
                sectionInstance = new TextSection(sectionData, sectionEl, callbacks);
                break;
            case 'drawing':
                sectionInstance = new DrawingSection(sectionData, sectionEl, callbacks);
                break;
            case 'plot':
                sectionInstance = new PlotSection(sectionData, sectionEl, callbacks);
                break;
            default:
                console.warn(`Unhandled section type "${type}" during creation. Defaulting to Text.`);
                sectionData.type = 'text';
                sectionInstance = new TextSection(sectionData, sectionEl, callbacks);
        }


        // --- Update State and UI ---
        this.sections.set(sectionId, sectionInstance);
        this.sectionData.push(sectionData); // Add the full data object

        // Renumber and update TOC. If it's a user action, this ensures the new
        // section appears in the correct place visually. If it's part of a bulk load,
        // it will be called once at the end by loadSectionsFromData.
        if (isUserAction) {
             this.normalizeOrdersAndRenumber(); // Sort data, set final order, update DOM numbers
             this.reorderSectionsInDOM();       // Ensure DOM order matches data order
             this.triggerTocUpdate();           // Tell TOC to re-render based on new data order
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
     * NEW: Fetches content for multiple section IDs and then loads them.
     */
    public async loadSectionContentsByIds(designId: string, sectionIds: string[]): Promise<void> {
        if (sectionIds.length === 0) {
            console.log("SectionManager: No section IDs provided to load.");
            this.handleEmptyState(); // Ensure empty state is shown if called with empty list
            return;
        }
        console.log(`SectionManager: Fetching content for ${sectionIds.length} sections...`);
        this.designId = designId;

        const sectionPromises = sectionIds.map(id =>
            DesignApi.designServiceGetSection({ designId: designId, sectionId: id })
                .then(response => {
                    // Assuming response is directly V1Section based on generated client
                    // Adjust if it's wrapped (e.g., response.section)
                    const apiSection: V1Section = response; // Adjust based on actual API response structure
                    return convertApiSectionToSectionData(apiSection);
                })
                .catch(error => {
                    console.error(`Failed to fetch or convert section ${id}:`, error);
                    // Return null or a specific marker for failed sections
                    return null;
                })
        );

        const results = await Promise.all(sectionPromises);

        // Filter out null results (failed loads/conversions) and sort by original order
        const loadedSectionsData = results.filter(data => data !== null) as SectionData[];
        // Sort based on the original order from the API/Design metadata
        loadedSectionsData.sort((a, b) => a.order - b.order);

        console.log(`SectionManager: Successfully fetched and converted ${loadedSectionsData.length} sections.`);
        this.loadSectionsFromData(loadedSectionsData); // Use the existing method to render
    }

    /**
     * Loads multiple sections from provided data (typically from API in Step 1.2).
     * Replaces the old loadSections that used sample data.
     */
    public loadSectionsFromData(sectionsData: SectionData[]): void {
        if (!this.sectionsContainer) return;
        console.log("SectionManager: Loading sections from data:", sectionsData.length);
        this.clearAllSections(); // Clear DOM, data, map
        let maxIdNum = 0;

        // Data should already be sorted by order from API or processing step
        sectionsData.forEach(data => {
            // Use createSection, passing the full SectionData object
            const createdSection = this.createSectionInternal(data, false); // isUserAction = false
            if (createdSection) {
                const idNumMatch = data.id.match(/\d+$/);
                if (idNumMatch) {
                    maxIdNum = Math.max(maxIdNum, parseInt(idNumMatch[0], 10));
                }
            }
        });

        this.nextSectionId = maxIdNum + 1;
        this.normalizeOrdersAndRenumber(); // Final renumbering after load
        this.triggerTocUpdate();           // Update TOC once after loading all
        this.handleEmptyState();           // Update empty state based on sectionData length
        console.log("SectionManager: Sections loaded and rendered from data.");
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
        // Show loading indicator?

        // --- API Call FIRST ---
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
            sectionInstance.removeElement(); // Use the public method
            this.sectionData = this.sectionData.filter(s => s.id !== sectionId);
            this.sections.delete(sectionId);
            this.normalizeOrdersAndRenumber();
            this.reorderSectionsInDOM(); // Ensure DOM order reflects data
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
        // If moving up, we position BEFORE the item currently at the target index.
        // If moving down, we position AFTER the item currently at the target index.
        const position = moveUp ? V1PositionType.PositionTypeBefore : V1PositionType.PositionTypeAfter;

        console.log(`Attempting to move section ${sectionId} ${position} section ${relativeSectionId} in design ${this.currentDesignId}`);
        // Show loading indicator?

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
            // Adjust local order slightly to force resorting, then normalize
            const currentOrder = this.sectionData[currentIndex].order;
            const targetOrder = this.sectionData[targetIndex].order;
            // Give it a temporary order between the target and its neighbor
            this.sectionData[currentIndex].order = moveUp ? targetOrder - 0.5 : targetOrder + 0.5;

            this.reorderAndRenumber(); // This normalizes orders, updates DOM numbers, reorders DOM, triggers TOC
            // toastManager.showToast("Section Moved", "Section reordered successfully.", "success", 1500);

        } catch (error) {
            console.error(`API Error moving section ${sectionId}:`, error);
            // toastManager.showToast("Move Error", `Failed to move section: ${error.message || 'Server Error'}`, "error");
            // Do NOT reorder locally if API failed
        } finally {
             // Hide loading indicator?
        }
    }

    /** Move a section up */
    private moveSectionUp(sectionId: string): void {
        this.moveSectionApi(sectionId, true);
    }


    /** Move a section down */
    private moveSectionDown(sectionId: string): void {
        this.moveSectionApi(sectionId, false);
    }


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
            sectionData.title = newTitle;
            this.triggerTocUpdate(); // Update TOC when title changes

            // --- API Call ---
            console.log(`Calling API to update title for section ${sectionId} in design ${this.currentDesignId}`);
            const updatePayload = createApiSectionUpdateObject({ title: newTitle }, sectionData.type);
            updatePayload.updateMask = "title";
            DesignApi.designServiceUpdateSection({
                sectionDesignId: this.currentDesignId,
                sectionId: sectionId,
                body: updatePayload,
            }).then(updatedSection => {
                console.log(`API: Successfully updated title for section ${sectionId}`, updatedSection);
                // Optional: Update local data with server response if needed, though unlikely for title
                // toastManager.showToast("Title Saved", `Section "${newTitle}" title saved.`, "success", 1500);
            }).catch(error => {
                console.error(`API Error updating title for section ${sectionId}:`, error);
                // toastManager.showToast("Save Error", `Failed to save title for section "${newTitle}".`, "error");
                // TODO: Consider reverting UI or marking section as unsaved? For now, just log/toast.
            });
        }
    }


    /** Update section content in data (called by BaseSection on save) */
    private updateSectionContent(sectionId: string, newContent: SectionData['content']): void {
        const sectionData = this.sectionData.find(s => s.id === sectionId);
        if (sectionData && this.currentDesignId) {
            sectionData.content = newContent;
            // No TOC update needed for content change

            // --- API Call ---
            console.log(`Calling API to update content for section ${sectionId} (type: ${sectionData.type}) in design ${this.currentDesignId}`);
            const updatePayload = createApiSectionUpdateObject({ content: newContent }, sectionData.type);
            // Determine the correct mask path based on the section type
            let maskPath: string = "";
            if (sectionData.type === 'text') maskPath = "textContent";
            else if (sectionData.type === 'drawing') maskPath = "drawing_content";
            else if (sectionData.type === 'plot') maskPath = "plot_content";
            updatePayload.updateMask = maskPath;
            DesignApi.designServiceUpdateSection({
                 sectionDesignId: this.currentDesignId,
                 sectionId: sectionId,
                 body: updatePayload,
             }).then(updatedSection => {
                 console.log(`API: Successfully updated content for section ${sectionId}`, updatedSection);
                 // Optional: Update local data with server response if needed (e.g., updated timestamp)
                 // const updatedData = convertApiSectionToSectionData(updatedSection);
                 // sectionData.content = updatedData.content; // Refresh content just in case
                 // Update timestamps if applicable in BaseSection/SectionData
                // toastManager.showToast("Content Saved", `Content for section "${sectionData.title}" saved.`, "success", 1500);
             }).catch(error => {
                 console.error(`API Error updating content for section ${sectionId}:`, error);
                // toastManager.showToast("Save Error", `Failed to save content for section "${sectionData.title}".`, "error");
                 // TODO: Consider marking section as unsaved?
             });
        }
    }


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
         // Ensure data is sorted before retrieving
        return this.sectionData
            .sort((a, b) => a.order - b.order)
            .map(sectionDataItem => {
                const sectionInstance = this.sections.get(sectionDataItem.id);
                if (sectionInstance) {
                    return sectionInstance.getDocumentData();
                }
                // Fallback if instance not found (should be rare)
                 console.warn(`Section instance not found for ID: ${sectionDataItem.id} during save.`);
                 return {
                     id: sectionDataItem.id,
                     title: sectionDataItem.title,
                     order: sectionDataItem.order,
                     type: sectionDataItem.type,
                     content: sectionDataItem.content
                 } as DocumentSection;
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

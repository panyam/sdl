// components/SectionManager.ts

import { Modal } from './Modal';
import { BaseSection } from './BaseSection';
import { TextSection } from './TextSection';
import { DrawingSection } from './DrawingSection';
import { PlotSection } from './PlotSection';
import { DocumentSection, SectionType, TextContent, DrawingContent, PlotContent, SectionData, SectionCallbacks } from './types';
import { TocItemInfo, TableOfContents } from './TableOfContents';
import { V1Section, /* other models if needed */ } from './apiclient'; // Import V1Section and response types if specific needed
import { DesignApi } from './Api'; // Import API client
import { convertApiSectionToSectionData } from './converters'; // Import the converter



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


    constructor() {
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
     * NEW: Stores the initial section IDs and metadata received from the API.
     * This data will be used later to load the actual section content.
     */
    public setInitialSectionInfo(ids: string[], metadata?: V1Section[] | null): void {
        console.log("SectionManager: Received initial section info", { count: ids.length, hasMetadata: !!metadata });
        this.initialSectionIds = ids || [];
        this.initialSectionsMetadata = metadata || null;
        // Do NOT load sections here yet. That's Step 1.2.
        // Do NOT call handleEmptyState here yet, as sectionData is still empty.
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

        // Bind section type selection (logic remains similar)
         document.addEventListener('click', (e: MouseEvent) => {
             const target = e.target as HTMLElement;
             const sectionTypeOption = target.closest('.section-type-option, button.section-type-option');

             if (sectionTypeOption && this.modal.getCurrentTemplate() === 'section-type-selector') {
                 e.preventDefault();
                 let sectionType: SectionType = 'text';
                 const typeText = sectionTypeOption.querySelector('span')?.textContent?.trim().toLowerCase() || '';

                 if (typeText === 'drawing') sectionType = 'drawing';
                 else if (typeText === 'plot') sectionType = 'plot';

                 const modalData = this.modal.getCurrentData();
                 const relativeToId = modalData?.relativeToId || null;
                 const position = modalData?.position || 'after';

                 this.createSection(sectionType, relativeToId, position); // Create section
                 this.modal.hide(); // Close modal
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

    /** Create a new section instance based on the type */
    private createSection(
        type: SectionType,
        relativeToId: string | null = null,
        position: 'before' | 'after' = 'after', // Default to 'after' if not specified
        initialData?: Partial<SectionData> // Changed to Partial<SectionData> to accept incomplete data during load
    ): BaseSection | null {
        // Handle load scenario: initialData is present, relativeToId/position are irrelevant for initial placement logic
        // Ensure initialData has an ID if provided, otherwise generate one
        const sectionId = initialData?.id || `section-${this.nextSectionId++}`;

        if (initialData?.order !== undefined) {
            relativeToId = null; // Ignore relative positioning during load
            position = 'after'; // Doesn't matter, order is explicit
        }

        // --- Section Data Creation (Simplified) ---
        let sectionOrder: number;
        const providedOrder = initialData?.order;

        if (providedOrder !== undefined) {
             sectionOrder = providedOrder; // Use provided order when loading
        } else {
             // Calculate temporary order for insertion sorting
             if (relativeToId) {
                 const relativeIndex = this.sectionData.findIndex(s => s.id === relativeToId);
                 if (relativeIndex !== -1) {
                     const relativeOrder = this.sectionData[relativeIndex].order;
                     sectionOrder = (position === 'before') ? relativeOrder - 0.5 : relativeOrder + 0.5;
                 } else {
                     sectionOrder = this.sectionData.length; // Fallback: append if relative ID not found
                 }
             } else {
                 sectionOrder = this.sectionData.length; // Append to end if no relative ID
             }
        }

        // Use provided data or generate defaults
        const sectionData: SectionData = {
            id: sectionId,
            designId: this.designId,
            type: type, // Type is required
            title: initialData?.title || SectionManager.getRandomTitle(type),
            content: initialData?.content || this.getDefaultContent(type),
            order: sectionOrder, // Will be normalized later
        };

        // --- DOM Element Creation (Simplified) ---
        if (!this.sectionTemplate || !this.sectionsContainer) return null;
        const sectionEl = this.sectionTemplate.cloneNode(true) as HTMLElement;
        sectionEl.classList.remove('hidden');
        sectionEl.id = sectionId;
        sectionEl.dataset.sectionId = sectionId;
        sectionEl.dataset.sectionType = type;

        // Insertion logic based purely on data sort later, unless it's a user action
        // Determine where to insert the new DOM element based on user interaction
        let insertBeforeEl: HTMLElement | null = null;
        if (initialData?.order === undefined && relativeToId) { // Only calculate DOM position for user adds
            const relativeEl = this.sectionsContainer.querySelector(`[data-section-id="${relativeToId}"]`) as HTMLElement
            if (relativeEl) {
                insertBeforeEl = (position === 'before') ? relativeEl : relativeEl.nextElementSibling as HTMLElement | null;
            }
        }

        // Perform the DOM insertion
        if (insertBeforeEl) { // Insert before a specific element (handles 'before' and 'after')
            this.sectionsContainer.insertBefore(sectionEl, insertBeforeEl);
        } else { // Append to the end (handles adding to empty list, loading, or fallback)
             this.sectionsContainer.appendChild(sectionEl); // Append to end if loading or no insert point found
        }

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

        // Only renumber/update TOC if it's a single user action, not during bulk load
        if (initialData?.order === undefined) {
             this.normalizeOrdersAndRenumber(); // Sort data, set final order, update DOM numbers
             this.triggerTocUpdate();           // Tell TOC to re-render
             this.handleEmptyState(); // Update empty state based on sectionData length
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
            const createdSection = this.createSection(data.type, null, 'after', data);
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
         if (this.sectionsContainer) this.sectionsContainer.innerHTML = '';
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
        if (!confirm('Are you sure you want to delete this section?')) return;
        const sectionInstance = this.sections.get(sectionId);
        if (sectionInstance) {
            sectionInstance.removeElement(); // Use the public method
            this.sectionData = this.sectionData.filter(s => s.id !== sectionId);
            this.sections.delete(sectionId);
            this.normalizeOrdersAndRenumber();
            this.triggerTocUpdate();
            this.handleEmptyState();
        }
    }


    /** Move a section up */
    private moveSectionUp(sectionId: string): void {
        const sectionIndex = this.sectionData.findIndex(s => s.id === sectionId);
        if (sectionIndex <= 0) return;
        this.sectionData[sectionIndex].order -= 1.5; // Adjust order for sorting
        this.reorderAndRenumber();
    }


    /** Move a section down */
    private moveSectionDown(sectionId: string): void {
        const sectionIndex = this.sectionData.findIndex(s => s.id === sectionId);
        if (sectionIndex === -1 || sectionIndex >= this.sectionData.length - 1) return;
        this.sectionData[sectionIndex].order += 1.5; // Adjust order for sorting
        this.reorderAndRenumber();
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
        const sectionData = this.sectionData.find(s => s.id === sectionId);
        if (sectionData && sectionData.title !== newTitle) {
            sectionData.title = newTitle;
            this.triggerTocUpdate(); // Update TOC when title changes
        }
    }


    /** Update section content in data (called by BaseSection on save) */
    private updateSectionContent(sectionId: string, newContent: SectionData['content']): void {
        const sectionData = this.sectionData.find(s => s.id === sectionId);
        if (sectionData) {
            sectionData.content = newContent;
            // No TOC update needed for content change
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

    /** Initialize the component */
    public static init(): SectionManager {
        return new SectionManager();
    }
}

// components/TextSection.ts

import { BaseSection } from './BaseSection';
import { SectionData, SectionCallbacks, TextContent } from './types';

// Import TinyMCE core, theme, icons, and model
import tinymce, { Editor } from 'tinymce/tinymce';
import 'tinymce/themes/silver/theme';
import 'tinymce/icons/default/icons';
import 'tinymce/models/dom/model';

export class TextSection extends BaseSection {
    protected static readonly ICON_SVG = `<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-full h-full"><path stroke-linecap="round" stroke-linejoin="round" d="M3.75 6.75h16.5M3.75 12h16.5m-16.5 5.25H12" /></svg>`;
    private editorInstance: Editor | null = null;

    constructor(data: SectionData, element: HTMLElement, callbacks: SectionCallbacks = {}) {
        super(data, element, callbacks);
    }

    /** Checks if the section is currently in edit mode */
    public isInEditMode(): boolean {
        return this.mode === 'edit';
    }

    /** Returns the (type) title for this section. */
    protected getSectionTypeTitle(): string {
      return "Text Section"
    }


    /** Returns the svg content to show for this section. */
    protected getSectionIconSvg(): string {
      return TextSection.ICON_SVG;
    }

    protected populateViewContent(): void {
        const viewContent = this.contentContainer?.querySelector('.section-view-content');
        if (viewContent) {
            const initialContent = (typeof this.data.content === 'string' && this.data.content.length > 0)
                ? this.data.content
                : '<p class="text-gray-400 dark:text-gray-500 italic">Click to add content...</p>';
            viewContent.innerHTML = initialContent;
        } else {
            console.warn(`View content area not found for text section ${this.data.id}`);
        }
    }

    // Ensure populateEditContent cleans up any lingering instance *before* init
    protected override populateEditContent(): void {
        const editorTarget = this.contentContainer?.querySelector('.text-editor-target');
        if (editorTarget instanceof HTMLElement) {
            const initialContent = (typeof this.data.content === 'string') ? this.data.content : '';
            const isDarkMode = document.documentElement.classList.contains('dark');
            const tinyMCEPublicPath = '/static/js/gen/';

            // --- Crucial: Destroy any existing instance before initializing a new one ---
            // This guards against race conditions or multiple calls
            if (this.editorInstance) {
                 console.warn(`Found existing editor instance in populateEditContent for ${this.data.id}. Removing before init.`);
                 try {
                    tinymce.remove(this.editorInstance);
                 } catch (removeError){
                    console.error("Error removing lingering editor instance:", removeError);
                 }
                 this.editorInstance = null;
            }
            // --------------------------------------------------------------------------

            tinymce.init({
                target: editorTarget,
                plugins: 'autolink lists link image charmap preview anchor searchreplace visualblocks code fullscreen insertdatetime media table help wordcount autoresize',
                toolbar: 'undo redo | formatselect | bold italic backcolor | alignleft aligncenter alignright alignjustify | bullist numlist outdent indent | removeformat | help',
                content_style: 'body { font-family:Helvetica,Arial,sans-serif; font-size:14px }',
                height: 300,
                menubar: false,
                statusbar: false,
                autoresize_bottom_margin: 10,
                autoresize_min_height: 200,
                skin_url: `${tinyMCEPublicPath}skins/ui/${isDarkMode ? 'oxide-dark' : 'oxide'}`,
                content_css: `${tinyMCEPublicPath}skins/content/${isDarkMode ? 'dark' : 'default'}/content.min.css`,
                license_key: 'gpl',
                setup: (editor: Editor) => {
                    editor.on('init', () => {
                        editor.setContent(initialContent || '');
                        // *** Assign the instance ONLY on successful init ***
                        this.editorInstance = editor;
                        editor.focus();
                        console.log(`TinyMCE initialized for section ${this.data.id} (Mode: ${isDarkMode ? 'Dark' : 'Light'})`);
                    });
                    editor.on('remove', () => {
                       // Ensure the instance variable is cleared if the editor is removed externally
                       // or during cleanup.
                       if (this.editorInstance === editor) {
                          this.editorInstance = null;
                          console.log(`Cleaned up editor instance variable on 'remove' event for ${this.data.id}`);
                       }
                    });
                    // --- Add Keydown Listener for Save Shortcut ---
                    editor.on('keydown', (e: KeyboardEvent) => {
                        // Check for Cmd/Ctrl + Enter
                        if ((e.metaKey || e.ctrlKey) && e.key === 'Enter') {
                            e.preventDefault(); // Prevent default Enter behavior (like adding a newline)
                            console.log(`Cmd/Ctrl+Enter detected in section ${this.data.id}, triggering save.`);
                            this.handleSaveClick(); // Trigger the save action
                        }
                    });
                }
            }).catch((error: any) => {
                console.error(`Error initializing TinyMCE for section ${this.data.id}:`, error);
                editorTarget.innerHTML = `<p class="text-red-500">Error initializing Rich Text Editor.</p>`;
            });
        } else {
            console.warn(`Edit content target area (.text-editor-target) not found for text section ${this.data.id}`);
        }
    }

    /**
     * Handles theme changes specifically for the TextSection.
     * If the section is currently in edit mode, it reinitializes the
     * TinyMCE editor to apply the correct theme skin and content CSS.
     */
    public override handleThemeChange(): void {
        console.log(`TextSection ${this.data.id}: Handling theme change.`);
        if (this.mode !== 'edit') {
            // Only need to act if the editor is currently active
            return;
        }

        const editorTarget = this.contentContainer?.querySelector('.text-editor-target');
        if (!editorTarget) {
            console.error(`Cannot reinitialize editor for theme change: Target element not found for section ${this.data.id}`);
            return;
        }

        // Get current content before destroying
        let currentContent = '';
        if (this.editorInstance && this.editorInstance.initialized) {
             try {
                 currentContent = this.editorInstance.getContent();
             } catch (error) {
                 console.error("Error getting content before editor reinitialization for theme change:", error);
             }
        } else {
            // Fallback to stored data if editor wasn't ready
            currentContent = typeof this.data.content === 'string' ? this.data.content : '';
        }

        // Destroy the existing editor instance if it exists
        if (this.editorInstance) {
            try {
                console.log(`Destroying existing editor for theme change: ${this.data.id}`);
                tinymce.remove(this.editorInstance);
                // The 'remove' event handler in setup should set this.editorInstance to null
            } catch (error) {
                console.error(`Error destroying TinyMCE instance during theme change for section ${this.data.id}:`, error);
                this.editorInstance = null; // Ensure it's null even on error
            }
        } else {
             console.log(`No active editor instance to destroy for theme change: ${this.data.id}`);
        }

        // Re-run the initialization logic using populateEditContent,
        // which now correctly reads the current document theme state.
        console.log(`Re-populating edit content to reinitialize editor for theme change: ${this.data.id}`);
        this.populateEditContent(); // This will re-initialize TinyMCE with the correct theme paths

        // We might need a slight delay or use the init callback to set content again
        // because populateEditContent runs async init
        // Check if editor becomes available shortly after re-init attempt
        const checkEditorInterval = setInterval(() => {
            if (this.editorInstance) {
                 clearInterval(checkEditorInterval);
                 this.editorInstance.setContent(currentContent);
                 console.log(`Content restored after editor reinitialization for ${this.data.id}`);
            }
            // Add a timeout safeguard if needed
        }, 50); // Check every 50ms

        // Failsafe timeout
         setTimeout(() => {
             if (!this.editorInstance) {
                 clearInterval(checkEditorInterval);
                 console.warn(`Editor instance still not available after timeout during theme change reinitialization for ${this.data.id}`);
             }
         }, 1000); // Stop checking after 1 second


        console.log(`Editor reinitialization triggered for theme change in section ${this.data.id}`);
    }

    protected bindViewModeEvents(): void {
        const viewContent = this.contentContainer?.querySelector('.section-view-content');
        if (viewContent) {
            viewContent.removeEventListener('click', this.handleViewClick);
            viewContent.addEventListener('click', this.handleViewClick.bind(this));
        }
    }

    protected getContentFromEditMode(): TextContent {
        if (this.editorInstance && this.editorInstance.initialized) {
            try {
                return this.editorInstance.getContent() || '';
            } catch (error) {
                console.error(`Error getting content from TinyMCE for section ${this.data.id}:`, error);
                return typeof this.data.content === 'string' ? this.data.content : '';
            }
        }
        console.warn(`TinyMCE instance not found or not initialized when getting content for section ${this.data.id}.`);
        return typeof this.data.content === 'string' ? this.data.content : '';
    }

    // Ensure switchToViewMode also reliably cleans up
    public override switchToViewMode(saveChanges: boolean): void {
        let contentToSave: TextContent | undefined = undefined;
        const wasInEditMode = this.mode === 'edit'; // Remember if we were editing

        if (wasInEditMode) {
            if (this.editorInstance && this.editorInstance.initialized) {
                 if (saveChanges) {
                    try {
                        contentToSave = this.editorInstance.getContent() || '';
                    } catch (error) {
                        console.error(`Error getting content from TinyMCE before destroy for section ${this.data.id}:`, error);
                    }
                 }
                // Always try to remove the editor instance
                try {
                    console.log(`Attempting to remove TinyMCE instance for ${this.data.id} on switch to view`);
                    tinymce.remove(this.editorInstance);
                    // The 'remove' event handler in setup should set this.editorInstance to null
                } catch (error) {
                    console.error(`Error removing TinyMCE instance for section ${this.data.id} on switch to view:`, error);
                    // Ensure it's null even on error, as the editor state is likely invalid
                    this.editorInstance = null;
                }
            } else if (this.editorInstance) {
                 // Instance exists but wasn't initialized, or removal failed previously
                 console.warn(`Editor instance for ${this.data.id} exists but likely invalid during switchToViewMode. Nullifying reference.`);
                 this.editorInstance = null; // Just nullify the reference
            }
        }

        // Process saving *after* attempting editor removal
        if (wasInEditMode && saveChanges && contentToSave !== undefined) {
            console.log(`switchToViewMode: Proceeding with save for section ${this.data.id}`);

            const newContent = contentToSave;
            if (newContent !== this.data.content) {
                this.data.content = newContent;
                this.callbacks.onContentChange?.(this.data.id, this.data.content);
                console.log(`Section ${this.data.id} content saved.`);
            } else {
                console.log(`Section ${this.data.id} content unchanged.`);
            }
        } else if (wasInEditMode && !saveChanges) {
            console.log(`switchToViewMode: Cancelling edit, no save for section ${this.data.id}`);
        }

        // Switch mode and load view template
        this.mode = 'view';
        console.log(`Switching ${this.data.id} to view mode.`);
        if (this.loadTemplate('view')) {
            this.populateViewContent();
            this.bindViewModeEvents();
        } else {
            console.error("Failed to load view template for section", this.data.id);
        }
    }

    public switchToEditMode(): void {
         if (this.mode === 'edit') {
            this.editorInstance?.focus();
            return;
        }

        if (this.editorInstance) {
            console.warn(`Found existing editor instance when switching to edit mode for ${this.data.id}. Attempting cleanup.`);
            try {
                tinymce.remove(this.editorInstance);
            } catch (e) { console.error("Error during cleanup remove:", e); }
            this.editorInstance = null;
        }

        this.mode = 'edit';
        console.log(`Switching ${this.data.id} to edit mode.`);
        if (this.loadTemplate('edit')) {
            this.populateEditContent(); // This will initialize TinyMCE
            this.bindEditModeEvents();
        } else {
             console.error("Failed to load edit template for section", this.data.id);
             this.mode = 'view'; // Revert mode if template fails
        }
    }

 
     /** Implement the abstract method from BaseSection */
     protected resizeContentForFullscreen(isEntering: boolean): void {
         console.log(`TextSection ${this.data.id}: Resizing content trigger. Is entering fullscreen: ${isEntering}`);
         // For TinyMCE with autoresize, explicit resizing is often not needed here.
         // The browser layout handles width changes, and autoresize handles height.
         // If specific resizing logic were needed (e.g., for a different editor), it would go here.
         // Example: this.editorInstance?.execCommand('mceAutoResize'); // Might force a recalc if needed
     }
}

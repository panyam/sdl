// FILE: ./web/views/components/LlmInteractionHandler.ts

import { Modal } from './Modal';
import { ToastManager } from './ToastManager';
import { LlmApi } from './Api';
import { SectionData } from './types';
import {
    LlmServiceSimpleLlmQueryRequest,
    LlmServiceGenerateTextContentRequest,
    LlmServiceReviewTextContentRequest,
    LlmServiceGenerateDefaultPromptsRequest // Import new request type
} from './apiclient'; // Import request types if needed for clarity

export class LlmInteractionHandler {
    private modal: Modal;
    private toastManager: ToastManager;
    private isLoading: boolean = false;

    // Store context while dialog is open
    private currentSectionData: SectionData | null = null;
    private currentApplyCallback: ((generatedText: string) => void) | null = null;
    private currentLlmResponseText: string | null = null;

    constructor(modal: Modal, toastManager: ToastManager) {
        this.modal = modal;
        this.toastManager = toastManager;
    }

    /**
     * Shows the LLM dialog for the given section.
     * @param sectionData Data of the section invoking the dialog.
     * @param applyCallback Optional callback to apply generated text (used by TextSection).
     */
    public showLlmDialog(sectionData: SectionData, applyCallback?: (generatedText: string) => void): void {
        if (!sectionData || !sectionData.type) {
            this.toastManager.showToast("Error", "Cannot open LLM dialog: Section data is missing.", "error");
            return;
        }
        if (this.isLoading) {
            console.warn("LLM Interaction already in progress.");
            return; // Prevent opening multiple dialogs if already loading
        }
        console.log(`Showing LLM dialog for section: ${sectionData.id}`);

        // Store context for the duration the dialog is potentially open
        this.currentSectionData = sectionData;
        this.currentApplyCallback = applyCallback || null;
        this.currentLlmResponseText = null; // Reset previous response

        const modalContentElement = this.modal.show('llm-dialog', {
            // Pass the handler's submit method as the modal's onSubmit callback
            onSubmit: this.handleDialogSubmit.bind(this)
            // Pass other callbacks if needed, like onApply for the results modal later
        });

        // Configure UI after the modal template is loaded into the DOM
        setTimeout(() => {
            if (!modalContentElement) {
                 console.error("LLM Dialog template root element not found after show.");
                 this.modal.hide(); // Hide if template loading failed
                 return;
            };
            this.updateCurrentSectionDisplay(modalContentElement, sectionData.title);
            this.setupTabsAndDefaultGenerationButtons(modalContentElement); // Combine setup
            this.configureDialog(modalContentElement, sectionData); // Populate prompts initially
             // Ensure initial tab state is correct
            // Activate the generate tab by default
            // Ensure initial tab state is correct
            const initialActiveTab = modalContentElement.querySelector<HTMLButtonElement>('.llm-tab-generate');
             initialActiveTab?.click(); // Activate generate tab by default
        }, 50); // Delay to allow modal DOM rendering
    }

    /** Sets up tabs and binds "Generate Default Prompt" buttons */
    private setupTabsAndDefaultGenerationButtons(modalContentElement: HTMLElement): void {
        const tabs = modalContentElement.querySelectorAll<HTMLButtonElement>('.llm-tab');
        const panes = modalContentElement.querySelectorAll<HTMLElement>('.llm-tab-pane');
        const generateDefaultBtn = modalContentElement.querySelector<HTMLButtonElement>('#llm-generate-default-prompt-btn');
        const verifyDefaultBtn = modalContentElement.querySelector<HTMLButtonElement>('#llm-verify-default-prompt-btn');

        tabs.forEach(tab => {
            tab.onclick = () => { // Use onclick for simplicity
                // Deactivate all tabs and panes
                tabs.forEach(t => {
                    t.classList.remove('llm-tab-active', 'border-blue-500', 'text-blue-600', 'dark:text-blue-400', 'dark:border-blue-400');
                    t.classList.add('text-gray-500', 'hover:text-gray-700', 'dark:text-gray-400', 'dark:hover:text-gray-300', 'hover:border-gray-300', 'dark:hover:border-gray-500', 'border-transparent');
                });
                panes.forEach(p => p.classList.add('hidden'));

                // Activate the clicked tab
                tab.classList.add('llm-tab-active', 'border-blue-500', 'text-blue-600', 'dark:text-blue-400', 'dark:border-blue-400');
                tab.classList.remove('text-gray-500', 'hover:text-gray-700', 'dark:text-gray-400', 'dark:hover:text-gray-300', 'hover:border-gray-300', 'dark:hover:border-gray-500', 'border-transparent');

                // Show the corresponding pane
                const targetPaneId = tab.dataset.tabTarget;
                if (targetPaneId) {
                    const targetPane = modalContentElement.querySelector<HTMLElement>(targetPaneId);
                    targetPane?.classList.remove('hidden');
                }
            };
        });

        // Bind "Generate Default" buttons
        if (generateDefaultBtn) {
             generateDefaultBtn.onclick = () => this.handleGenerateDefaultPromptsClick();
        }
        if (verifyDefaultBtn) {
             verifyDefaultBtn.onclick = () => this.handleGenerateDefaultPromptsClick();
        }
    }

    /** Updates the display of the current section title in the LLM dialog */
    private updateCurrentSectionDisplay(modalContentElement: HTMLElement, sectionTitle: string): void {
         const currentSectionElement = modalContentElement.querySelector<HTMLElement>('#llm-current-section');
         if (currentSectionElement) {
             currentSectionElement.textContent = sectionTitle;
         } else {
             console.warn("Could not find #llm-current-section in modal content after showing.");
         }
    }

    /** Configures the LLM dialog based on the section type */
    private configureDialog(modalContentElement: HTMLElement, sectionData: SectionData): void {
        const generatePromptEl = modalContentElement.querySelector<HTMLElement>('#llm-generate-prompt');
        const verifyPromptEl = modalContentElement.querySelector<HTMLElement>('#llm-verify-prompt');
        const verifyTabButton = modalContentElement.querySelector<HTMLButtonElement>('.llm-tab-verify');
        const generateDefaultBtn = modalContentElement.querySelector<HTMLButtonElement>('#llm-generate-default-prompt-btn');
        const verifyDefaultBtn = modalContentElement.querySelector<HTMLButtonElement>('#llm-verify-default-prompt-btn');

        const isTextSection = sectionData.type === 'text';
        verifyTabButton?.classList.toggle('hidden', !isTextSection);

        // --- Populate Get Answer Prompt ---
        if (generatePromptEl) {
             if (sectionData.getAnswerPrompt) {
                 generatePromptEl.textContent = sectionData.getAnswerPrompt;
                 generateDefaultBtn?.classList.add('hidden'); // Hide button if prompt exists
             } else {
                 generatePromptEl.innerHTML = `<span class="italic text-gray-400 dark:text-gray-500">No specific prompt saved. Default will be used.</span>`;
                 generateDefaultBtn?.classList.remove('hidden'); // Show button if no prompt
             }
        }

        // --- Populate Verify Prompt ---
        if (isTextSection) {
            if (verifyPromptEl) {
                if (sectionData.verifyAnswerPrompt) {
                    verifyPromptEl.textContent = sectionData.verifyAnswerPrompt;
                    verifyDefaultBtn?.classList.add('hidden'); // Hide button
                } else {
                    verifyPromptEl.innerHTML = `<span class="italic text-gray-400 dark:text-gray-500">No specific prompt saved. Default will be used.</span>`;
                    verifyDefaultBtn?.classList.remove('hidden'); // Show button
                }
            }
        } else {
            // Non-text sections don't use verify prompts yet
            if (verifyPromptEl) verifyPromptEl.innerHTML = `<span class="italic text-gray-400 dark:text-gray-500">Verification only available for text sections.</span>`;
            verifyDefaultBtn?.classList.add('hidden');
        }
    }

    /** Handles the submission logic when the dialog's submit button is clicked */
    private async handleDialogSubmit(): Promise<void> {
        const sectionData = this.currentSectionData; // Use stored data at the start
        if (this.isLoading || !sectionData) {
            console.warn("LLM submit called while loading or without section data.");
            return;
        }

        const dialogContent = this.modal.getContentElement(); // Get modal content element early
        if (!dialogContent) {
            console.error("LLM Dialog content element not found during submit.");
            return;
        }

        this.isLoading = true;
        console.log(`LLM dialog submit initiated for section ${sectionData.id}`);

        const activeTab = dialogContent.querySelector<HTMLButtonElement>('.llm-tab-active');
        const activePaneId = activeTab?.dataset.tabTarget;
        const submitButton = dialogContent.querySelector<HTMLButtonElement>('button[data-modal-action="submit"]');

        // Show loading state on button
        if (submitButton) {
            submitButton.disabled = true;
            submitButton.innerHTML = `<svg class="animate-spin -ml-1 mr-3 h-5 w-5 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24"> <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle> <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path> </svg> Processing...`;
        }

        let apiCall: Promise<any>;
        let actionType: 'generate' | 'review' | 'custom' = 'custom';
        // *** Define path parameters separately ***
        const pathParams = {
            designId: sectionData.designId,
            sectionId: sectionData.id,
        };

        try {
            // Determine which API call to make based on the active tab
            if (activePaneId === '#llm-generate-content') {
                actionType = 'generate';
                // *** Construct request object for the API client ***
                const request = {
                    ...pathParams, // Spread path params
                    body: {        // Add body object
                       // promptOverride: "...", // Get from UI if needed later
                    }
                };
                console.log("Calling GenerateTextContent API with request:", request);
                apiCall = LlmApi.llmServiceGenerateTextContent(request); // Pass combined object

            } else if (activePaneId === '#llm-verify-content' && sectionData.type === 'text') {
                actionType = 'review';
                 // *** Construct request object for the API client ***
                const request = {
                    ...pathParams, // Spread path params
                    body: {        // Add body object
                        // promptOverride: "...", // Get from UI if needed later
                    }
                };
                console.log("Calling ReviewTextContent API with request:", request);
                apiCall = LlmApi.llmServiceReviewTextContent(request); // Pass combined object
            } else if (activePaneId === '#llm-custom-content') {
                actionType = 'custom';
                const customPromptInput = dialogContent.querySelector<HTMLTextAreaElement>('#custom-prompt');
                const customPrompt = customPromptInput?.value.trim() || '';
                if (!customPrompt) {
                    this.toastManager.showToast("Input Error", "Please enter a custom prompt.", "warning");
                    throw new Error("Custom prompt empty");
                }
                // *** Construct request object for the API client ***
                // SimpleLlmQuery might expect *all* args in body based on its POST /v1/llm/query definition
                const request = {
                    body: { // All fields inside body
                        designId: sectionData.designId,
                        sectionId: sectionData.id,
                        prompt: customPrompt,
                    }
                };
                console.log("Calling SimpleLlmQuery API (Custom) with request:", request);
                apiCall = LlmApi.llmServiceSimpleLlmQuery(request); // Pass object with body
            } else {
                 throw new Error(`Invalid or unsupported LLM action for tab: ${activePaneId} and type: ${sectionData.type}`);
            }

            console.log(`Calling LLM API (${actionType}) for section ${sectionData.id}`);
            const response = await apiCall;
            console.log("LLM API Response:", response);

            this.currentLlmResponseText = response.generatedText || response.reviewText || response.responseText || null; // Store response for potential apply

            // Restore button before hiding modal
            if (submitButton) {
                submitButton.disabled = false;
                submitButton.textContent = 'Submit';
            }
            await this.modal.hide(); // Hide prompt dialog

            // Prepare data for results modal
            let applyCallbackForResults: ((data: any) => void) | undefined = undefined;
            if (actionType === 'generate' && sectionData.type === 'text' && this.currentApplyCallback) {
                // Pass *this* handler's apply method to the results modal
                applyCallbackForResults = this.handleApplyLlmResult.bind(this);
            }

            // Show results modal
            const resultsModalContent = this.modal.show('llm-results', {
                onApply: applyCallbackForResults // Pass the callback
            });

            // Populate results modal content after it's shown
            if (resultsModalContent) {
                setTimeout(() => { // Delay ensures elements are ready
                    const contentArea = resultsModalContent.querySelector<HTMLElement>('#llm-results-content');
                    if (contentArea) contentArea.textContent = this.currentLlmResponseText || "LLM returned no text.";

                    const applyButton = resultsModalContent.querySelector<HTMLButtonElement>('#llm-results-apply');
                    if (applyButton) applyButton.classList.toggle('hidden', !applyCallbackForResults); // Show button only if callback exists
                }, 50);
            }

        } catch (error: any) {
            console.error("Error during LLM interaction:", error);
            if (submitButton) { // Ensure button is reset on error
                submitButton.disabled = false;
                submitButton.textContent = 'Submit';
            }
            const errorMsg = error.message || (error.response ? await error.response.text() : 'Unknown LLM API error');
            this.toastManager.showToast("LLM Error", `Failed: ${errorMsg}`, "error");
        } finally {
            this.isLoading = false;
            // Ensure button is reset if error occurred before assignment
             if (submitButton && submitButton.disabled) {
                 submitButton.disabled = false;
                 submitButton.textContent = 'Submit';
             }
        }
    }

    /** Handles click on "Generate Default Prompt" buttons */
    private async handleGenerateDefaultPromptsClick(): Promise<void> {
      const sectionData = this.currentSectionData;
      if (!sectionData) return;

      const dialogContent = this.modal.getContentElement();
      if (!dialogContent) return;

      // Could show loading state on the button itself
      const genBtn = dialogContent.querySelector<HTMLButtonElement>('#llm-generate-default-prompt-btn');
      const verifyBtn = dialogContent.querySelector<HTMLButtonElement>('#llm-verify-default-prompt-btn');
       if(genBtn) genBtn.disabled = true;
      if(verifyBtn) verifyBtn.disabled = true;

      console.log(`Requesting default prompts for section ${sectionData.id}`);
      this.toastManager.showToast("Generating...", "Generating default prompts...", "info", 2000);

      try {
          const request: LlmServiceGenerateDefaultPromptsRequest = {
              designId: sectionData.designId,
              sectionId: sectionData.id,
              body: {},
          };
          const response = await LlmApi.llmServiceGenerateDefaultPrompts(request);

          // Update the internal sectionData (so dialog redisplay works if needed)
          // This assumes SectionData has these fields populated by GetSection/GetDesign
          sectionData.getAnswerPrompt = response.getAnswerPrompt || "";
          sectionData.verifyAnswerPrompt = response.verifyAnswerPrompt || "";

          // Update the displayed prompts in the current dialog
          this.configureDialog(dialogContent, sectionData); // Re-run config to update display & hide buttons
          this.toastManager.showToast("Success", "Default prompts generated and saved.", "success");
        } catch (error: any) {
          console.error("Error generating default prompts:", error);
          this.toastManager.showToast("Error", `Failed to generate default prompts: ${error.message || 'Server error'}`, "error");
        } finally {
          // Re-enable buttons
          if(genBtn) genBtn.disabled = false;
          if(verifyBtn) verifyBtn.disabled = false;
        }
     }


    /**
     * Called by the results modal's Apply button click (via Modal.ts).
     * This method then calls the original callback provided by BaseSection/TextSection.
     */
    private handleApplyLlmResult(): void {
        if (this.currentApplyCallback && this.currentLlmResponseText !== null) {
            console.log("LlmInteractionHandler: Applying result via stored callback...");
            this.currentApplyCallback(this.currentLlmResponseText); // Call the original callback
        } else {
            console.error("Apply callback or LLM response text is missing.");
            this.toastManager.showToast("Error", "Could not apply result.", "error");
        }
         // Clear stored response after apply attempt
         this.currentLlmResponseText = null;
    }
}

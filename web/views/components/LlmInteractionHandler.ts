// FILE: ./web/views/components/LlmInteractionHandler.ts

import { Modal } from './Modal';
import { ToastManager } from './ToastManager';
import { DesignApi, LlmApi } from './Api'; // Import DesignApi
import { SectionData } from './types';
import {
    LlmServiceSimpleLlmQueryRequest,
    LlmServiceGenerateTextContentRequest,
    LlmServiceReviewTextContentRequest,
    LlmServiceGenerateDefaultPromptsRequest,
    DesignServiceUpdateSectionRequest // Import update request type
} from './apiclient'; // Import request types

export class LlmInteractionHandler {
    private modal: Modal;
    private toastManager: ToastManager;
    private isLoading: boolean = false;

    // Store context while dialog is open
    private currentSectionData: SectionData | null = null;
    private currentApplyCallback: ((generatedText: string) => void) | null = null;
    private currentLlmResponseText: string | null = null;

    // Store original prompts to detect edits
    private originalGetAnswerPrompt: string = "";
    private originalVerifyAnswerPrompt: string = "";

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
            return;
        }
        console.log(`Showing LLM dialog for section: ${sectionData.id}`);

        this.currentSectionData = sectionData;
        // Store original prompts for comparison later
        this.originalGetAnswerPrompt = sectionData.getAnswerPrompt || "";
        this.originalVerifyAnswerPrompt = sectionData.verifyAnswerPrompt || "";

        this.currentApplyCallback = applyCallback || null;
        this.currentLlmResponseText = null;

        const modalContentElement = this.modal.show('llm-dialog', {
            onSubmit: this.handleDialogSubmit.bind(this)
        });

        setTimeout(() => {
            if (!modalContentElement) {
                 console.error("LLM Dialog template root element not found after show.");
                 this.modal.hide();
                 return;
            };
            this.updateCurrentSectionDisplay(modalContentElement, sectionData.title);
            this.setupTabsAndRefreshButtons(modalContentElement); // Renamed and updated
            this.configureDialog(modalContentElement, sectionData);
             // Ensure initial tab state is correct
            const initialActiveTab = modalContentElement.querySelector<HTMLButtonElement>('.llm-tab-generate');
             initialActiveTab?.click();
        }, 50);
    }

    /** Sets up tabs and binds "Refresh Default Prompt" buttons */
    private setupTabsAndRefreshButtons(modalContentElement: HTMLElement): void { // Renamed
        const tabs = modalContentElement.querySelectorAll<HTMLButtonElement>('.llm-tab');
        const panes = modalContentElement.querySelectorAll<HTMLElement>('.llm-tab-pane');
        const refreshGenerateBtn = modalContentElement.querySelector<HTMLButtonElement>('#llm-refresh-generate-prompt-btn'); // New ID
        const refreshVerifyBtn = modalContentElement.querySelector<HTMLButtonElement>('#llm-refresh-verify-prompt-btn'); // New ID

        tabs.forEach(tab => {
            tab.onclick = () => { // Use onclick for simplicity
                tabs.forEach(t => t.classList.remove('llm-tab-active', 'border-blue-500', 'text-blue-600', 'dark:text-blue-400', 'dark:border-blue-400'));
                tabs.forEach(t => t.classList.add('text-gray-500', 'hover:text-gray-700', 'dark:text-gray-400', 'dark:hover:text-gray-300', 'hover:border-gray-300', 'dark:hover:border-gray-500', 'border-transparent'));
                panes.forEach(p => p.classList.add('hidden'));

                tab.classList.add('llm-tab-active', 'border-blue-500', 'text-blue-600', 'dark:text-blue-400', 'dark:border-blue-400');
                tab.classList.remove('text-gray-500', 'hover:text-gray-700', 'dark:text-gray-400', 'dark:hover:text-gray-300', 'hover:border-gray-300', 'dark:hover:border-gray-500', 'border-transparent');

                const targetPaneId = tab.dataset.tabTarget;
                if (targetPaneId) {
                    const targetPane = modalContentElement.querySelector<HTMLElement>(targetPaneId);
                    targetPane?.classList.remove('hidden');
                }
            };
        });

        // Bind "Refresh Default" buttons
        if (refreshGenerateBtn) {
             refreshGenerateBtn.onclick = () => this.handleRefreshDefaultPromptClick('get_answer'); // Pass type
        }
        if (refreshVerifyBtn) {
             refreshVerifyBtn.onclick = () => this.handleRefreshDefaultPromptClick('verify_answer'); // Pass type
        }
    }

    /** Configures the LLM dialog - Populates textareas */
    private configureDialog(modalContentElement: HTMLElement, sectionData: SectionData): void {
        const generatePromptInput = modalContentElement.querySelector<HTMLTextAreaElement>('#llm-generate-prompt-input'); // New ID
        const verifyPromptInput = modalContentElement.querySelector<HTMLTextAreaElement>('#llm-verify-prompt-input'); // New ID
        const verifyTabButton = modalContentElement.querySelector<HTMLButtonElement>('.llm-tab-verify');

        const isTextSection = sectionData.type === 'text';
        verifyTabButton?.classList.toggle('hidden', !isTextSection);

        // Populate Generate Prompt Textarea
        if (generatePromptInput) {
             generatePromptInput.value = sectionData.getAnswerPrompt || ""; // Use original value
        }

        // Populate Verify Prompt Textarea
        if (isTextSection) {
            if (verifyPromptInput) {
                 verifyPromptInput.value = sectionData.verifyAnswerPrompt || ""; // Use original value
                 verifyPromptInput.disabled = false;
            }
        } else {
            if (verifyPromptInput) {
                verifyPromptInput.value = "Verification only available for text sections.";
                verifyPromptInput.disabled = true; // Disable textarea if not applicable
            }
        }
    }


    private updateCurrentSectionDisplay(modalContentElement: HTMLElement, sectionTitle: string): void {
         const currentSectionElement = modalContentElement.querySelector<HTMLElement>('#llm-current-section');
         if (currentSectionElement) currentSectionElement.textContent = sectionTitle;
    }


    /** Handles the submission logic: Saves edited prompts THEN runs LLM action */
    private async handleDialogSubmit(): Promise<void> {
        const sectionData = this.currentSectionData;
        if (this.isLoading || !sectionData) {
            console.warn("LLM submit called while loading or without section data.");
            return;
        }

        const dialogContent = this.modal.getContentElement();
        if (!dialogContent) {
            console.error("LLM Dialog content element not found during submit.");
            return;
        }

        this.isLoading = true;
        const submitButton = dialogContent.querySelector<HTMLButtonElement>('button[data-modal-action="submit"]');
        this.showLoadingOnButton(submitButton, true); // Show loading

        try {
            // --- Step 1: Save any edited prompts ---
            const activeTab = dialogContent.querySelector<HTMLButtonElement>('.llm-tab-active');
            let promptSaveError = null;

            if (activeTab?.dataset.tabTarget === '#llm-generate-content') {
                const generatePromptInput = dialogContent.querySelector<HTMLTextAreaElement>('#llm-generate-prompt-input');
                const currentPromptText = generatePromptInput?.value ?? ""; // Get current value
                if (currentPromptText !== this.originalGetAnswerPrompt) { // Compare with original
                    console.log("Generate prompt edited, saving...");
                    promptSaveError = await this.savePromptUpdate('get_answer_prompt', currentPromptText);
                }
            } else if (activeTab?.dataset.tabTarget === '#llm-verify-content' && sectionData.type === 'text') {
                const verifyPromptInput = dialogContent.querySelector<HTMLTextAreaElement>('#llm-verify-prompt-input');
                const currentPromptText = verifyPromptInput?.value ?? "";
                if (currentPromptText !== this.originalVerifyAnswerPrompt) {
                    console.log("Verify prompt edited, saving...");
                    promptSaveError = await this.savePromptUpdate('verify_answer_prompt', currentPromptText);
                }
            }
             // No need to save for 'custom' tab

            if (promptSaveError) {
                 throw promptSaveError; // Stop if saving prompt failed
            }
             // --- End Step 1 ---


            // --- Step 2: Perform the main LLM action ---
            console.log(`LLM dialog submit initiated for section ${sectionData.id} - Action Phase`);
            const activePaneId = activeTab?.dataset.tabTarget;
            let apiCall: Promise<any>;
            let actionType: 'generate' | 'review' | 'custom' = 'custom';
            const pathParams = { designId: sectionData.designId, sectionId: sectionData.id };

            if (activePaneId === '#llm-generate-content') {
                actionType = 'generate';
                const request = { ...pathParams, body: {} };
                console.log("Calling GenerateTextContent API with request:", request);
                apiCall = LlmApi.llmServiceGenerateTextContent(request);
            } else if (activePaneId === '#llm-verify-content' && sectionData.type === 'text') {
                actionType = 'review';
                const request = { ...pathParams, body: {} };
                console.log("Calling ReviewTextContent API with request:", request);
                apiCall = LlmApi.llmServiceReviewTextContent(request);
            } else if (activePaneId === '#llm-custom-content') {
                actionType = 'custom';
                const customPromptInput = dialogContent.querySelector<HTMLTextAreaElement>('#custom-prompt');
                const customPrompt = customPromptInput?.value.trim() || '';
                if (!customPrompt) {
                    this.toastManager.showToast("Input Error", "Please enter a custom prompt.", "warning");
                    throw new Error("Custom prompt empty");
                }
                const request = { body: { designId: sectionData.designId, sectionId: sectionData.id, prompt: customPrompt } };
                console.log("Calling SimpleLlmQuery API (Custom) with request:", request);
                apiCall = LlmApi.llmServiceSimpleLlmQuery(request);
            } else {
                 throw new Error(`Invalid or unsupported LLM action for tab: ${activePaneId} and type: ${sectionData.type}`);
            }

            const response = await apiCall;
            console.log("LLM API Response:", response);
            this.currentLlmResponseText = response.generatedText || response.reviewText || response.responseText || null;

             // --- End Step 2 ---


            // --- Step 3: Show results ---
            this.showLoadingOnButton(submitButton, false); // Hide loading before showing results
            await this.modal.hide(); // Hide prompt dialog

            let applyCallbackForResults: ((data: any) => void) | undefined = undefined;
            if (actionType === 'generate' && sectionData.type === 'text' && this.currentApplyCallback) {
                applyCallbackForResults = this.handleApplyLlmResult.bind(this);
            }

            const resultsModalContent = this.modal.show('llm-results', { onApply: applyCallbackForResults });
            if (resultsModalContent) {
                setTimeout(() => {
                    const contentArea = resultsModalContent.querySelector<HTMLElement>('#llm-results-content');
                    if (contentArea) contentArea.textContent = this.currentLlmResponseText || "LLM returned no text.";
                    const applyButton = resultsModalContent.querySelector<HTMLButtonElement>('#llm-results-apply');
                    if (applyButton) applyButton.classList.toggle('hidden', !applyCallbackForResults);
                }, 50);
            }
             // --- End Step 3 ---

        } catch (error: any) {
            console.error("Error during LLM interaction (Submit):", error);
            const errorMsg = error.message || (error.response ? await error.response.text() : 'Unknown LLM API error');
            this.toastManager.showToast("LLM Error", `Failed: ${errorMsg}`, "error");
        } finally {
            this.isLoading = false;
            this.showLoadingOnButton(submitButton, false); // Ensure button is reset
        }
    }


     /** Handles click on "Refresh Default Prompt" buttons */
     private async handleRefreshDefaultPromptClick(promptType: 'get_answer' | 'verify_answer'): Promise<void> {
       const sectionData = this.currentSectionData;
       if (!sectionData) return;

       const dialogContent = this.modal.getContentElement();
       if (!dialogContent) return;

       // Determine which button and textarea to target
       const isGetAnswer = promptType === 'get_answer';
       const buttonId = isGetAnswer ? '#llm-refresh-generate-prompt-btn' : '#llm-refresh-verify-prompt-btn';
       const textareaId = isGetAnswer ? '#llm-generate-prompt-input' : '#llm-verify-prompt-input';
       const button = dialogContent.querySelector<HTMLButtonElement>(buttonId);
       const textarea = dialogContent.querySelector<HTMLTextAreaElement>(textareaId);

       if (!button || !textarea) {
         console.error(`Could not find button or textarea for prompt type: ${promptType}`);
         return;
       }

       this.showLoadingOnButton(button, true, 'Refreshing...'); // Show loading on the specific button
       this.toastManager.showToast("Generating...", "Generating default prompt...", "info", 2000);

       try {
           const request: LlmServiceGenerateDefaultPromptsRequest = {
               designId: sectionData.designId,
               sectionId: sectionData.id,
               body: {},
           };
           const response = await LlmApi.llmServiceGenerateDefaultPrompts(request);

           // Update the specific textarea and internal data
           const newPrompt = isGetAnswer ? response.getAnswerPrompt : response.verifyAnswerPrompt;
           textarea.value = newPrompt || "";

           if (isGetAnswer) {
                sectionData.getAnswerPrompt = newPrompt || "";
                this.originalGetAnswerPrompt = newPrompt || ""; // Update original to prevent save on submit
           } else {
                sectionData.verifyAnswerPrompt = newPrompt || "";
                this.originalVerifyAnswerPrompt = newPrompt || ""; // Update original
           }

           this.toastManager.showToast("Success", "Default prompt refreshed and saved.", "success");
         } catch (error: any) {
           console.error("Error refreshing default prompt:", error);
           this.toastManager.showToast("Error", `Failed to refresh prompt: ${error.message || 'Server error'}`, "error");
         } finally {
           this.showLoadingOnButton(button, false); // Reset the specific button
         }
      }

    /** Helper to save an updated prompt via API */
    private async savePromptUpdate(maskPath: 'get_answer_prompt' | 'verify_answer_prompt', newValue: string): Promise<Error | null> {
        if (!this.currentSectionData) {
             const err = new Error("Cannot save prompt: Section data missing.");
             console.error(err.message);
             return err;
        }
        const { designId, id: sectionId } = this.currentSectionData;

        console.log(`Saving prompt update for ${maskPath}...`);
        this.toastManager.showToast("Saving...", `Saving updated ${maskPath}...`, "info", 1500);

        // Construct the request body correctly based on the maskPath
        const sectionUpdatePayload: any = {};
        if (maskPath === 'get_answer_prompt') {
            sectionUpdatePayload.getAnswerPrompt = newValue;
        } else if (maskPath === 'verify_answer_prompt') {
             sectionUpdatePayload.verifyAnswerPrompt = newValue;
        }

        const request: DesignServiceUpdateSectionRequest = {
            sectionDesignId: designId,
            sectionId: sectionId,
            body: {
                section: sectionUpdatePayload, // Send only the field being updated
                updateMask: maskPath,          // Specify the field in the mask
            },
        };

        try {
             await DesignApi.designServiceUpdateSection(request);
             // Update internal state upon successful save
            if (maskPath === 'get_answer_prompt') {
                 this.currentSectionData.getAnswerPrompt = newValue;
                 this.originalGetAnswerPrompt = newValue; // Update original after successful save
            } else {
                 this.currentSectionData.verifyAnswerPrompt = newValue;
                 this.originalVerifyAnswerPrompt = newValue; // Update original
            }
             this.toastManager.showToast("Saved", "Prompt updated successfully.", "success", 1500);
             return null; // Indicate success
         } catch (error: any) {
             console.error(`Error saving prompt ${maskPath}:`, error);
             const errorMsg = error.message || (error.response ? await error.response.text() : 'Server error');
             this.toastManager.showToast("Save Error", `Failed to save prompt: ${errorMsg}`, "error");
             return new Error(`Failed to save prompt: ${errorMsg}`); // Return the error
         }
     }


     private showLoadingOnButton(button: HTMLButtonElement | null, isLoading: boolean, loadingText: string = "Processing...") {
         if (!button) return;
         if (isLoading) {
             button.disabled = true;
             button.innerHTML = `<svg class="animate-spin -ml-1 mr-2 h-4 w-4" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24"> <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle> <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path> </svg> ${loadingText}`;
         } else {
             button.disabled = false;
             // Restore original text - might need to store it if it's dynamic
             if (button.id === 'llm-dialog-submit') button.textContent = 'Submit';
             else if (button.id.startsWith('llm-refresh-')) button.innerHTML = `<svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"> <path stroke-linecap="round" stroke-linejoin="round" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m-15.357-2A8.001 8.001 0 0119.418 15m0 0h-5" /> </svg> Refresh Default`;
             else button.textContent = 'Submit'; // Default fallback
         }
     }


    private handleApplyLlmResult(): void {
        if (this.currentApplyCallback && this.currentLlmResponseText !== null) {
            console.log("LlmInteractionHandler: Applying result via stored callback...");
            this.currentApplyCallback(this.currentLlmResponseText);
        } else {
            console.error("Apply callback or LLM response text is missing.");
            this.toastManager.showToast("Error", "Could not apply result.", "error");
        }
         this.currentLlmResponseText = null;
    }
}

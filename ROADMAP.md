
**LLM Roadmap:**

**Phase 1: Foundation & Basic Interaction**

1.  **Backend: LLM Client Integration:**
    *   **Task:** Add a new Go package/service (`llm` or similar) responsible for interacting with a chosen LLM provider's API (e.g., OpenAI, Gemini).
    *   **Details:**
        *   Handle API key management (via environment variables loaded in `main.go`).
        *   Implement a basic function like `callLlm(prompt string) (string, error)`.
        *   Choose an initial LLM provider and model.
    *   **Difficulty:** Medium (requires external API integration, key handling).
    *   **Files:** New `services/llm/client.go`, `main.go` (for key env var).

2.  **Backend: Create `LlmService` gRPC Service:**
    *   **Task:** Define a new gRPC service (`LlmService`) with a simple initial RPC method (e.g., `SimpleLlmQuery`).
    *   **Details:**
        *   Define `LlmService.proto` with `SimpleLlmQueryRequest` (containing `prompt`, `design_id`, `section_id`) and `SimpleLlmQueryResponse` (containing `response_text`).
        *   Implement `LlmService` in Go (`services/llm_service.go`), which uses the LLM client from Step 1.
        *   Register `LlmService` in `services/server.go`.
        *   Expose `LlmService` via the gRPC Gateway in `web/api.go`.
        *   Regenerate Protobuf/Gateway code (`make proto`).
    *   **Difficulty:** Medium (Protobuf definitions, service implementation, Gateway registration).
    *   **Files:** `proto/leetcoach/v1/llm_service.proto`, `services/llm_service.go`, `services/server.go`, `web/api.go`, `gen/go/*`, `web/views/components/apiclient/*`.

3.  **Frontend: Basic Modal Wiring:**
    *   **Task:** Make the LLM button in `Section.html` open the existing `llm-dialog` modal and trigger a *very basic* API call. Display the raw response.
    *   **Details:**
        *   Modify `BaseSection.ts` (`openLlmDialog`): Ensure it correctly opens the `llm-dialog` modal.
        *   Add a basic `handleSubmit` function (initially tied to the modal's "Submit" button) in `BaseSection.ts` or a new `LlmDialog.ts` component.
        *   This handler should:
            *   Get the current section ID and design ID.
            *   Construct a *hardcoded* basic prompt (e.g., "Explain system design").
            *   Call the new `LlmService.simpleLlmQuery` endpoint using the generated API client.
            *   On success, show the `llm-results` modal (`Modal.ts`) and display the raw `response_text` inside `#llm-results-content`.
            *   Handle API errors (show Toast).
    *   **Difficulty:** Medium (Frontend API call, modal interaction logic).
    *   **Files:** `web/views/templates/Section.html`, `web/views/components/BaseSection.ts`, `web/views/components/Modal.ts`, `web/views/components/Api.ts` (add `LlmApi`), `web/views/templates/LlmDialog.html`, `web/views/templates/LlmResults.html`.

**Phase 2: Content Generation & Review (Text Sections)**

4.  **Backend: Enhance API for Text Content:**
    *   **Task:** Add specific RPCs for generating and reviewing *text* section content. Fetch content via `ContentService`.
    *   **Details:**
        *   Add RPCs like `GenerateTextContent(GenerateTextContentRequest{design_id, section_id, prompt_template})` and `ReviewTextContent(ReviewTextContentRequest{design_id, section_id})`.
        *   Implement these in `LlmService`.
        *   `GenerateTextContent`: Fetch section title (`DesignService`), formulate a prompt using the title and `prompt_template`, call LLM.
        *   `ReviewTextContent`: Fetch current section content (`ContentService.GetContent`), formulate a "review this:" prompt, call LLM.
        *   Regenerate Protobuf/Gateway code.
    *   **Difficulty:** Medium (API design, inter-service calls: DesignService/ContentService).
    *   **Files:** `proto/leetcoach/v1/llm_service.proto`, `services/llm_service.go`, `services/designs.go` (maybe helper func), `services/content.go` (maybe helper func), `gen/go/*`, `web/views/components/apiclient/*`.

5.  **Frontend: Implement Generate/Review UI (Text Sections):**
    *   **Task:** Update the `llm-dialog` modal and associated TS logic to handle "Generate" and "Review" actions specifically for text sections.
    *   **Details:**
        *   Modify `LlmDialog.html` tab switching logic (can be basic JS for now).
        *   In `BaseSection.ts` or a dedicated `LlmDialogHandler.ts`:
            *   When the modal opens for a *text* section:
                *   Show appropriate pre-defined prompts based on the section title (e.g., in the "Generate" tab).
                *   Enable the "Review" tab.
            *   On "Generate" submit: Call the new `GenerateTextContent` RPC.
            *   On "Review" submit: Call the new `ReviewTextContent` RPC.
            *   Display results in the `llm-results` modal.
            *   Add an "Apply" button to `llm-results` modal: If the user clicks Apply after *generating* content, update the `TextSection`'s editor content (if in edit mode) or saved content (`ContentService.SetContent`) and switch to view mode.
    *   **Difficulty:** Medium-High (UI logic, state management in modal, interacting with `TextSection`, `ContentService` API).
    *   **Files:** `web/views/templates/LlmDialog.html`, `web/views/templates/LlmResults.html`, `web/views/components/BaseSection.ts` (or new `LlmDialogHandler.ts`), `web/views/components/TextSection.ts`, `web/views/components/Modal.ts`, `web/views/components/Api.ts`.

**Phase 3: Contextual Awareness & Drawing Support**

6.  **Backend: Context Gathering Service:**
    *   **Task:** Create a helper function or potentially enhance `LlmService` RPCs to gather context (e.g., other section titles, summaries of other sections' content) based on a request.
    *   **Details:**
        *   Define required context (e.g., all section titles, content summary of section X).
        *   Implement logic using `DesignService` (for titles/IDs) and `ContentService` (for content, potentially summarizing it).
        *   Modify relevant `LlmService` RPCs (or add new ones) to accept context flags and pass the gathered context within the prompt to the LLM.
    *   **Difficulty:** Medium-High (Context definition, potential summarization logic, efficient fetching).
    *   **Files:** `services/llm_service.go`, `services/designs.go`, `services/content.go`.

7.  **Frontend & Backend: Generate Interview Questions (Text Section):**
    *   **Task:** Implement the feature to generate interview questions based on the current text section's content.
    *   **Details:**
        *   Add `GenerateQuestions` RPC to `LlmService` (takes `design_id`, `section_id`).
        *   Backend implementation fetches content (`ContentService`) and calls LLM with appropriate prompt ("Generate interview questions about the following text: ...").
        *   Frontend: Add a "Questions" tab/option in `llm-dialog`. Call the new RPC and display results in `llm-results`.
    *   **Difficulty:** Medium (Builds on existing patterns).
    *   **Files:** `proto/leetcoach/v1/llm_service.proto`, `services/llm_service.go`, `web/views/templates/LlmDialog.html`, `web/views/templates/LlmResults.html`, `web/views/components/BaseSection.ts` (or handler), `web/views/components/Api.ts`.

8.  **Frontend & Backend: Review Drawing (via Text Description):**
    *   **Task:** Allow users to describe their drawing and get LLM feedback.
    *   **Details:**
        *   Frontend (`llm-dialog` for DrawingSection): Add a textarea for the user to describe their diagram.
        *   Enhance `ReviewTextContent` RPC or add `ReviewDrawingDescription` RPC to accept this text.
        *   Backend calls LLM with a prompt tailored to reviewing architecture descriptions.
        *   Display feedback in `llm-results`.
    *   **Difficulty:** Medium (Mainly UI change, backend logic is similar to text review).
    *   **Files:** `web/views/templates/LlmDialog.html`, `web/views/components/BaseSection.ts` (or handler), `proto/leetcoach/v1/llm_service.proto`, `services/llm_service.go`.

**Phase 4: Advanced Features & Refinements**

9.  **Frontend & Backend: Suggest Next Sections:**
    *   Implement API/Logic as described in Step 5 of prior thought process. Integrate with "+" button flow.
10. **Progressive Hints:** Requires more sophisticated state management and potentially a structured knowledge base or more complex LLM prompting.
11. **Consistency Checks / Gap Analysis:** Requires robust context gathering (Step 6) and complex prompts.
12. **Advanced Drawing Integration:** Attempting to parse Excalidraw JSON -> Textual description automatically for review.

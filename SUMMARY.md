
**LeetCoach Project Summary (Folder-Based - Updated)**

**1. Overview:**

LeetCoach is a web application designed to help users prepare for system design interviews. Its core functionality involves creating, editing, and viewing multi-section "Design" documents. It supports different section types (Text, Drawing, Plot), user authentication (local & OAuth), utilizes a Go backend with a TypeScript/React frontend, and incorporates LLM-based features for suggesting new sections and generating/reviewing section content using customizable prompts.

**2. Root Directory (`./`)**

*   **Purpose:** Contains the main application entry point, build configurations (implied `Makefile`), environment setup (including LLM API keys), core application orchestration, and top-level documentation.
*   **Key Files:**
    *   `main.go`: Main executable entry point. Handles environment loading (`godotenv` - includes `OPENAI_API_KEY`, `OPENAI_MODEL`), flag parsing, initialization of `App`. Sets up logging (`slog`, `dev.go`).
    *   `app.go`: Defines the `App` struct responsible for managing and starting servers (gRPC, Web). Handles graceful shutdown.
    *   `.env`, `.env.dev`: Environment variable files.
    *   `SUMMARY.md`: This detailed summary document.
    *   `INSTRUCTIONS.md`: Development conventions.
    *   `dev.go`: Development utilities (`slog` pretty-printer).

**3. Services (`./services/`)**

*   **Purpose:** Defines and implements backend business logic via gRPC services. Handles data persistence, core operations, and LLM interactions.
*   **Key Files:**
    *   `server.go`: Initializes and starts the gRPC server, registering all services (`DesignService`, `ContentService`, `TagService`, `AdminService`, `LlmService`). Instantiates and injects dependencies (e.g., `LLMClient`, `DesignService`, `ContentService` into `LlmService`).
    *   `designs.go`: Implements `DesignService`. Manages Design metadata (filesystem `design.json`) and Section metadata (filesystem `sections/<id>/main.json`). Handles CRUD for designs and Add/Delete/Move/Title updates for sections. Crucially:
        *   Contains helpers for prompt file paths (`getSectionPromptPath`), reading (`readPromptFile`), and writing (`writePromptFile`) within section directories (`sections/<id>/prompts/`).
        *   `AddSection` accepts optional `initial_` prompts from the request and writes them to files.
        *   `GetSection`/`GetDesign` read `main.json`, then attempt to read corresponding prompt files (`get_answer.md`, `verify.md`) and populate fields on the in-memory Go `Section` struct *before* conversion to Protobuf.
    *   `content.go`: Implements `ContentService`. Manages section content bytes (`content.<name>`) in the filesystem. Used by `LlmService` to retrieve content for review.
    *   `tags.go`: Implements `TagService` (Datastore).
    *   `auth.go`: Implements internal auth logic (Datastore).
    *   `admin.go`: Implements `AdminService`.
    *   `llm_service.go`: Implements `LlmService`.
        *   `SimpleLlmQuery`: Basic prompt execution.
        *   `SuggestSections`: Generates Title, Type, Description, and default `GetAnswerPrompt`/`VerifyPrompt` text for new sections based on existing titles.
        *   `GenerateTextContent`/`ReviewTextContent`: Reads the appropriate prompt (`get_answer.md`/`verify.md`) from the filesystem (via `DesignService` helpers). If the prompt file doesn't exist, generates a default prompt in memory based on the section title. Calls `LLMClient`.
        *   `GenerateDefaultPrompts`: Generates default prompts based on title and *saves* them to the filesystem (`get_answer.md`, `verify.md`) via `DesignService` helpers.
    *   `models.go`: Defines Go structs. `Section` struct now includes non-persistent `GetAnswerPrompt` and `VerifyAnswerPrompt` fields (marked `json:"-"`) used for temporary storage after reading files.
    *   `converters.go`: Helper functions. `SectionToProto` now copies prompt fields directly from the Go struct (which were populated by the service reading files) to the proto message. No longer reads files itself.
    *   `clientmgr.go`: Manages Datastore client connections.
    *   `base.go`, `gcds.go`, `idgen.go`, `constants.go`: Utilities, Datastore wrapper, ID generation, constants.
    *   `designs_test.go`: Tests for `DesignService`.

**4. LLM Client (`./services/llm/`)**

*   **Purpose:** Encapsulates interaction with external LLM APIs.
*   **Key Files:**
    *   `client.go`: Defines `LLMClient` interface. Provides `openaiClient` implementation. Includes `MockLLMClient`.
    *   `client_test.go`: Unit tests for client/mock.

**5. Protobuf Definitions (`./proto/leetcoach/v1/`)**

*   **Purpose:** Defines gRPC service contracts and messages.
*   **Key Files:**
    *   `models.proto`: `Section` message now has `get_answer_prompt` and `verify_answer_prompt` (string, non-optional) to hold current prompt text read from files.
    *   `designs.proto`: `AddSectionRequest` has optional `initial_get_answer_prompt`, `initial_verify_prompt` (strings).
    *   `llm_service.proto`: `SuggestedSection` includes `get_answer_prompt`, `verify_answer_prompt`. Added `GenerateDefaultPrompts` RPC and messages.

**6. Web Layer (`./web/`)**

*   **Purpose:** HTTP handling, API gateway, auth, sessions, SSR.
*   **Key Files:** (Largely unchanged by recent prompt logic) `server.go`, `app.go`, `api.go`, `user.go`.

**7. Server-Side Views & Logic (`./web/views/`)**

*   **Purpose:** SSR using Go templates.
*   **Key Files:** (Largely unchanged by recent prompt logic) `main.go`, `views.go`, `HomePage.go`, `DesignEditorPage.go`, etc.

**8. HTML Templates (`./web/views/templates/`)**

*   **Purpose:** Go templates and client-side template definitions.
*   **Key Files:**
    *   `TemplateRegistry.html`: **(Updated)**
        *   `llm-dialog`: Displays the current prompt (fetched via `Section` proto) for Generate/Verify tabs. Includes "Generate Default Prompt" buttons, shown if the current prompt is empty. Tab setup remains.
        *   `llm-results`: Includes conditional "Apply" button.
        *   `section-type-selector`: Includes "Suggest Sections" flow elements.
        *   Section view/edit templates remain.
    *   Other templates (`BasePage.html`, `HomePage.html`, `Section.html`, etc.) largely unchanged.

**9. Frontend Components (`./web/views/components/`)**

*   **Purpose:** TypeScript managing interactive frontend UI.
*   **Key Files:**
    *   `BaseSection.ts`: **(Refactored)** Abstract base. Uses `LlmInteractionHandler` and `FullscreenHandler`. Consolidated event binding (`_bindEvents`) and element finding (`_findElement`). Manages view/edit modes via `_renderCurrentMode`. Abstract `updateInternalContent`, `getApplyCallback`.
    *   `TextSection.ts`, `DrawingSection.tsx`, `PlotSection.ts`: Concrete subclasses implementing abstract methods from `BaseSection`. Text/Drawing set `allowEditOnClick = false`.
    *   `LlmInteractionHandler.ts`: **(Updated)** Manages LLM dialog.
        *   `showLlmDialog` receives full `sectionData` (including current prompts).
        *   `configureDialog` displays current prompts or shows "Generate Default" button.
        *   Binds and handles "Generate Default Prompts" button click (calls `GenerateDefaultPrompts` API, updates UI on success).
        *   `handleDialogSubmit` calls appropriate backend RPC (Generate/Review/Custom); backend reads prompts from files for Generate/Review.
        *   Manages results modal display and "Apply" callback.
    *   `FullscreenHandler.ts`: Manages fullscreen logic.
    *   `SectionManager.ts`: Handles "Suggest Sections" flow, passes `initial_` prompts to `AddSection` API when creating from suggestion.
    *   `TemplateLoader.ts`: Added `loadInto` helper.
    *   `Modal.ts`: Uses `loadInto`. Handles `onSubmit`/`onApply`.
    *   `Api.ts`: Exports API clients.
    *   Other components (`DesignEditorPage.ts`, etc.) largely unchanged by prompt logic.

**10. Auto-Generated API Client (`./web/views/components/apiclient/`)**

*   **Purpose:** Generated TypeScript client library.
*   **Key Files:** Updated with latest proto changes (Section prompts, AddSection initial prompts, GenerateDefaultPrompts RPC).

**11. Static Assets (`./web/static/`)**

*   **Purpose:** Static files.
*   **Key Files:** `tailwind.css` updated for fullscreen height.

**12. Attic (`./.attic/`)**

*   **Purpose:** Older/deprecated code.

**13. Key Data Flows (Updated):**

*   **Suggest Sections:** Flow updated to include prompts in suggestions and pass initial prompts during `AddSection`.
*   **LLM Interaction (Generate/Review):** Section LLM Button -> `LlmInteractionHandler.showLlmDialog` (passes section data with current prompts) -> Modal UI displays current prompt or "Generate Default" button -> (Optional: Click Generate Default -> `GenerateDefaultPrompts` API -> Backend saves files -> UI updates) -> Click Submit -> `LlmInteractionHandler.handleDialogSubmit` -> Calls `GenerateTextContent`/`ReviewTextContent` API -> Backend `LlmService` reads prompt *file*, executes -> Response -> Results Modal -> (Optional Apply).


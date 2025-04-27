---

**LeetCoach Project Summary (Folder-Based - Updated)**

**1. Overview:**

LeetCoach is a web application designed to help users prepare for system design interviews. Its core functionality involves creating, editing, and viewing multi-section "Design" documents. It supports different section types (Text, Drawing, Plot), user authentication (local & OAuth), utilizes a Go backend with a TypeScript/React frontend, and now incorporates LLM-based features for content suggestion and generation/review.

**2. Root Directory (`./`)**

*   **Purpose:** Contains the main application entry point, build configurations (implied `Makefile`), environment setup (including LLM API keys), core application orchestration, and top-level documentation.
*   **Key Files:**
    *   `main.go`: Main executable entry point. Handles environment loading (`godotenv` - now includes `OPENAI_API_KEY`, `OPENAI_MODEL`), flag parsing, and initialization of `App`. Sets up logging (`slog`, `dev.go`).
    *   `app.go`: Defines the `App` struct responsible for managing and starting the different servers (gRPC, Web). Handles graceful shutdown signals.
    *   `.env`, `.env.dev`: Environment variable files.
    *   `SUMMARY.md`: This detailed summary document.
    *   `INSTRUCTIONS.md`: Development conventions and guidelines for the project.
    *   `dev.go`: Development utilities, specifically a pretty-printer for `slog`.

**3. Services (`./services/`)**

*   **Purpose:** Defines and implements the backend business logic via gRPC services. Handles data persistence, core operations, and LLM interactions.
*   **Key Files:**
    *   `server.go`: Initializes and starts the gRPC server, registering all defined services (`DesignService`, `ContentService`, `TagService`, `AdminService`, `LlmService`). Instantiates and injects dependencies (e.g., `LLMClient`, `DesignService`, `ContentService` into `LlmService`).
    *   `designs.go`: Implements `DesignService`. Manages Design metadata and Section metadata (ID, type, title). Uses the filesystem for persistence. Handles CRUD for designs and Add/Delete/Move/Title updates for sections. `AddSection` now accepts an optional `title`. Employs per-design mutexes. Contains helpers like `readDesignMetadata`, `writeDesignMetadata`, `readSectionData`, `writeSectionData`.
    *   `content.go`: Implements `ContentService`. Manages section content bytes in the filesystem. Used by `LlmService` to retrieve content for review.
    *   `tags.go`: Implements `TagService`. Manages tags via Datastore.
    *   `auth.go`: Implements internal auth logic via Datastore.
    *   `admin.go`: Implements `AdminService`.
    *   `llm_service.go`: **(New)** Implements `LlmService`. Handles RPCs like `SimpleLlmQuery`, `SuggestSections`, `GenerateTextContent`, `ReviewTextContent`. Uses the `LLMClient` interface, interacts with `DesignService` (for titles) and `ContentService` (for content). Contains logic for constructing prompts and parsing LLM responses (e.g., `parseSuggestions`).
    *   `models.go`: Defines Go structs (`Design`, `Section`, `ContentMetadata`, `Tag`, `User`, etc.).
    *   `converters.go`: Helper functions for converting between Go structs and Protobuf messages.
    *   `clientmgr.go`: Manages connections/clients for backend dependencies (Datastore).
    *   `base.go`: Basic utilities (e.g., `EnsureLoggedIn`).
    *   `gcds.go`: Generic Datastore wrapper.
    *   `idgen.go`: Logic for generating unique IDs.
    *   `constants.go`: Defines constants.
    *   `designs_test.go`: Unit/integration tests for `DesignService`.

**4. LLM Client (`./services/llm/`)**

*   **Purpose:** Encapsulates interaction with external LLM APIs.
*   **Key Files:**
    *   `client.go`: **(New)** Defines the `LLMClient` interface (`SimpleQuery`). Provides an `openaiClient` implementation using `go-openai`, reading configuration (`OPENAI_API_KEY`, `OPENAI_MODEL`) from environment variables. Includes `MockLLMClient` for testing.
    *   `client_test.go`: **(New)** Unit tests for client initialization and the mock client.

**5. Protobuf Definitions (`./proto/leetcoach/v1/`)**

*   **Purpose:** Defines the gRPC service contracts and message structures.
*   **Key Files:**
    *   `llm_service.proto`: **(New)** Defines `LlmService` with RPCs (`SimpleLlmQuery`, `SuggestSections`, `GenerateTextContent`, `ReviewTextContent`) and corresponding request/response messages (including `SuggestedSection`). Includes HTTP gateway annotations.
    *   `designs.proto`: Defines `DesignService` and related messages. `AddSectionRequest` now implicitly supports `title` via its `Section` field.
    *   `content.proto`, `tags.proto`, `models.proto`: Define other services and common types.

**6. Web Layer (`./web/`)**

*   **Purpose:** Handles HTTP requests, serves frontend, API gateway, auth, sessions.
*   **Key Files:**
    *   `server.go`: Starts HTTP server, middleware (logging, CORS).
    *   `app.go`: Defines `LCApp`, initializes session (`scs`), auth (`oneauth`), API (`LCApi`), Views (`LCViews`).
    *   `api.go`: Implements `LCApi` (gRPC Gateway). Proxies REST requests to backend gRPC services. Injects `LoggedInUserId` into gRPC metadata. Registers `LlmService` handler. Includes improved error handling display.
    *   `user.go`: Implements `oneauth.UserStore` bridge.

**7. Server-Side Views & Logic (`./web/views/`)**

*   **Purpose:** Server-side rendering using Go templates and data preparation logic.
*   **Key Files:**
    *   `main.go`: Initializes `LCViews`, template engine (`tmplr`), global functions, view routing.
    *   `views.go`: Main view routing logic.
    *   `HomePage.go`, `DesignEditorPage.go`, etc.: Page-specific Go structs and `Load` methods. (Largely unchanged by recent features).

**8. HTML Templates (`./web/views/templates/`)**

*   **Purpose:** Go `html/template` files for SSR and client-side template definitions.
*   **Key Files:**
    *   `BasePage.html`: Main layout.
    *   `HomePage.html`: Includes `DesignList.html`. Uses `create-design-modal`.
    *   `DesignEditorPage.html`: Includes `TableOfContents.html`, `DocumentTitle.html`, `SectionsList.html`.
    *   `TemplateRegistry.html`: **(Updated)** Crucial client-side templates:
        *   `section-type-selector`: Now includes "Suggest Sections" button, suggestion container/card template, loading indicator.
        *   `llm-dialog`: Includes tabs (Generate, Custom, Verify) and content panes.
        *   `llm-results`: Includes content display area and conditional "Apply" button.
        *   `text-section-view`, `text-section-edit`, `drawing-section-view`, `drawing-section-edit`, `plot-section-view`, `plot-section-edit`: Templates loaded by `BaseSection`.
    *   `Section.html`: Defines the basic section structure including header controls (like `.section-ai` LLM button).

**9. Frontend Components (`./web/views/components/`)**

*   **Purpose:** TypeScript code managing interactive frontend experience.
*   **Key Files:**
    *   `BaseSection.ts`: **(Refactored)** Abstract base. No longer contains direct LLM or Fullscreen logic. Uses `_findElement` for DOM querying, `_bindEvents` for consolidated event handling (delegation). Instantiates `LlmInteractionHandler` and `FullscreenHandler`. Delegates LLM button click to handler. Provides abstract `resizeContentForFullscreen` and `updateInternalContent`, and `getApplyCallback`. Manages view/edit mode switching via `_renderCurrentMode`.
    *   `TextSection.ts`: Concrete section using TinyMCE. Implements `updateInternalContent`, `applyGeneratedContent`, and overrides `getApplyCallback`.
    *   `DrawingSection.tsx`: Concrete section using Excalidraw. Sets `allowEditOnClick = false`. Implements `updateInternalContent`.
    *   `PlotSection.ts`: Concrete section placeholder. Sets `allowEditOnClick = false`. Implements `updateInternalContent`.
    *   `LlmInteractionHandler.ts`: **(New)** Manages LLM dialog lifecycle: showing, tab setup/state, prompt construction (basic), calling appropriate `LlmApi` methods (`SimpleLlmQuery`, `GenerateTextContent`, `ReviewTextContent`), handling loading states, displaying results via `llm-results` modal, handling "Apply" callback flow.
    *   `FullscreenHandler.ts`: **(New)** Manages fullscreen state (`isFullscreen`), DOM class manipulation (`lc-section-fullscreen`, etc.), event listeners (keydown, resize, button clicks), hides/shows toolbar buttons, calls `resizeCallback`. Has `destroy` method.
    *   `SectionManager.ts`: Manages section collection (add, delete, move). Handles "Suggest Sections" button click in `section-type-selector` modal: calls `LlmApi.llmServiceSuggestSections`, renders suggestions using `suggested-section-card` template, handles suggested card clicks by calling `handleSectionTypeSelection` with title. `handleSectionTypeSelection` now passes title to `DesignApi.designServiceAddSection`.
    *   `TemplateLoader.ts`: Utility class. Added `loadInto` helper method for loading template children directly into a target element.
    *   `Modal.ts`: Singleton modal manager. Updated `show` to use `TemplateLoader.loadInto`. Handles `onSubmit` and `onApply` callbacks via `data-modal-action`.
    *   `Api.ts`: Configures and exports API client instances, including `LlmApi`. Handles auth token injection.
    *   `DesignEditorPage.ts`, `HomePage.ts`, `LoginPage.ts`, `DocumentTitle.ts`, `TableOfContents.ts`, `ToastManager.ts`, `ThemeManager.ts`: Other components managing specific page/UI logic.
    *   `types.ts`, `converters.ts`: Core types and API/frontend type conversion utilities.

**10. Auto-Generated API Client (`./web/views/components/apiclient/`)**

*   **Purpose:** TypeScript client library generated from OpenAPI spec (derived from gRPC Gateway).
*   **Key Files:** Updated with `LlmServiceApi` and related request/response models.

**11. Static Assets (`./web/static/`)**

*   **Purpose:** Serves static files (CSS, JS bundles, images, libraries).
*   **Key Files:**
    *   `css/tailwind.css`: Compiled Tailwind CSS (includes fullscreen height fix).
    *   `js/gen/`: Output directory for Webpack bundles.

**12. Attic (`./.attic/`)**

*   **Purpose:** Older/deprecated code.

**13. Key Data Flows (Updated):**

*   **Authentication:** Unchanged.
*   **Design Loading/Saving:** Unchanged core flow, but `handleSaveClick` in `TextSection` now also updates design timestamp.
*   **New Design (Suggest):** "+" Button -> `SectionManager.openSectionTypeSelector` -> Modal shows -> Click "Suggest" -> `SectionManager.handleSuggestSectionsClick` -> `LlmApi.SuggestSections` -> Backend `LlmService.SuggestSections` -> LLM -> Response -> `SectionManager` renders suggestion cards -> Click Suggestion Card -> `SectionManager.handleSectionTypeSelection` (with title) -> `DesignApi.AddSection` -> Backend `DesignService.AddSection` (uses title).
*   **LLM Interaction (Simple Query Example):** Section LLM Button -> `BaseSection._bindEvents` -> `LlmInteractionHandler.showLlmDialog` -> Modal shows (`llm-dialog`) -> User interacts (selects tab, enters prompt) -> Click Submit -> `Modal` calls `LlmInteractionHandler.handleDialogSubmit` -> `LlmInteractionHandler` determines action -> Calls appropriate `LlmApi` method (e.g., `SimpleLlmQuery`) -> Backend `LlmService` -> `llm.LLMClient` -> LLM -> Response -> `LlmInteractionHandler` hides dialog -> Shows `llm-results` modal -> (Optional Apply Click) -> `Modal` calls `LlmInteractionHandler.handleApplyLlmResult` -> Calls `TextSection.applyGeneratedContent`.


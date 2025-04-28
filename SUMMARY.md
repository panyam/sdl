
**LeetCoach Project Summary **

**1. Overview:**

LeetCoach is a web application aimed at assisting users in preparing for system design interviews. It enables the creation, editing, and viewing of multi-section "Design" documents. Supported section types include Text (TinyMCE), Drawing (Excalidraw), and Plot (JSON config). The application features user authentication (local & OAuth via `oneauth`), a Go backend utilizing gRPC services, and a TypeScript frontend without a major framework (except for Excalidraw). It leverages LLM capabilities (via OpenAI) for suggesting sections, generating/reviewing content using customizable prompts stored on the filesystem, and generating default prompts. Persistence uses a hybrid approach: filesystem for core design/section data managed by `DesignStore`, and Google Cloud Datastore for users, tags, and auth entities.

**2. Root Directory (`./`)**

*   **Purpose:** Contains the main application entry point, build configurations (implied `Makefile`), environment setup (including `OPENAI_API_KEY`, `OPENAI_MODEL`), core application orchestration (`App`), logging setup, and top-level documentation.
*   **Key Files:**
    *   `main.go`: Main executable entry point. Loads environment (`godotenv`), parses flags, initializes `App`, sets up logging (`slog`, `dev.go`).
    *   `app.go`: Defines the `App` struct, responsible for managing and starting gRPC and Web servers, handling graceful shutdown.
    *   `.env`, `.env.dev`: Environment variable files (contains API keys, ports, etc.).
    *   `SUMMARY.md`: This detailed summary document.
    *   `INSTRUCTIONS.md`: Development conventions (TypeScript, Go templates, Tailwind, no major JS framework).
    *   `dev.go`: Development utilities (`slog` pretty-printer).

**3. Services (`./services/`)**

*   **Purpose:** Defines and implements backend business logic via gRPC. Handles data persistence logic (filesystem via `DesignStore`, Datastore via `ClientMgr`), core operations, and LLM interactions.
*   **Key Files:**
    *   `server.go`: Initializes and starts the gRPC server. **Instantiates `DesignStore`** and injects it where needed (`DesignService`, `ContentService`, `LlmService`). Instantiates `LLMClient` and injects it into `LlmService`. Registers all gRPC services.
    *   `designs_store.go`: **(New/Refactored)** Encapsulates ALL filesystem interactions for designs, sections, prompts, and content paths. Provides methods like `Read/WriteDesignMetadata`, `Read/WriteSectionData`, `Read/WritePromptFile`, `GetContentPath`, `DeleteDesign`, `DeleteSection`, `ListDesignIDs`. Ensures directories exist. Handles filename sanitization. Used by `DesignService`, `ContentService`, and `LlmService`.
    *   `designs.go`: **(Refactored)** Implements `DesignService`. Manages Design/Section *metadata manipulation* (e.g., adding/removing section IDs in `Design.SectionIds`). **Delegates all filesystem I/O (read/write/delete metadata, path generation) to `DesignStore`.** Handles logic for adding/deleting/moving sections *within the design's list*. Reads initial prompts from request during `AddSection` and writes them via `DesignStore`. Reads prompts via `DesignStore` for `GetDesign`/`GetSection` to populate transient fields on Go structs.
    *   `content.go`: **(Refactored)** Implements `ContentService`. Manages section *content bytes* (e.g., `content.main`, `content.light.svg`). **Delegates path logic to `DesignStore`.** Reads/writes content files. **Does NOT update metadata timestamps.**
    *   `llm_service.go`: **(Refactored)** Implements `LlmService`. Orchestrates LLM calls. **Reads section/design metadata and prompt files directly via `DesignStore`.** Reads content bytes via `ContentService`. Writes default prompts via `DesignStore` (`GenerateDefaultPrompts`). Uses the injected `LLMClient`.
    *   `tags.go`: Implements `TagService` interacting with Datastore via `ClientMgr`.
    *   `auth.go`: Implements internal user/identity/channel logic for authentication, interacting with Datastore via `ClientMgr`.
    *   `admin.go`: Implements `AdminService` (dev environment).
    *   `models.go`: Defines Go structs (e.g., `Design`, `Section`, `User`, `Tag`). `Section` struct includes non-persistent `GetAnswerPrompt`, `VerifyAnswerPrompt` fields (`json:"-"`) populated during reads.
    *   `converters.go`: Helper functions for converting between Go structs and Protobuf messages. `SectionToProto` copies prompt fields from the Go struct.
    *   `clientmgr.go`: Manages Google Cloud **Datastore** client connections and provides typed `DataStore` wrappers.
    *   `base.go`, `gcds.go`, `idgen.go`, `constants.go`: Base service utilities, Datastore wrapper logic, ID generation (filesystem based for designs/sections), constants (e.g., `defaultDesignsBasePath`).
    *   `designs_test.go`, `designs_store_test.go`, `content_test.go`: Unit tests for services and the store.

**4. LLM Client (`./services/llm/`)**

*   **Purpose:** Provides an abstraction layer over specific LLM API interactions (currently OpenAI).
*   **Key Files:**
    *   `client.go`: Defines `LLMClient` interface, `openaiClient` implementation using `go-openai`, and `MockLLMClient` for testing. Reads API key/model from environment variables.
    *   `client_test.go`: Unit tests for the client/mock.

**5. Protobuf Definitions (`./proto/leetcoach/v1/`)**

*   **Purpose:** Defines the gRPC service contracts (RPC methods) and message structures used for communication between frontend gateway and backend services.
*   **Key Files:**
    *   `models.proto`: Defines messages like `Design`, `Section` (includes `get_answer_prompt`, `verify_answer_prompt` strings).
    *   `designs.proto`: Defines `DesignService` RPCs (`CreateDesign`, `GetDesign`, `UpdateDesign`, `DeleteDesign`, `AddSection`, `UpdateSection`, `DeleteSection`, `MoveSection`). `AddSectionRequest` includes optional initial prompt strings.
    *   `content.proto`: Defines `ContentService` RPCs (`GetContent`, `SetContent`).
    *   `llm_service.proto`: Defines `LlmService` RPCs (`SimpleLlmQuery`, `SuggestSections`, `GenerateTextContent`, `ReviewTextContent`, `GenerateDefaultPrompts`). Includes `SuggestedSection` message with prompt fields.
    *   `tag_service.proto`, `admin_service.proto`: Service definitions for tags and admin functions.

**6. Web Layer (`./web/`)**

*   **Purpose:** Handles incoming HTTP requests, serves static assets, acts as the API gateway (gRPC-Gateway), manages user sessions and authentication flows.
*   **Key Files:**
    *   `server.go`: Starts the main HTTP web server, applies middleware (logging, CORS for dev).
    *   `app.go`: Defines `LCApp`. Initializes web application components: `scs` for sessions, `oneauth` for authentication handlers (local, Google, GitHub), `LCApi` (gRPC gateway), `LCViews` (template rendering). Connects `oneauth` to user storage logic.
    *   `api.go`: Configures and runs the `grpc-gateway`. Includes middleware to extract the logged-in user ID from the session and inject it into the outgoing gRPC metadata for backend services. Implements custom error handling for gateway responses.
    *   `user.go`: Implements the `oneauth.UserStore` interface, providing methods (`GetUserByID`, `EnsureAuthUser`, `ValidateUsernamePassword`) for `oneauth` to interact with the backend `AuthService` (via `ClientMgr`) or mock users.

**7. Server-Side Views & Logic (`./web/views/`)**

*   **Purpose:** Defines Go structs representing page data and handles the logic for loading data required by server-side rendered Go templates. Manages the template rendering pipeline.
*   **Key Files:**
    *   `main.go`: Defines `ViewContext`, `View` interface, `ViewRenderer` function. Initializes the `templar` template engine, loads templates from `./web/views/templates`, defines shared template functions.
    *   `views.go`: Sets up HTTP routes for different pages (e.g., "/", "/login", "/designs/{id}/edit"), mapping them to specific `View` implementations and template files.
    *   `HomePage.go`, `DesignEditorPage.go`, `DesignViewerPage.go`, `LoginPage.go`, `Header.go`, `DesignList.go`, `Paginator.go`: Define structs holding data for specific pages/components (e.g., `HomePage` holds `Header` and `DesignListView`). Implement the `Load` method to fetch necessary data from backend services (via `ClientMgr`) based on the request.
    *   `utils.go`: Helper functions (e.g., `randomDesignName`).

**8. HTML Templates (`./web/views/templates/`)**

*   **Purpose:** Contains Go HTML templates (`*.html`) for server-side rendering and the client-side template registry. Uses Tailwind CSS classes for styling.
*   **Key Files:**
    *   `BasePage.html`: Defines the main HTML document structure (`<html>`, `<head>`, `<body>`). Includes placeholders for header, body content, modals (`ModalContainer`), toasts (`ToastContainer`), and JavaScript includes. Sets up basic dark mode handling.
    *   `Header.html`: Template for the top navigation bar, including logo, site name, theme toggle button, mobile menu button, and placeholder for extra buttons.
    *   `ModalContainer.html`, `ToastContainer.html`: Define the structure for modal and toast UI elements.
    *   `TemplateRegistry.html`: **Crucial for client-side UI.** Contains definitions for dynamically loaded components:
        *   `llm-dialog`: LLM interaction modal with tabs (Generate, Custom, Verify), prompt textareas, "Refresh Default" buttons.
        *   `llm-results`: Modal to display LLM responses, with conditional "Apply" button.
        *   `section-type-selector`: Modal for choosing new section types (Text, Draw, Plot) and includes the "Suggest Sections" flow elements.
        *   `suggested-section-card`: Template for displaying an LLM-suggested section.
        *   `create-design-modal`: Modal for starting a new design (Blank or from template).
        *   Section View/Edit Templates (`text-section-view`, `text-section-edit`, `drawing-section-view`, etc.): HTML structure for displaying or editing the content of each section type.
    *   `Section.html`: Template for the wrapper around a single section, including header (number, title, type icon), controls (move, delete, add, settings, LLM, fullscreen), and the content area.
    *   `HomePage.html`, `DesignEditorPage.html`, `LoginPage.html`, `DesignList.html`, `TableOfContents.html`, `DocumentTitle.html`: Page-specific templates composing smaller components.

**9. Frontend Components (`./web/views/components/`)**

*   **Purpose:** TypeScript classes responsible for managing the interactive UI, handling user events, making API calls to the gRPC-Gateway, and manipulating the DOM based on templates loaded from `TemplateRegistry.html`.
*   **Key Files:**
    *   `Api.ts`: Configures and exports instances of the auto-generated API clients (`DesignApi`, `ContentApi`, `LlmApi`), including setting the base path (`/api`) and adding an interceptor to inject authentication tokens.
    *   `BaseSection.ts`: Abstract base class for sections. Handles view/edit mode switching, loading templates, common controls (title edit, LLM button, fullscreen via `FullscreenHandler`), event binding. Defines abstract methods for subclasses.
    *   `TextSection.ts`, `DrawingSection.tsx`, `PlotSection.ts`: Concrete implementations of `BaseSection`.
        *   `TextSection`: Integrates with TinyMCE for rich text editing, handles content saving/loading via `ContentApi`. Provides `applyGeneratedContent` callback for LLM results.
        *   `DrawingSection`: Integrates with Excalidraw (using React/ReactDOM), handles saving/loading drawing data (`main` JSON) and SVG previews (`light.svg`, `dark.svg`) via `ContentApi`. Handles theme changes for Excalidraw.
        *   `PlotSection`: Provides basic view/edit for plot configuration (JSON textarea).
    *   `SectionManager.ts`: Manages the collection of `BaseSection` instances on the editor page. Handles adding sections (calls `DesignApi.addSection`), deleting sections (`DesignApi.deleteSection`), moving sections (`DesignApi.moveSection`). Orchestrates the "Suggest Sections" flow (`LlmApi.suggestSections`). Updates the `TableOfContents`.
    *   `LlmInteractionHandler.ts`: Manages the LLM modal (`llm-dialog`, `llm-results`). Handles submitting requests to `LlmApi` (`GenerateTextContent`, `ReviewTextContent`, `SimpleLlmQuery`, `GenerateDefaultPrompts`). Handles saving edited prompts via `DesignApi.updateSection`.
    *   `DesignEditorPage.ts`: Entry point for the editor page. Initializes all managers (`ThemeManager`, `Modal`, `ToastManager`, `DocumentTitle`, `SectionManager`, `TableOfContents`). Handles the initial loading of design data (`DesignApi.getDesign`) and coordinates the loading of section content.
    *   `HomePage.ts`: Entry point for the design list page. Handles the "Create New" modal (`create-design-modal`).
    *   `LoginPage.ts`: Handles the logic for the login/signup form toggle.
    *   `TableOfContents.ts`: Manages the TOC sidebar UI, updates based on `SectionManager`, handles scrolling, mobile drawer.
    *   `DocumentTitle.ts`: Handles inline editing of the design title and calls `DesignApi.updateDesign`.
    *   `Modal.ts`: Singleton manager for showing/hiding modals based on templates from `TemplateRegistry.html`. Handles `onSubmit`/`onApply` callbacks.
    *   `ToastManager.ts`: Singleton manager for displaying toast notifications.
    *   `ThemeManager.ts`: Manages light/dark/system theme switching and updates relevant UI elements (like toggle button icon). Notifies sections of theme changes.
    *   `TemplateLoader.ts`: Utility for loading HTML content from `TemplateRegistry.html`.
    *   `FullscreenHandler.ts`: Manages entering/exiting fullscreen mode for sections.
    *   `converters.ts`: Utility functions for mapping between API enums/structs and frontend types (e.g., `V1SectionType` <-> `SectionType`).
    *   `types.ts`: Defines core TypeScript interfaces and types used across frontend components (`SectionData`, `SectionContent`, `SectionType`, etc.).

**10. Auto-Generated API Client (`./web/views/components/apiclient/`)**

*   **Purpose:** Contains the TypeScript client library automatically generated from the Protobuf definitions (likely using `openapi-generator` with the gRPC-Gateway output). Provides typed methods for calling the backend API.
*   **Key Files:** Generated files reflecting the gRPC service definitions.

**11. Static Assets (`./web/static/`)**

*   **Purpose:** Stores static files served directly by the web server.
*   **Key Files:**
    *   `css/tailwind.css`: Compiled Tailwind CSS output.
    *   `js/gen/`: Location for bundled JavaScript from TypeScript components (implied by `INSTRUCTIONS.md`) and potentially third-party JS assets like TinyMCE skins/plugins.
    *   Images, icons, fonts if any.

**12. Attic (`./.attic/`)**

*   **Purpose:** Contains older or deprecated code for reference, not actively used.


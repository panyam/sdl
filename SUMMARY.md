---

**LeetCoach Project Summary (Folder-Based)**

**1. Overview:**

LeetCoach is a web application designed to help users prepare for system design interviews. Its core functionality involves creating, editing, and viewing multi-section "Design" documents. It supports different section types (Text, Drawing, Plot), user authentication (local & OAuth), and utilizes a Go backend with a TypeScript/React frontend.

**2. Root Directory (`./`)**

*   **Purpose:** Contains the main application entry point, build configurations (implied `Makefile`), environment setup, core application orchestration, and top-level documentation.
*   **Key Files:**
    *   `main.go`: Main executable entry point. Handles environment loading (`godotenv`), flag parsing, and initialization of `App`. Sets up logging (`slog`, `dev.go`).
    *   `app.go`: Defines the `App` struct responsible for managing and starting the different servers (gRPC, Web). Handles graceful shutdown signals.
    *   `.env`, `.env.dev`: Environment variable files (API keys, ports, etc.).
    *   `SUMMARY.md`: This detailed summary document.
    *   `INSTRUCTIONS.md`: Development conventions and guidelines for the project.
    *   `dev.go`: Development utilities, specifically a pretty-printer for `slog`.

**3. Services (`./services/`)**

*   **Purpose:** Defines and implements the backend business logic via gRPC services. Handles data persistence and core operations.
*   **Key Files:**
    *   `server.go`: Initializes and starts the gRPC server, registering all defined services (`DesignService`, `ContentService`, `TagService`, `AdminService`).
    *   `designs.go`: Implements `DesignService`. Manages Design metadata (name, owner, visibility, section order) and Section metadata (ID, type, title). Uses the filesystem for persistence (`<basePath>/<designId>/design.json`, `<basePath>/<designId>/sections/<sectionId>/main.json`). Handles CRUD for designs and Add/Delete/Move/Title updates for sections. Employs per-design mutexes (`sync.Map`) for concurrency control.
    *   `content.go`: Implements `ContentService`. Manages the storage/retrieval of actual section content (raw bytes) in the filesystem (`<basePath>/<designId>/sections/<sectionId>/content.<name>`). Works closely with `DesignService` for path management and potentially uses the same mutexes.
    *   `tags.go`: Implements `TagService`. Manages tags associated with designs. Uses Google Cloud Datastore via `ClientMgr`.
    *   `auth.go`: Implements authentication-related logic (user creation/lookup on OAuth callback). Uses Datastore via `ClientMgr`. Not exposed as a direct gRPC service but used internally by the web layer.
    *   `admin.go`: Implements `AdminService` (likely for development/debugging purposes).
    *   `models.go`: Defines the Go structs (`Design`, `Section`, `ContentMetadata`, `Tag`, `User`, `Identity`, `Channel`, `AuthFlow`, `BaseModel`, `StringMapField`) used by the services, including Datastore tags.
    *   `converters.go`: Contains helper functions to convert between Go service structs (`models.go`) and gRPC Protobuf message types (`*.pb.go`).
    *   `clientmgr.go`: Manages connections/clients for backend dependencies, primarily Google Cloud Datastore. Provides typed `DataStore` wrappers (`TagDS`, `UserDS`, etc.). Also used by the web layer to interact with services.
    *   `base.go`: Basic utilities for services, like extracting the logged-in user ID from gRPC context metadata.
    *   `gcds.go`: Generic wrapper (`DataStore[T]`) for interacting with Google Cloud Datastore, simplifying common operations (Get, Save, Query, Delete). Defines `StringMapField` for flexible map storage.
    *   `idgen.go` (Implied): Logic for generating unique IDs (used in `DesignService` for new designs/sections).
    *   `constants.go`: Defines constants like default paths and feature flags (e.g., `ENFORCE_LOGIN`).
    *   `designs_test.go`: Unit/integration tests specifically for the `DesignService`, focusing on filesystem interactions and metadata manipulation.

**4. Web Layer (`./web/`)**

*   **Purpose:** Handles HTTP requests, serves the frontend application, provides the API gateway, and manages user authentication and sessions.
*   **Key Files:**
    *   `server.go`: Defines the `web.Server` struct, responsible for starting the main HTTP server. Includes middleware for logging (`withLogger`) and CORS (for development).
    *   `app.go`: Defines `LCApp`, the core web application struct. Initializes session management (`scs`), authentication (`oneauth` with Google/GitHub/Local providers), API (`LCApi`), and Views (`LCViews`). Sets up routing for `/auth`, `/api`, `/static`, and `/`.
    *   `api.go`: Implements the `LCApi` which sets up the gRPC Gateway (`runtime.ServeMux`). This gateway proxies RESTful JSON requests under `/api/v1/` to the backend gRPC services. Crucially, it injects the authenticated `LoggedInUserId` (obtained via `oneauth` middleware) into the outgoing gRPC metadata.
    *   `user.go`: Implements the `oneauth.UserStore` interface for `LCApp`, bridging `oneauth` with the backend `AuthService` (via `ClientMgr`) to fetch/create users during login/callback flows. Includes mock user logic for testing.

**5. Server-Side Views & Logic (`./web/views/`)**

*   **Purpose:** Handles the server-side rendering of HTML pages using Go templates and manages the logic for preparing data for these views.
*   **Key Files:**
    *   `main.go`: Initializes the `LCViews` handler and the `tmplr` template rendering engine. Defines global template functions. Sets up routing for view-related paths (`/`, `/designs/...`, `/login`, etc.).
    *   `views.go`: Defines the main routing logic within the views layer, mapping URL paths to specific view handlers (e.g., `/` maps to `HomePage`).
    *   `HomePage.go`, `DesignEditorPage.go`, `LoginPage.go`, `DesignViewerPage.go`, `DesignList.go`, etc.: Define the Go structs corresponding to each page template. Their `Load` methods fetch necessary data (e.g., designs from `DesignService` via `ClientMgr`) required by the template before rendering.
    *   `Header.go`, `Paginator.go`: Reusable Go structs/logic for parts of views (like the header data or pagination calculations).
    *   `GenericPage.go`, `BasePage.go`: Base structs providing common fields (like `Title`, `Header`) for page views.

**6. HTML Templates (`./web/views/templates/`)**

*   **Purpose:** Contains the Go `html/template` files used for server-side rendering.
*   **Key Files:**
    *   `BasePage.html`: The main layout template, includes header, modal container, toast container, and defines blocks (`BodySection`, `ExtraHeadSection`, etc.) for content injection.
    *   `HomePage.html`: Template for the main design listing page. Includes `DesignList.html`.
    *   `DesignEditorPage.html`: Template for the design editing interface. Includes `TableOfContents.html`, `DocumentTitle.html`, `SectionsList.html`. Defines `ExtraHeaderButtons` block.
    *   `LoginPage.html`: Template for the login/signup page.
    *   `Header.html`, `DesignList.html`, `TableOfContents.html`, `Section.html`, `DocumentTitle.html`, `SectionsList.html`: Component templates included within page templates.
    *   `ModalContainer.html`, `ToastContainer.html`: Wrappers for modal and toast UI elements.
    *   `TemplateRegistry.html`: **Crucial:** Contains *client-side* HTML templates (section view/edit modes, modals like `create-design-modal`, `section-type-selector`, `llm-dialog`) identified by `data-template-id`. These are cloned and used by frontend JavaScript (`TemplateLoader.ts`).
    *   `LlmDialog.html`, `LlmResults.html`, `SectionTypeSelector.html`: Templates specifically for modals, included within `ModalContainer.html` and defined within `TemplateRegistry.html`.

**7. Frontend Components (`./web/views/components/`)**

*   **Purpose:** Contains the TypeScript (and some React TSX) code that manages the interactive frontend experience, primarily on the Design Editor page.
*   **Key Files:**
    *   `DesignEditorPage.ts`: Main orchestrator for the editor page (`/designs/{id}/edit`). Initializes all other components, loads initial design data (metadata via `DesignApi`), and triggers section content loading.
    *   `HomePage.ts`: Manages logic for the listing page (`/`), specifically handling the "Create New" button click to show the `create-design-modal` and then redirecting based on user choice.
    *   `LoginPage.ts`: Handles simple UI toggling between Sign In and Sign Up modes on the login page.
    *   `SectionManager.ts`: Manages the collection of sections (add, delete, move, reorder). Interacts with `DesignService` for structural changes and updates the `TableOfContents`. No longer handles content loading/saving.
    *   `BaseSection.ts`: Abstract base class for sections. Handles common UI (header, controls, title editing), view/edit mode switching via `TemplateLoader`, fullscreen toggling, and initiates content loading (`loadContent`) via `ContentService`. Provides hooks for subclasses.
    *   `TextSection.ts`: Concrete section for rich text using TinyMCE. Implements content loading/saving (`refreshContentFromServer`, `handleSaveClick`) and editor initialization/destruction.
    *   `DrawingSection.tsx`: Concrete section for diagrams using Excalidraw (via `ExcalidrawWrapper.tsx`). Implements content loading/saving (including SVG previews) and React component mounting/unmounting.
    *   `PlotSection.ts`: Placeholder concrete section for plots.
    *   `ExcalidrawWrapper.tsx`: React component specifically to wrap and manage the `@excalidraw/excalidraw` library instance.
    *   `DocumentTitle.ts`: Manages the display and editing of the main Design title, interacts with `DesignService` for updates.
    *   `TableOfContents.ts`: Renders and manages interactions with the sidebar Table of Contents.
    *   `Modal.ts`: Singleton class to manage showing/hiding modal dialogs, loading content from `TemplateRegistry.html` via `TemplateLoader`.
    *   `ToastManager.ts`: Singleton class to display brief notification messages.
    *   `ThemeManager.ts`: Manages light/dark/system theme switching using `localStorage` and updates CSS classes on the `<html>` element. Notifies sections of changes.
    *   `TemplateLoader.ts`: Utility class to load and clone HTML templates from the `#template-registry` div based on `data-template-id`.
    *   `types.ts`: Defines core TypeScript interfaces and types used across frontend components (`SectionType`, `SectionData`, `DrawingContent`, etc.).
    *   `converters.ts`: Frontend utility functions to map between API models (`V1Section`, `V1SectionType`) and frontend types (`SectionData`, `SectionType`).
    *   `Api.ts`: Configures and exports instances of the auto-generated API clients (`DesignApi`, `ContentApi`, `TagApi`). Includes a `fetchApi` interceptor to automatically add the `Authorization: Bearer <token>` header by reading the `LeetCoachAuthToken` cookie.
    *   `samples.ts`: Sample data for testing/prototyping.

**8. Auto-Generated API Client (`./web/views/components/apiclient/`)**

*   **Purpose:** Contains the TypeScript client library automatically generated from the OpenAPI specification derived from the gRPC Gateway. Provides typed access to the backend API.
*   **Key Files:**
    *   `runtime.ts`: Core runtime logic for the generated client (fetch wrapper, request building, error handling).
    *   `apis/`: Contains individual API client classes (`DesignServiceApi.ts`, `ContentServiceApi.ts`, `TagServiceApi.ts`).
    *   `models/`: Contains TypeScript interfaces corresponding to the Protobuf messages used in the API requests and responses.
    *   `index.ts`: Exports all APIs and models for easy importing.

**9. Static Assets (`./web/static/`)**

*   **Purpose:** Serves static files like CSS, bundled JavaScript, images, and third-party libraries (like TinyMCE assets).
*   **Key Files (Examples):**
    *   `css/tailwind.css`: Compiled Tailwind CSS.
    *   `js/gen/`: Output directory for Webpack bundles (e.g., `HomePage.js`, `DesignEditorPage.js`).
    *   `js/gen/tinymce/`: Static assets required by TinyMCE (skins, icons, etc.).

**10. Attic (`./.attic/`)**

*   **Purpose:** Contains older or deprecated code that might be referenced later but is not currently active (e.g., previous implementations of drawing handling).

**11. Key Data Flows:**

*   **Authentication:** User logs in via Web UI -> Go `oneauth` handles flow -> Session/Cookie created -> Frontend reads cookie (`LeetCoachAuthToken`) -> `Api.ts` adds Bearer token to API calls -> `api.go` (Gateway) reads token/session -> Injects `LoggedInUserId` into gRPC metadata -> Backend services (`EnsureLoggedIn`) read metadata for authorization.
*   **Design Loading:** Server renders `DesignEditorPage.html` with `DesignId` -> `DesignEditorPage.ts` calls `DesignApi.getDesign` (including metadata) -> `SectionManager` creates shells -> `DesignEditorPage.ts` iterates, calls `section.loadContent()` -> Each section calls `ContentApi.getContent` -> Renders view.
*   **Content Saving:** User clicks save in section -> `SectionSubclass.handleSaveClick` -> Gets content from editor -> Calls `ContentApi.setContent` (for `main`, `light.svg`, etc.) -> Calls `switchToViewMode(true)`.
*   **Metadata Saving:** Title edits call `DesignApi.updateDesign` or `DesignApi.updateSection`. Add/Delete/Move call corresponding `DesignApi` methods.
*   **New Design (Current):** User clicks button -> `HomePage.ts` shows modal -> User clicks template/blank card -> `HomePage.ts` redirects browser to `/designs/new` or `/designs/new?templateId=...` -> `DesignEditorPage.go` handles `/designs/new`, calls backend `DesignService.CreateDesign` (template handling TBD), gets new ID -> Redirects to `/designs/{id}/edit`.

---

**Prompt for Next LLM:**

```
You are an expert software developer tasked with continuing work on the LeetCoach web application. You have been provided with a detailed project summary organized by folder structure, outlining the technologies (Go, gRPC, gRPC-Gateway, Go Templates, HTMX, TypeScript, Tailwind CSS, React for Excalidraw), architecture, components, data persistence (filesystem, Datastore), and key data flows.

**Context:** The project allows users to create multi-section system design documents. A recent refactor introduced a dedicated `ContentService` for handling raw section content, separating it from the `DesignService` which manages metadata. The current focus is on improving the "Create New Design" user experience.

**Current Task State:**
1.  A modal dialog (`create-design-modal`) has been added to the `TemplateRegistry.html`.
2.  The "Create New" button on the `HomePage.html` (design listing) now triggers this modal via `HomePage.ts`.
3.  When the user clicks an option (e.g., "Blank Design", "API Design Template") in the modal, `HomePage.ts` now performs a browser **redirect** to `/designs/new` (for blank) or `/designs/new?templateId={id}` (for templates).

**Project Conventions (Review `INSTRUCTIONS.md` if necessary):**
*   Frontend: Primarily TypeScript components (`./web/views/components/`), HTML templates (`./web/views/templates/`), Tailwind CSS. React/TSX only for specific libraries like Excalidraw. Client-side templates are stored in `TemplateRegistry.html` and loaded via `TemplateLoader.ts`. API interaction via auto-generated client in `apiclient/`.
*   Backend: Go with gRPC services (`./services/`) exposed via gRPC-Gateway (`./web/api.go`). Server-side rendering via Go templates (`./web/views/`, `./web/views/templates/`). Filesystem for design/content persistence, Datastore for tags/auth.
*   Work incrementally. Generate code changes for specific files. Explain your reasoning. Ask clarifying questions if the requirements or existing structure are unclear.

**Your Next Task:**

Modify the backend Go code to handle the `templateId` query parameter passed during the redirect to `/designs/new`.

1.  **Modify `web/views/DesignEditorPage.go`:** In the `Load` method, when `v.DesignId == ""`, check `r.URL.Query().Get("templateId")`.
2.  **Pass `templateId` to Backend:** Modify the call to `client.CreateDesign` within `DesignEditorPage.Load` to include the extracted `templateId`. This will require updating the `CreateDesignRequest` protobuf message and regenerating the Go/TS code.
3.  **Modify `services/designs.go` (`CreateDesign`):**
    *   Accept the `templateId` from the request.
    *   If `templateId` is present:
        *   **Define Template Structure:** Decide how templates will be represented (e.g., a Go struct defining sections, titles, initial content).
        *   **Load Template Definition:** Implement logic to load the specified template definition (e.g., from embedded files, a dedicated `./data/templates` directory with JSON files).
        *   **Apply Template:** Instead of creating just an empty design, create the `design.json` with the template's name/description and the list of section IDs defined in the template.
        *   **Create Sections:** Loop through the sections defined in the template:
            *   Generate section IDs.
            *   Create the section subdirectories (`<basePath>/<designId>/sections/<sectionId>`).
            *   Create the `main.json` file for each section with its metadata (type, title).
            *   **Crucially:** Create the initial *content* files (e.g., `content.main`) within each section directory, populating them with the initial content specified in the template definition. This might involve writing default HTML, JSON, etc. based on the section type. Ensure content is stored correctly (e.g., base64 encoded if that's the standard for the `ContentService`, although `ContentService` deals with raw bytes, the template definition might store it as plain text initially).
    *   If `templateId` is *not* present, execute the existing logic to create a blank design.
4.  **Regenerate Protobuf Code:** After modifying `.proto` files, run the necessary `make proto` or protoc commands to update Go (`./gen/go`) and TypeScript (`./web/views/components/apiclient/`) code.

Start by outlining the necessary Protobuf changes and then modify the Go code in `DesignEditorPage.go` and `services/designs.go`. Define a simple structure for the template definition (e.g., a map or struct in Go). For now, you can hardcode one or two simple template definitions directly in the `CreateDesign` function for testing, before implementing file loading.
```

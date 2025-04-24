
**LeetCoach System Summary**

**1. Overview:**

*   LeetCoach is a web application designed for system design interview preparation.
*   Core functionality revolves around creating and editing multi-section "Design" documents.
*   Supports different section types: Text (with rich editing via TinyMCE), Drawing (using **Excalidraw via React**), and Plot (placeholder).
*   Features user authentication (local email/password, Google, GitHub OAuth).

**2. Architecture & Technologies:**

*   **Backend (Go):**
    *   Uses gRPC for defining services (`DesignService`, `TagService`, potential `AdminService`).
    *   `DesignService` manages designs and their sections. **Crucially, it uses a filesystem-based persistence strategy.** Designs are stored in directories (`<basePath>/<designId>/`), with metadata in `design.json` and each section's content in `sections/<sectionId>.json` (content is often base64 encoded JSON, e.g., for drawings). Concurrency is managed via per-design mutexes.
    *   `TagService` and authentication-related services appear configured to use **Google Cloud Datastore** (via `ClientMgr`).
    *   Provides core logic for CRUD operations on designs and sections, plus section reordering. Expects drawing content as `[]byte` (likely base64 encoded `"excalidraw/json"`).
*   **Web Layer (Go):**
    *   Standard HTTP server (`net/http`).
    *   **gRPC Gateway:** Exposes backend gRPC services as a RESTful JSON API under `/api/v1/`. Handles marshalling between JSON and Protobuf.
    *   **Authentication (`oneauth` + `scs`):** Handles OAuth flows and local logins. Manages user sessions using `scs` (server-side sessions likely stored via default mechanism or potentially Datastore if configured). Injects authenticated user ID into outgoing gRPC metadata for backend context.
    *   **Server-Side Rendering (Go Templates):** Renders initial HTML views (`web/views/`, `web/views/templates/`) using Go's `html/template` engine managed by `tmplr`. Passes initial data (like Design ID, potentially initial list data) to the frontend.
    *   **HTMX:** Used for enhancing some views, notably the design list page (`/`) for sorting, searching, and deleting items directly from the table.
*   **Frontend (TypeScript / React):**
    *   Located in `web/views/components/`.
    *   Largely **Vanilla TypeScript**, but uses **React and ReactDOM** specifically for mounting the **Excalidraw** component within the `DrawingSection`.
    *   **Single-Page Editor Experience (`DesignEditorPage.ts`):** Manages the primary design editing interface.
    *   **Component-Based:** Structured into manageable components (`SectionManager`, `DocumentTitle`, `TableOfContents`, `BaseSection` and its derivatives `TextSection`, `DrawingSection` (integrates Excalidraw), `PlotSection`, `Modal`, `ToastManager`, `ThemeManager`).
    *   **API Client (`apiclient/`, `Api.ts`):** Auto-generated TypeScript client interacts with the gRPC Gateway's REST API. `Api.ts` configures the base path and crucially adds the `Authorization: Bearer ...` header by reading a session cookie (`LeetCoachAuthToken`).
    *   **Client-Side Templates (`TemplateRegistry.html`):** Contains reusable HTML snippets for modals, section view/edit modes (e.g., `text-section-edit`, `drawing-section-edit`), etc., cloned and used by TypeScript components.
    *   **Styling:** Tailwind CSS.

**3. Key Data Flows:**

*   **Loading a Design:** User navigates to `/designs/{id}/edit` -> Go backend renders `DesignEditorPage.html` embedding the Design ID -> `DesignEditorPage.ts` initializes -> Fetches design metadata (`/api/v1/designs/{id}`) -> Fetches content for each section (`/api/v1/designs/{id}/sections/{secId}`) -> `SectionManager` uses fetched data to render sections. `DrawingSection` receives JSON scene data (`{elements, appState}`) and renders a placeholder view.
*   **Saving Changes (Incremental):**
    *   User edits title -> `DocumentTitle.ts` calls `PATCH /api/v1/designs/{id}` with `update_mask=["name"]`.
    *   User edits section title -> `BaseSection` triggers callback -> `SectionManager.ts` calls `PATCH /api/v1/designs/{id}/sections/{secId}` with `update_mask=["title"]`.
    *   User edits text section content -> `TextSection` captures HTML -> `BaseSection` triggers callback -> `SectionManager.ts` calls `PATCH ...` with `update_mask=["text_content"]`.
    *   User edits drawing section content -> `DrawingSection` captures Excalidraw scene data (`{elements, appState}`) -> `BaseSection` triggers callback -> `SectionManager.ts` calls `PATCH ...` with `update_mask=["drawing_content"]`, sending the scene data (base64 encoded JSON).
    *   User adds section -> `SectionManager.ts` calls `POST /api/v1/designs/{id}/sections` with position/relative info -> Renders new section locally based on API response.
    *   User deletes section -> `SectionManager.ts` calls `DELETE /api/v1/designs/{id}/sections/{secId}` -> Removes section locally on success.
    *   User moves section -> `SectionManager.ts` calls `POST /api/v1/designs/{id}/sections/{secId}:move` with position/relative info -> Reorders locally on success.

**4. Current State & Recent Progress:**

*   Backend services for Design and Section CRUD+Move operations are implemented using the filesystem strategy.
*   Frontend `DesignEditorPage` is functional and integrated with the backend API for loading, saving (title, text content), add, delete, move.
*   Frontend `DrawingSection` edit mode now **integrates the Excalidraw component** using React, allowing users to create and edit drawings. Saving drawing content (`drawing_content`) via the API is functional.
*   Authentication flow is working.
*   Basic UI components (TOC, Modals, Toasts, Theme Toggle) are functional.

**5. Known WIP / Next Steps:**

*   **Drawing Sections:** Edit mode uses Excalidraw. View mode currently shows JSON data or a simple placeholder; needs proper rendering (e.g., static SVG export/display). Backend handles base64 JSON, format is `"excalidraw/json"`.
*   **Plot Sections:** Frontend editor needs integration with actual libraries (e.g., Mermaid, Chart.js). Backend saving/loading needs testing with specific plot data formats.
*   **Error Handling/UI Feedback:** Improve user feedback during API calls (loading spinners, more descriptive error toasts). Handle potential inconsistencies if API calls fail partially.
*   **Full Save/Export:** The "Save" and "Export" buttons in the header are currently placeholders. Their full functionality needs implementation.
*   **Persistence Strategy Review:** The mixed filesystem (Designs/Sections) and Datastore (Auth/Tags) approach should be reviewed for long-term suitability and consistency.
*   **Testing:** Backend tests for section operations exist. Frontend testing infrastructure is needed.
*   **Tagging:** Integration of the Tag service with Designs needs implementation.

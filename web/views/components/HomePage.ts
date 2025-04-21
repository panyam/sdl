// components/HomePage.ts

import { ThemeManager } from './ThemeManager'; // For theme consistency if needed later
import { Modal } from './Modal';
import { ToastManager } from './ToastManager';

// Define the structure for a single document item in the list
interface DocumentListItem {
    id: string; // Unique ID for the document (e.g., 'doc-123')
    title: string;
    lastModified: Date; // Use Date object for easier sorting/formatting
}

/**
 * Manages the listing page logic
 */
class HomePage {
    private modal: Modal;
    private toastManager: ToastManager;
    private documents: DocumentListItem[] = []; // To store the fetched documents

    constructor() {
        ThemeManager.init(); // Initialize theme handling
        this.modal = Modal.getInstance();
        this.toastManager = ToastManager.getInstance();

        console.log("HomePage initialized");

        this.fetchDocuments(); // Fetch initial data
    }

    /**
     * Simulates fetching document list data.
     * In a real app, this would make an API call.
     */
    private fetchDocuments(): void {
        console.log("Fetching documents...");
        // Simulate API delay (optional)
        setTimeout(() => {
            // Mock Data
            this.documents = [
                { id: "doc-abc", title: "Design Twitter", lastModified: new Date(Date.now() - 2 * 60 * 60 * 1000) }, // 2 hours ago
                { id: "doc-def", title: "Design YouTube", lastModified: new Date(Date.now() - 25 * 60 * 60 * 1000) }, // Yesterday
                { id: "doc-ghi", title: "Design TinyURL", lastModified: new Date(Date.now() - 5 * 24 * 60 * 60 * 1000) }, // 5 days ago
                { id: "doc-jkl", title: "Design Netflix", lastModified: new Date(Date.now() - 30 * 60 * 1000) }, // 30 mins ago
            ];
            console.log("Mock documents fetched:", this.documents);
            // Next step will be to call a render method here
            // this.renderDocumentList();
        }, 50); // Simulate 50ms delay
    }

    // --- Methods for rendering, sorting, searching, actions will go here ---


    /**
     * Static initializer
     */
    public static init(): HomePage {
        return new HomePage();
    }
}

// Initialize the HomePage when the DOM is fully loaded
document.addEventListener('DOMContentLoaded', () => {
    HomePage.init();
});

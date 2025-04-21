// components/ToastManager.ts

/**
 * Toast types for styling
 */
type ToastType = 'success' | 'error' | 'info' | 'warning';

/**
 * Manages toast notifications
 */
export class ToastManager {
    private static instance: ToastManager | null = null;
    private container: HTMLElement | null;
    private template: HTMLElement | null;
    private toasts: Map<string, HTMLElement> = new Map();
    private counter: number = 0;

    /**
     * Private constructor for singleton pattern
     */
    private constructor() {
        // this.container = document.querySelector('.fixed.bottom-4.left-4.z-50');
        // Select the container using its ID from index.html
        this.container = document.getElementById('toast-container');
        this.template = document.getElementById('toast-template');
    }

    /**
     * Get the ToastManager instance (singleton)
     */
    public static getInstance(): ToastManager {
        if (!ToastManager.instance) {
            ToastManager.instance = new ToastManager();
        }
        return ToastManager.instance;
    }

    /**
     * Show a toast notification
     * @param title Toast title
     * @param message Toast message
     * @param type Toast type for styling
     * @param duration Duration in ms (default: 4000)
     */
    public showToast(title: string, message: string, type: ToastType = 'info', duration: number = 4000): string {
        if (!this.container || !this.template) return '';

        // Create a unique ID for this toast
        const id = `toast-${Date.now()}-${this.counter++}`;

        // Clone the template
        const toast = this.template.cloneNode(true) as HTMLElement;
        toast.id = id;
        toast.classList.remove('hidden');

        // Set content
        const titleElement = toast.querySelector('.toast-title');
        const messageElement = toast.querySelector('.toast-message');
        if (titleElement) titleElement.textContent = title;
        if (messageElement) messageElement.textContent = message;

        // Set style based on type
        const iconContainer = toast.querySelector('.flex-shrink-0');
        if (iconContainer) {
            // Clear existing icon
            iconContainer.innerHTML = '';

            // Add appropriate icon based on type
            let icon: string;
            let borderColor: string;
            
            switch (type) {
                case 'success':
                    icon = '<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-green-500" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" /></svg>';
                    borderColor = 'border-green-500';
                    break;
                case 'error':
                    icon = '<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-red-500" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clip-rule="evenodd" /></svg>';
                    borderColor = 'border-red-500';
                    break;
                case 'warning':
                    icon = '<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-yellow-500" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clip-rule="evenodd" /></svg>';
                    borderColor = 'border-yellow-500';
                    break;
                case 'info':
                default:
                    icon = '<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-blue-500" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clip-rule="evenodd" /></svg>';
                    borderColor = 'border-blue-500';
                    break;
            }

            // Add icon
            iconContainer.innerHTML = icon;

            // Set border color
            const borderElement = toast.querySelector('.border-l-4');
            if (borderElement) {
                borderElement.className = borderElement.className.replace(/border-[a-z]+-500/g, borderColor);
            }
        }

        // Add close button handler
        const closeButton = toast.querySelector('.toast-close');
        if (closeButton) {
            closeButton.addEventListener('click', () => {
                this.hideToast(id);
            });
        }

        // Add to container
        this.container.appendChild(toast);
        this.toasts.set(id, toast);

        // Show toast with animation
        setTimeout(() => {
            toast.classList.remove('scale-95', 'opacity-0');
            toast.classList.add('scale-100', 'opacity-100');
        }, 10);

        // Auto-hide after duration
        if (duration > 0) {
            setTimeout(() => {
                this.hideToast(id);
            }, duration);
        }

        return id;
    }

    /**
     * Hide a toast notification
     * @param id Toast ID
     */
    public hideToast(id: string): void {
        const toast = this.toasts.get(id);
        if (!toast) return;

        // Hide with animation
        toast.classList.remove('scale-100', 'opacity-100');
        toast.classList.add('scale-95', 'opacity-0');

        // Remove after animation
        setTimeout(() => {
            toast.remove();
            this.toasts.delete(id);
        }, 300);
    }

    /**
     * Hide all toast notifications
     */
    public hideAllToasts(): void {
        this.toasts.forEach((_, id) => {
            this.hideToast(id);
        });
    }

    /**
     * Initialize the component
     */
    public static init(): ToastManager {
        return ToastManager.getInstance();
    }
}

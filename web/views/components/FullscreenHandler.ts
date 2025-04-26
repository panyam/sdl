// FILE: ./web/views/components/FullscreenHandler.ts

interface FullscreenHandlerElements {
    sectionElement: HTMLElement;
    contentContainer: HTMLElement | null;
    fullscreenButton: HTMLElement | null;
    exitFullscreenButton: HTMLElement | null;
    // Toolbar buttons to hide/show
    moveUpButton?: HTMLElement | null;
    moveDownButton?: HTMLElement | null;
    addBeforeButton?: HTMLElement | null;
    addAfterButton?: HTMLElement | null;
    settingsButton?: HTMLElement | null;
    deleteButton?: HTMLElement | null;
}

type ResizeCallback = (isEntering: boolean) => void;

export class FullscreenHandler {
    private elements: FullscreenHandlerElements;
    private resizeCallback: ResizeCallback;
    private isFullscreen: boolean = false;

    // Bound event listeners for easy removal
    private boundHandleKeyDown = this.handleKeyDown.bind(this);
    private boundHandleResize = this.handleResize.bind(this);
    private boundEnter = this.enter.bind(this);
    private boundExit = this.exit.bind(this);


    constructor(elements: FullscreenHandlerElements, resizeCallback: ResizeCallback) {
        this.elements = elements;
        this.resizeCallback = resizeCallback;
        this.bindEnterButton();
    }

    private bindEnterButton(): void {
        if (this.elements.fullscreenButton) {
            // Clear any previous listener before adding
            this.elements.fullscreenButton.removeEventListener('click', this.boundEnter);
            this.elements.fullscreenButton.addEventListener('click', this.boundEnter);
        }
    }

    public enter(): void {
        if (this.isFullscreen || !this.elements.sectionElement || !this.elements.exitFullscreenButton) return;
        console.log(`Entering fullscreen for section`); // Removed section ID for generic handler
        this.isFullscreen = true;

        // Apply CSS classes
        this.elements.sectionElement.classList.add('lc-section-fullscreen', 'flex', 'flex-col');
        this.elements.sectionElement.classList.remove('mb-6');
        this.elements.contentContainer?.classList.add('flex-grow', 'overflow-auto');
        document.body.classList.add('lc-fullscreen-active');

        // Show/Hide buttons
        this.toggleToolbarButtons(false); // Hide toolbar buttons
        this.elements.exitFullscreenButton?.classList.remove('hidden');

        // Bind exit triggers
        this.elements.exitFullscreenButton.onclick = this.boundExit; // Use onclick for simplicity on exit button
        window.addEventListener('keydown', this.boundHandleKeyDown);
        window.addEventListener('resize', this.boundHandleResize);

        // Trigger resize callback
        requestAnimationFrame(() => {
            this.resizeCallback(true);
        });
    }

    public exit(): void {
        if (!this.isFullscreen || !this.elements.sectionElement || !this.elements.exitFullscreenButton) return;
        console.log(`Exiting fullscreen for section`);
        this.isFullscreen = false;

        // Remove CSS classes
        this.elements.sectionElement.classList.remove('lc-section-fullscreen', 'flex', 'flex-col');
        this.elements.sectionElement.classList.add('mb-6');
        this.elements.contentContainer?.classList.remove('flex-grow', 'overflow-auto');
        document.body.classList.remove('lc-fullscreen-active');

        // Show/Hide buttons
        this.elements.exitFullscreenButton?.classList.add('hidden');
        this.toggleToolbarButtons(true); // Show toolbar buttons

        // Unbind exit triggers
        this.elements.exitFullscreenButton.onclick = null; // Remove onclick
        window.removeEventListener('keydown', this.boundHandleKeyDown);
        window.removeEventListener('resize', this.boundHandleResize);

        // Trigger resize callback
        requestAnimationFrame(() => {
            this.resizeCallback(false);
        });
    }

    private toggleToolbarButtons(show: boolean): void {
        const buttons = [
            this.elements.moveUpButton,
            this.elements.moveDownButton,
            this.elements.addBeforeButton,
            this.elements.addAfterButton,
            this.elements.settingsButton,
            this.elements.deleteButton,
        ];
        buttons.forEach(btn => btn?.classList.toggle('hidden', !show));
    }

    private handleKeyDown(event: KeyboardEvent): void {
        if (this.isFullscreen && event.key === 'Escape') {
            this.exit();
        }
    }

    private handleResize(): void {
        if (this.isFullscreen) {
            this.resizeCallback(true); // Notify section on resize while fullscreen
        }
    }

    /**
     * Removes event listeners to prevent memory leaks. Call when the section is destroyed.
     */
    public destroy(): void {
        console.log("Destroying FullscreenHandler listeners");
        if (this.elements.fullscreenButton) {
            this.elements.fullscreenButton.removeEventListener('click', this.boundEnter);
        }
         if (this.elements.exitFullscreenButton) {
            this.elements.exitFullscreenButton.onclick = null;
         }
        window.removeEventListener('keydown', this.boundHandleKeyDown);
        window.removeEventListener('resize', this.boundHandleResize);
    }
}

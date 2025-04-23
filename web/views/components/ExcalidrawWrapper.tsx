// ./web/views/components/ExcalidrawWrapper.tsx
import React, { useEffect, useRef, useState, useCallback } from 'react';
import { Excalidraw, exportToBlob, getSceneVersion } from '@excalidraw/excalidraw';
import type { ExcalidrawElement, AppState, BinaryFiles, ExcalidrawAPIRefValue } from '@excalidraw/excalidraw/types/types'; // Correct import path

// Define the props for the wrapper
interface ExcalidrawWrapperProps {
    initialData?: { // Match the structure Excalidraw expects for initial data
        elements?: readonly ExcalidrawElement[];
        appState?: Partial<AppState>;
        files?: BinaryFiles;
    };
    onChange: (elements: readonly ExcalidrawElement[], appState: AppState, files: BinaryFiles) => void;
    theme?: 'light' | 'dark'; // Optional theme prop
}

export const ExcalidrawWrapper: React.FC<ExcalidrawWrapperProps> = ({ initialData, onChange, theme }) => {
    const excalidrawApiRef = useRef<ExcalidrawAPIRefValue | null>(null); // Ref to access Excalidraw API

    // Use state to manage the Excalidraw component itself to potentially force re-renders if needed
    // Although often just passing props like theme is sufficient.
    const [ExcalidrawComponent, setExcalidrawComponent] = useState<any>(null);

    useEffect(() => {
        // Dynamically import Excalidraw component only on the client-side
        import('@excalidraw/excalidraw').then(module => setExcalidrawComponent(() => module.Excalidraw));
    }, []);


    // Use useCallback for the onChange handler to prevent unnecessary re-renders of Excalidraw
    const handleExcalidrawChange = useCallback(
        (elements: readonly ExcalidrawElement[], appState: AppState, files: BinaryFiles) => {
            // We only care about the scene version for triggering onChange,
            // as comparing potentially large elements/files arrays is expensive.
             if (onChange) {
                onChange(elements, appState, files);
             }
        },
        [onChange] // Dependency array includes the onChange callback
    );

    // --- Optional: Add functions to interact with Excalidraw API via ref if needed ---
    // Example: Exporting image data (could be triggered by external save button)
    const exportImage = async (type = 'png') => {
        if (!excalidrawApiRef.current?.getSceneElements || !excalidrawApiRef.current?.getAppState) {
            console.error("Excalidraw API not ready for export");
            return null;
        }
        const elements = excalidrawApiRef.current.getSceneElements();
        const appState = excalidrawApiRef.current.getAppState();
        // files might be needed depending on export type and content
        // const files = excalidrawApiRef.current.getFiles();

        try {
            const blob = await exportToBlob({
                elements,
                appState,
                // files, // Pass files if necessary
                mimeType: type === 'png' ? 'image/png' : 'image/svg+xml',
                // ... other export options
            });
            return blob;
        } catch (error) {
            console.error('Error exporting Excalidraw image:', error);
            return null;
        }
    };
    // --------------------------------------------------------------------------

    if (!ExcalidrawComponent) {
         return <div>Loading Excalidraw...</div>; // Or some placeholder
    }

    return (
        <div style={{ height: '100%', width: '100%' }} className="excalidraw-wrapper">
            <Excalidraw
                ref={excalidrawApiRef} // Assign ref
                initialData={initialData}
                onChange={handleExcalidrawChange}
                theme={theme || 'light'} // Use passed theme or default
                // You can add more Excalidraw props here as needed
                // Example: UI options
                // UIOptions={{ canvasActions: { loadScene: false } }}
            />
        </div>
    );
};

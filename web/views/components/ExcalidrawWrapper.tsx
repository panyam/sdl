// FILE: ./web/views/components/ExcalidrawWrapper.tsx
import React, { useEffect, useRef, useState, useCallback } from 'react';
// Try importing types directly from the main package or potentially a dedicated types export
import { Excalidraw, exportToBlob, getSceneVersion, convertToExcalidrawElements } from '@excalidraw/excalidraw';
// Adjust type imports based on actual package structure - check node_modules/@excalidraw/excalidraw/types if needed
// Import types from the specific 'types' path
import type {
    AppState,
    BinaryFiles,
    ExcalidrawImperativeAPI // This specific type name might vary, check node_modules/@excalidraw/excalidraw/types/types.d.ts or similar if it still fails
} from '@excalidraw/excalidraw/types/types'; // <-- Changed import path

type ExcalidrawElement = any;

// Define the props for the wrapper
interface ExcalidrawWrapperProps {
    initialData?: { // Match the structure Excalidraw expects for initial data
        elements?: readonly ExcalidrawElement[]; // Use the imported type
        appState?: Partial<AppState>; // Use the imported type
        files?: BinaryFiles; // Use the imported type
    };
    onChange: (elements: readonly ExcalidrawElement[], appState: AppState, files: BinaryFiles) => void;
    theme?: 'light' | 'dark'; // Optional theme prop
    // Add the excalidrawAPI prop callback type definition
    excalidrawAPI?: (api: ExcalidrawImperativeAPI) => void;
}

export const ExcalidrawWrapper: React.FC<ExcalidrawWrapperProps> = ({
    initialData,
    onChange,
    theme,
    excalidrawAPI // Receive the callback prop
 }) => {
    // Ref now holds the correct type ExcalidrawImperativeAPI or null
    const excalidrawApiRef = useRef<ExcalidrawImperativeAPI | null>(null);

    const [ExcalidrawComponent, setExcalidrawComponent] = useState<any>(null);

    useEffect(() => {
        // Dynamic import remains the same
        import('@excalidraw/excalidraw').then(module => setExcalidrawComponent(() => module.Excalidraw));
    }, []);

    const handleExcalidrawChange = useCallback(
        (elements: readonly ExcalidrawElement[], appState: AppState, files: BinaryFiles) => {
            if (onChange) {
                onChange(elements, appState, files);
             }
        },
        [onChange]
    );

    // --- Export function ---
    const exportImage = async (type = 'png') => {
        // Check if the API ref is current and has the necessary methods
        if (!excalidrawApiRef.current?.getSceneElements || !excalidrawApiRef.current?.getAppState || !excalidrawApiRef.current?.getFiles) {
            console.error("Excalidraw API not ready for export or missing methods");
            return null;
        }
        const elements = excalidrawApiRef.current.getSceneElements();
        const appState = excalidrawApiRef.current.getAppState();
        const files = excalidrawApiRef.current.getFiles(); // Get files object

        try {
            const blob = await exportToBlob({
                elements,
                appState,
                files, // <-- Provide the files object
                mimeType: type === 'png' ? 'image/png' : 'image/svg+xml',
                // ... other export options
            });
            return blob;
        } catch (error) {
            console.error('Error exporting Excalidraw image:', error);
            return null;
        }
    };
    // ---------------------


    if (!ExcalidrawComponent) {
         return <div className="p-4 text-center text-gray-500 dark:text-gray-400">Loading Drawing Editor...</div>; // Placeholder
    }

    return (
        <div style={{ height: '100%', width: '100%' }} className="excalidraw-wrapper">
            {/* Pass the excalidrawAPI prop correctly */}
            <ExcalidrawComponent
                excalidrawAPI={(api: ExcalidrawImperativeAPI) => {
                    excalidrawApiRef.current = api; // Store the API handle in the ref
                    if (excalidrawAPI) {
                        excalidrawAPI(api); // Call the passed callback prop
                    }
                }}
                initialData={initialData}
                onChange={handleExcalidrawChange}
                theme={theme || 'light'}
                // REMOVE the direct ref prop: ref={excalidrawApiRef}
            />
        </div>
    );
};

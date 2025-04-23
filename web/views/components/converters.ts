// ./web/views/components/converters.ts

import { SectionData, SectionType, TextContent, DrawingContent, PlotContent } from './types';
import {
    V1Section, V1SectionType,
    V1TextSectionContent, V1DrawingSectionContent, V1PlotSectionContent,
    // Renaming the long type for clarity
    SectionObjectContainingOnlyTheFieldsToBeUpdatedTheServerWillUseTheUpdateMaskToKnowWhichFieldsFromThisSectionMessageToApplyToTheStoredSection as ApiSectionUpdateObject
} from './apiclient'; // Adjust path as needed

// --- Type Mapping ---

export function mapApiSectionTypeToFrontend(apiType?: V1SectionType): SectionType {
    switch (apiType) {
        case V1SectionType.SectionTypeText: return 'text';
        case V1SectionType.SectionTypeDrawing: return 'drawing';
        case V1SectionType.SectionTypePlot: return 'plot';
        // Default to text or throw an error if unspecified is invalid
        case V1SectionType.SectionTypeUnspecified:
        default:
             console.warn(`Received unspecified or unknown section type: ${apiType}, defaulting to 'text'.`);
             return 'text';
    }
}

export function mapFrontendSectionTypeToApi(type: SectionType): V1SectionType {
    switch (type) {
        case 'text': return V1SectionType.SectionTypeText;
        case 'drawing': return V1SectionType.SectionTypeDrawing;
        case 'plot': return V1SectionType.SectionTypePlot;
        default:
             console.error(`Cannot map unknown frontend section type: ${type}`);
             return V1SectionType.SectionTypeUnspecified; // Indicate error
    }
}

// --- Content Mapping ---

// Export this function
export function extractContentFromApiSection(apiSection: V1Section): SectionData['content'] {
    const type = mapApiSectionTypeToFrontend(apiSection.type); // Determine type first

    // Use the determined type to check the correct content field
    if (type === 'text' && apiSection.textContent) {
        return apiSection.textContent.htmlContent || '';
    }
    if (type === 'drawing' && apiSection.drawingContent) {
        // Assuming drawing data is stored as a stringified JSON
        try {
            return {
                format: apiSection.format || 'placeholder_drawing_json', // Use format if available
                data: JSON.parse(apiSection.drawingContent.data || '{}')
            } as DrawingContent;
        } catch (e) {
            console.error(`Failed to parse drawing content data for section ${apiSection.id}:`, e, "Data:", apiSection.drawingContent.data);
            return { format: 'placeholder_drawing_json', data: {} } as DrawingContent; // Default fallback
        }
    }
    if (type === 'plot' && apiSection.plotContent) {
         // Assuming plot data is stored as a stringified JSON
        try {
            return {
                format: apiSection.format || 'placeholder_plot_json', // Use format if available
                data: JSON.parse(apiSection.plotContent.data || '{}')
            } as PlotContent;
        } catch (e) {
            console.error(`Failed to parse plot content data for section ${apiSection.id}:`, e, "Data:", apiSection.plotContent.data);
            return { format: 'placeholder_plot_json', data: {} } as PlotContent; // Default fallback
        }
    }
    // Fallback for unknown or missing content appropriate for the type
     console.warn(`Content field missing or type mismatch for section ${apiSection.id}, type ${apiSection.type}. Returning default content.`);
     switch(type) { // Return appropriate default based on determined type
        case 'drawing': return { format: 'placeholder_drawing', data: {} } as DrawingContent;
        case 'plot': return { format: 'placeholder_plot', data: {} } as PlotContent;
        case 'text':
        default: return '';
     }
}


export function mapFrontendContentToApiUpdate(
    type: SectionType,
    content: SectionData['content']
): Partial<ApiSectionUpdateObject> {
    const update: Partial<ApiSectionUpdateObject> = {};
    // Always set the type in the update object based on the frontend type
    update.type = mapFrontendSectionTypeToApi(type);

    switch (type) {
        case 'text':
            // Ensure content is a string for text type
            update.textContent = { htmlContent: typeof content === 'string' ? content : '' };
            // update.contentType = 'text/html'; // Optional: set standard content type
            break;
        case 'drawing':
            const drawingContent = content as DrawingContent;
            // Ensure data is stringified
            update.drawingContent = { data: JSON.stringify(drawingContent?.data ?? {}) };
            update.format = drawingContent?.format || 'placeholder_drawing_json'; // Pass format back, ensure default
            // update.contentType = 'application/json'; // Or specific format MIME type
            break;
        case 'plot':
            const plotContent = content as PlotContent;
             // Ensure data is stringified
            update.plotContent = { data: JSON.stringify(plotContent?.data ?? {}) };
             update.format = plotContent?.format || 'placeholder_plot_json'; // Pass format back, ensure default
           // update.contentType = 'application/json'; // Or specific format MIME type
            break;
        default:
             console.error(`Unknown section type encountered in mapFrontendContentToApiUpdate: ${type}`);
             // Avoid sending invalid content
             break;
    }
    return update;
}


// --- Full Conversion ---

export function convertApiSectionToSectionData(apiSection: V1Section): SectionData {
    if (!apiSection?.id) {
        throw new Error("Cannot convert API section without an ID");
    }
    const frontendType = mapApiSectionTypeToFrontend(apiSection.type);
    return {
        id: apiSection.id,
        designId: apiSection.designId || "",
        type: frontendType,
        title: apiSection.title || `Untitled ${frontendType} Section`, // Provide default title
        // Use || 0 safely as order is number | undefined
        order: apiSection.order || 0, // Use order from API, default to 0 if undefined
        content: extractContentFromApiSection(apiSection),
    };
}

export function createApiSectionUpdateObject(
    updates: { title?: string; content?: SectionData['content'] },
    currentType: SectionType
): ApiSectionUpdateObject {
     const apiUpdate: Partial<ApiSectionUpdateObject> = {}; // Use Partial for easier construction
     if (updates.title !== undefined) {
         apiUpdate.title = updates.title;
     }
     if (updates.content !== undefined) {
         const contentUpdate = mapFrontendContentToApiUpdate(currentType, updates.content);
         Object.assign(apiUpdate, contentUpdate);
     }
     // The returned object needs to conform to ApiSectionUpdateObject,
     // even if fields are optional in Partial during construction.
     // It's assumed the backend handles the partial update based on presence.
     return apiUpdate as ApiSectionUpdateObject;
}

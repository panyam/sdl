// ./web/views/components/converters.ts

import { SectionData, SectionType, ExcalidrawSceneData } from './types';
import {
    V1Section, V1SectionType,
    // Renaming the long type for clarity
    ConsolidateSectionUpdatesIntoOneRPCUsingPATCHAndFieldMask as ApiSectionUpdateObject
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

/**
 * Converts a V1Section proto (metadata only) to the frontend SectionData structure.
 * Initializes content to null, as it needs to be fetched separately.
 */
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
        getAnswerPrompt: apiSection.getAnswerPrompt || "",
        verifyAnswerPrompt: apiSection.verifyAnswerPrompt || "",
    };
}

export function createApiSectionUpdateObject(
    updates: { title?: string; },
    currentType: SectionType
): ApiSectionUpdateObject {
     const apiUpdate: Partial<ApiSectionUpdateObject> = {}; // Use Partial for easier construction
     if (updates.title !== undefined) {
        apiUpdate.section = apiUpdate.section || {}
        apiUpdate.section.title = updates.title;
     }
     // The returned object needs to conform to ApiSectionUpdateObject,
     // even if fields are optional in Partial during construction.
     // It's assumed the backend handles the partial update based on presence.
     return apiUpdate as ApiSectionUpdateObject;
}

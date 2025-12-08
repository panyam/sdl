import { createConnectTransport } from '@connectrpc/connect-web';
import { createClient } from '@connectrpc/connect';
import { create } from '@bufbuild/protobuf';
import {
  SystemsService,
} from '../../gen/sdl/v1/services/systems_pb';
import {
  ListSystemsRequestSchema,
  GetSystemRequestSchema,
  GetSystemContentRequestSchema,
  type SystemInfo,
  type SystemProject
} from '../../gen/sdl/v1/models/systems_pb';

/**
 * SystemsServiceClient provides access to the Systems gRPC API
 */
export class SystemsServiceClient {
  private client: ReturnType<typeof createClient<typeof SystemsService>>;
  
  constructor(baseUrl?: string) {
    const transport = createConnectTransport({
      baseUrl: baseUrl || window.location.origin + '/api'
    });
    
    this.client = createClient(SystemsService, transport);
  }
  
  /**
   * List all available systems
   */
  async listSystems(): Promise<SystemInfo[]> {
    const request = create(ListSystemsRequestSchema);
    const response = await this.client.listSystems(request);
    return response.systems;
  }
  
  /**
   * Get a specific system with all metadata
   */
  async getSystem(id: string, version?: string): Promise<SystemProject> {
    const request = create(GetSystemRequestSchema, {
      id,
      version: version || ''
    });
    
    const response = await this.client.getSystem(request);
    if (!response.system) {
      throw new Error(`System ${id} not found`);
    }
    
    return response.system;
  }
  
  /**
   * Get SDL and recipe content for a system
   */
  async getSystemContent(id: string, version?: string): Promise<{
    sdlContent: string;
    recipeContent: string;
    readmeContent: string;
  }> {
    const request = create(GetSystemContentRequestSchema, {
      id,
      version: version || ''
    });
    
    const response = await this.client.getSystemContent(request);
    return {
      sdlContent: response.sdlContent,
      recipeContent: response.recipeContent,
      readmeContent: response.readmeContent
    };
  }
}

// Export a default instance
export const systemsService = new SystemsServiceClient();

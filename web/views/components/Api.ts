
// Example in DesignEditorPage.ts or a dedicated api service wrapper
import { Configuration, DesignServiceApi, ContentServiceApi, LlmServiceApi /*, other APIs */ } from './apiclient'; // <-- Add ContentServiceApi

function getCookie(cname: string) : string {
  let name = cname + "=";
  let decodedCookie = decodeURIComponent(document.cookie);
  let ca = decodedCookie.split(';');
  for(let i = 0; i <ca.length; i++) {
    let c = ca[i];
    while (c.charAt(0) == ' ') {
      c = c.substring(1);
    }
    if (c.indexOf(name) == 0) {
      return c.substring(name.length, c.length);
    }
  }
  return "";
}

// Function to get your token (implement this based on your auth flow)
function getAuthToken(): string | null | Promise<string | null> {
    return getCookie("LeetCoachAuthToken")
}

const apiConfig = new Configuration({
    basePath: '/api', // Your gRPC Gateway base path
    // basePath: '/api/v1', // Ensure this points to your gRPC Gateway prefix (usually /api/v1)
    // --- Authentication ---
    fetchApi: async (input: RequestInfo | URL, init?: RequestInit): Promise<Response> => {
        // Define a wrapper around fetch to inject the token
        const token = await getAuthToken();
        const headers = new Headers(init?.headers);
        if (token) {
            headers.set('Authorization', `Bearer ${token}`);
        }
        // Ensure Content-Type is set for relevant methods if not already present
        // if (init?.method === 'POST' || init?.method === 'PUT' || init?.method === 'PATCH') {
        //     if (!headers.has('Content-Type')) {
        //         headers.set('Content-Type', 'application/json'); // Default, adjust if needed
        //     }
        // }
        const modifiedInit = { ...init, headers };
        return fetch(input, modifiedInit);
    },

    // Note: Using fetchApi interceptor is generally preferred over accessToken function
    // with openapi-generator v6+ for more robust header injection.
    // accessToken: async () => {
    //    // Use a function for dynamic token retrieval
    //    const token = await getAuthToken();
    //    return token ? `Bearer ${token}` : undefined; // Return undefined if no token
    // },
});

// Instantiate your API clients
export const DesignApi = new DesignServiceApi(apiConfig);
export const ContentApi = new ContentServiceApi(apiConfig); // <-- Export ContentServiceApi instance
export const LlmApi = new LlmServiceApi(apiConfig); // <-- Export ContentServiceApi instance


// Example in DesignEditorPage.ts or a dedicated api service wrapper
import { Configuration, DesignServiceApi /*, other APIs */ } from './apiclient';

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
// This might involve reading from session storage, a cookie handled by the server,
// or calling an endpoint that returns the current user's token.
function getAuthToken(): string | null | Promise<string | null> {
    // Placeholder: Retrieve token logic
    // Example: Read from a meta tag set by the server
    // const meta = document.querySelector<HTMLMetaElement>('meta[name="csrf-token"]'); // Or a dedicated auth token meta tag
    // return meta ? meta.content : null;
    // Example: Read from session storage (if applicable)
    return getCookie("LeetCoachAuthToken")
     // return sessionStorage.getItem('authToken');
    // Example: Return null if not logged in
}

const apiConfig = new Configuration({
    basePath: '/api', // Your gRPC Gateway base path
    // --- Authentication ---
    accessToken: async () => {
        // Use a function for dynamic token retrieval
        const token = await getAuthToken();
        return `Bearer ${token}`
    },
    // OR for API Key (if you use that instead/additionally)
    // apiKey: async () => {
    //    const key = await getApiKey();
    //    return key ? `YOUR_API_KEY_PREFIX ${key}` : undefined; // Adjust prefix if needed
    // },
    // --- Other Config ---
    // middleware: [...] // Add fetch middleware if needed (e.g., for complex logging/error handling)
});

// Instantiate your API clients
export const DesignApi = new DesignServiceApi(apiConfig);

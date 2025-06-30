// FileSystem client interface and implementations

export interface FileSystemClient {
  id: string;
  name: string;
  icon: string;
  isReadOnly: boolean;
  
  listFiles(path: string): Promise<string[]>;
  readFile(path: string): Promise<string>;
  writeFile(path: string, content: string): Promise<void>;
  deleteFile(path: string): Promise<void>;
  createDirectory(path: string): Promise<void>;
}

// Response types for server filesystem API
interface FileInfo {
  name: string;
  path: string;
  isDirectory: boolean;
  size?: number;
  modTime?: string;
}

interface ListFilesResponse {
  files: FileInfo[];
}

// Local filesystem client - connects to server-hosted filesystems
export class LocalFileSystemClient implements FileSystemClient {
  constructor(
    private serverPath: string,  // e.g., "http://localhost:8080/api/filesystems"
    public id: string,           // e.g., "examples"
    public name: string,         // e.g., "Examples"
    public isReadOnly: boolean,
    public icon: string
  ) {}

  async listFiles(path: string): Promise<string[]> {
    try {
      // Normalize path - ensure it starts with /
      const normalizedPath = path.startsWith('/') ? path : `/${path}`;
      const url = `${this.serverPath}/${this.id}${normalizedPath}`;
      
      const response = await fetch(url);
      if (!response.ok) {
        throw new Error(`Failed to list files: ${response.statusText}`);
      }
      
      // Check content type
      const contentType = response.headers.get('content-type');
      if (contentType?.includes('application/json')) {
        // Server returns JSON
        const data: ListFilesResponse = await response.json();
        return data.files.map(f => {
          // Return full path with directory indicator
          const fullPath = normalizedPath === '/' 
            ? `/${f.name}` 
            : `${normalizedPath}/${f.name}`.replace(/\/+/g, '/');
          return f.isDirectory ? `${fullPath}/` : fullPath;
        });
      } else {
        // Fallback: parse HTML directory listing
        const html = await response.text();
        return this.parseHtmlListing(html, normalizedPath);
      }
    } catch (error) {
      console.error(`Failed to list files at ${path}:`, error);
      return [];
    }
  }

  private parseHtmlListing(html: string, basePath: string): string[] {
    const parser = new DOMParser();
    const doc = parser.parseFromString(html, 'text/html');
    const links = doc.querySelectorAll('a');
    
    const files: string[] = [];
    links.forEach(link => {
      const href = link.getAttribute('href');
      if (href && href !== '../') {
        // Convert relative paths to absolute paths
        const fullPath = basePath === '/' 
          ? `/${href}` 
          : `${basePath}/${href}`.replace(/\/+/g, '/');
        files.push(fullPath);
      }
    });
    
    return files;
  }

  async readFile(path: string): Promise<string> {
    try {
      const normalizedPath = path.startsWith('/') ? path : `/${path}`;
      const url = `${this.serverPath}/${this.id}${normalizedPath}`;
      
      const response = await fetch(url);
      if (!response.ok) {
        throw new Error(`Failed to read file: ${response.statusText}`);
      }
      
      return await response.text();
    } catch (error) {
      console.error(`Failed to read file ${path}:`, error);
      throw error;
    }
  }

  async writeFile(path: string, content: string): Promise<void> {
    if (this.isReadOnly) {
      throw new Error(`Filesystem '${this.name}' is read-only`);
    }
    
    try {
      const normalizedPath = path.startsWith('/') ? path : `/${path}`;
      const url = `${this.serverPath}/${this.id}${normalizedPath}`;
      
      const response = await fetch(url, {
        method: 'PUT',
        headers: {
          'Content-Type': 'text/plain',
        },
        body: content
      });
      
      if (!response.ok) {
        throw new Error(`Failed to write file: ${response.statusText}`);
      }
    } catch (error) {
      console.error(`Failed to write file ${path}:`, error);
      throw error;
    }
  }

  async deleteFile(path: string): Promise<void> {
    if (this.isReadOnly) {
      throw new Error(`Filesystem '${this.name}' is read-only`);
    }
    
    try {
      const normalizedPath = path.startsWith('/') ? path : `/${path}`;
      const url = `${this.serverPath}/${this.id}${normalizedPath}`;
      
      const response = await fetch(url, {
        method: 'DELETE'
      });
      
      if (!response.ok) {
        throw new Error(`Failed to delete file: ${response.statusText}`);
      }
    } catch (error) {
      console.error(`Failed to delete file ${path}:`, error);
      throw error;
    }
  }

  async createDirectory(path: string): Promise<void> {
    if (this.isReadOnly) {
      throw new Error(`Filesystem '${this.name}' is read-only`);
    }
    
    try {
      const normalizedPath = path.startsWith('/') ? path : `/${path}`;
      const url = `${this.serverPath}/${this.id}${normalizedPath}`;
      
      const response = await fetch(url, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ type: 'directory' })
      });
      
      if (!response.ok) {
        throw new Error(`Failed to create directory: ${response.statusText}`);
      }
    } catch (error) {
      console.error(`Failed to create directory ${path}:`, error);
      throw error;
    }
  }
}

// GitHub filesystem client - read-only access to GitHub repositories
export class GitHubFileSystemClient implements FileSystemClient {
  public readonly isReadOnly = true;
  private apiBase = 'https://api.github.com';
  private rawBase = 'https://raw.githubusercontent.com';
  
  constructor(
    public id: string,          // e.g., "github-examples"
    public name: string,        // e.g., "GitHub Examples"
    private repoOwner: string,  // e.g., "panyam"
    private repoName: string,   // e.g., "sdl"
    private branch: string,     // e.g., "main"
    private basePath: string,   // e.g., "/examples"
    public icon: string
  ) {}

  async listFiles(path: string): Promise<string[]> {
    try {
      // Construct GitHub API path
      const normalizedPath = path.startsWith('/') ? path : `/${path}`;
      const repoPath = `${this.basePath}${normalizedPath}`.replace(/\/+/g, '/').replace(/\/$/, '');
      const apiUrl = `${this.apiBase}/repos/${this.repoOwner}/${this.repoName}/contents${repoPath}?ref=${this.branch}`;
      
      const response = await fetch(apiUrl, {
        headers: {
          'Accept': 'application/vnd.github.v3+json'
        }
      });
      
      if (response.status === 404) {
        return [];
      }
      
      if (!response.ok) {
        throw new Error(`GitHub API error: ${response.statusText}`);
      }
      
      const items = await response.json();
      
      // Handle single file response
      if (!Array.isArray(items)) {
        return [];
      }
      
      return items.map((item: any) => {
        const itemPath = item.path.replace(this.basePath, '').replace(/^\//, '');
        const fullPath = `/${itemPath}`;
        return item.type === 'dir' ? `${fullPath}/` : fullPath;
      });
    } catch (error) {
      console.error(`Failed to list GitHub files at ${path}:`, error);
      return [];
    }
  }

  async readFile(path: string): Promise<string> {
    try {
      const normalizedPath = path.startsWith('/') ? path : `/${path}`;
      const repoPath = `${this.basePath}${normalizedPath}`.replace(/\/+/g, '/');
      const rawUrl = `${this.rawBase}/${this.repoOwner}/${this.repoName}/${this.branch}${repoPath}`;
      
      const response = await fetch(rawUrl);
      if (!response.ok) {
        throw new Error(`Failed to read file from GitHub: ${response.statusText}`);
      }
      
      return await response.text();
    } catch (error) {
      console.error(`Failed to read GitHub file ${path}:`, error);
      throw error;
    }
  }

  async writeFile(_path: string, _content: string): Promise<void> {
    throw new Error(`GitHub filesystem '${this.name}' is read-only`);
  }

  async deleteFile(_path: string): Promise<void> {
    throw new Error(`GitHub filesystem '${this.name}' is read-only`);
  }

  async createDirectory(_path: string): Promise<void> {
    throw new Error(`GitHub filesystem '${this.name}' is read-only`);
  }
}

// IndexedDB filesystem client - browser-local storage
export class IndexedDBFileSystemClient implements FileSystemClient {
  public readonly isReadOnly = false;
  private dbName: string;
  private db: IDBDatabase | null = null;
  
  constructor(
    public id: string,
    public name: string,
    public icon: string
  ) {
    this.dbName = `sdl-fs-${id}`;
  }

  private async ensureDB(): Promise<IDBDatabase> {
    if (this.db) return this.db;
    
    return new Promise((resolve, reject) => {
      const request = indexedDB.open(this.dbName, 1);
      
      request.onerror = () => reject(new Error('Failed to open IndexedDB'));
      
      request.onsuccess = () => {
        this.db = request.result;
        resolve(this.db);
      };
      
      request.onupgradeneeded = (event) => {
        const db = (event.target as IDBOpenDBRequest).result;
        
        // Create object store for files
        if (!db.objectStoreNames.contains('files')) {
          const store = db.createObjectStore('files', { keyPath: 'path' });
          store.createIndex('directory', 'directory');
        }
      };
    });
  }

  async listFiles(path: string): Promise<string[]> {
    const db = await this.ensureDB();
    const transaction = db.transaction(['files'], 'readonly');
    const store = transaction.objectStore('files');
    
    return new Promise((resolve, reject) => {
      const files: string[] = [];
      const normalizedPath = path.endsWith('/') ? path : `${path}/`;
      
      const request = store.openCursor();
      
      request.onsuccess = (event) => {
        const cursor = (event.target as IDBRequest).result;
        if (cursor) {
          const file = cursor.value;
          // Check if file is in the requested directory
          if (file.directory === normalizedPath || 
              (path === '/' && file.directory === '/')) {
            files.push(file.path);
          }
          cursor.continue();
        } else {
          resolve(files);
        }
      };
      
      request.onerror = () => reject(new Error('Failed to list files'));
    });
  }

  async readFile(path: string): Promise<string> {
    const db = await this.ensureDB();
    const transaction = db.transaction(['files'], 'readonly');
    const store = transaction.objectStore('files');
    
    return new Promise((resolve, reject) => {
      const request = store.get(path);
      
      request.onsuccess = () => {
        const file = request.result;
        if (file && file.content !== undefined) {
          resolve(file.content);
        } else {
          reject(new Error(`File not found: ${path}`));
        }
      };
      
      request.onerror = () => reject(new Error('Failed to read file'));
    });
  }

  async writeFile(path: string, content: string): Promise<void> {
    const db = await this.ensureDB();
    const transaction = db.transaction(['files'], 'readwrite');
    const store = transaction.objectStore('files');
    
    // Extract directory from path
    const directory = path.substring(0, path.lastIndexOf('/') + 1) || '/';
    
    return new Promise((resolve, reject) => {
      const request = store.put({
        path,
        directory,
        content,
        isDirectory: false,
        modTime: new Date().toISOString()
      });
      
      request.onsuccess = () => resolve();
      request.onerror = () => reject(new Error('Failed to write file'));
    });
  }

  async deleteFile(path: string): Promise<void> {
    const db = await this.ensureDB();
    const transaction = db.transaction(['files'], 'readwrite');
    const store = transaction.objectStore('files');
    
    return new Promise((resolve, reject) => {
      const request = store.delete(path);
      
      request.onsuccess = () => resolve();
      request.onerror = () => reject(new Error('Failed to delete file'));
    });
  }

  async createDirectory(path: string): Promise<void> {
    const db = await this.ensureDB();
    const transaction = db.transaction(['files'], 'readwrite');
    const store = transaction.objectStore('files');
    
    // Ensure path ends with /
    const dirPath = path.endsWith('/') ? path : `${path}/`;
    const parentDir = dirPath.substring(0, dirPath.lastIndexOf('/', dirPath.length - 2) + 1) || '/';
    
    return new Promise((resolve, reject) => {
      const request = store.put({
        path: dirPath,
        directory: parentDir,
        isDirectory: true,
        modTime: new Date().toISOString()
      });
      
      request.onsuccess = () => resolve();
      request.onerror = () => reject(new Error('Failed to create directory'));
    });
  }
}
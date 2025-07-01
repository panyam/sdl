import { FileSystemClient, LocalFileSystemClient, GitHubFileSystemClient } from '../filesystem-clients.js';

export interface FileNode {
  name: string;
  path: string;
  fsId: string;
  isDirectory: boolean;
  isReadOnly?: boolean;
  children?: FileNode[];
  expanded?: boolean;
}

export class MultiFSExplorer {
  private container: HTMLElement;
  private selectedFile: string | null = null;
  private onFileSelect?: (path: string, fsId: string) => void;
  private onFileCreate?: (path: string, fsId: string) => void;
  private fileSystems: FileSystemClient[] = [];
  private fileTreesByFS: Map<string, FileNode[]> = new Map();

  constructor(container: HTMLElement) {
    this.container = container;
    this.initializeDefaultFileSystems();
  }

  private initializeDefaultFileSystems() {
    // Default file systems using FileSystemClient implementations
    this.fileSystems = [
      new LocalFileSystemClient(
        '/api/filesystems',  // Server path for filesystem API
        'examples',          // ID
        'Examples',          // Display name
        false,               // Not read-only
        '📚'                // Icon
      ),
      new GitHubFileSystemClient(
        'github-examples',   // ID
        'GitHub Examples',   // Display name
        'panyam',           // Repo owner
        'sdl',              // Repo name
        'main',             // Branch
        '/examples',        // Base path in repo
        '🐙'                // Icon
      )
    ];
  }

  setFileSelectHandler(handler: (path: string, fsId: string) => void) {
    this.onFileSelect = handler;
  }

  setFileCreateHandler(handler: (path: string, fsId: string) => void) {
    this.onFileCreate = handler;
  }
  
  getFileSystem(fsId: string): FileSystemClient | undefined {
    return this.fileSystems.find(fs => fs.id === fsId);
  }
  
  getFileSystems(): FileSystemClient[] {
    return this.fileSystems;
  }

  async loadFileSystem(fs: FileSystemClient) {
    try {
      // Use the FileSystemClient interface to list files
      const files = await this.loadDirectoryRecursive(fs, '/');
      
      const tree = await this.buildFileTree(files, fs);
      this.fileTreesByFS.set(fs.id, tree);
    } catch (error) {
      console.error(`Failed to load filesystem ${fs.name}:`, error);
      this.fileTreesByFS.set(fs.id, []);
    }
  }

  private async loadDirectoryRecursive(fs: FileSystemClient, path: string): Promise<string[]> {
    const allFiles: string[] = [];
    
    try {
      const files = await fs.listFiles(path);
      
      for (const file of files) {
        allFiles.push(file);
        
        // If it's a directory (ends with /), recursively load its contents
        if (file.endsWith('/')) {
          const subFiles = await this.loadDirectoryRecursive(fs, file);
          allFiles.push(...subFiles);
        }
      }
    } catch (error) {
      console.error(`Failed to load directory ${path}:`, error);
    }
    
    return allFiles;
  }

  private async buildFileTree(files: string[], fs: FileSystemClient): Promise<FileNode[]> {
    const root: FileNode[] = [];
    const nodeMap = new Map<string, FileNode>();

    // Sort files to ensure directories come before their contents
    files.sort();

    files.forEach(filePath => {
      const parts = filePath.split('/').filter(p => p);
      let currentPath = '';
      let parentNodes = root;

      parts.forEach((part, index) => {
        currentPath += '/' + part;
        
        if (!nodeMap.has(currentPath)) {
          const isDirectory = index < parts.length - 1 || filePath.endsWith('/');
          const node: FileNode = {
            name: part,
            path: currentPath,
            fsId: fs.id,
            isDirectory,
            isReadOnly: fs.isReadOnly,
            children: isDirectory ? [] : undefined,
            expanded: false
          };
          
          nodeMap.set(currentPath, node);
          parentNodes.push(node);
        }

        const node = nodeMap.get(currentPath)!;
        if (index < parts.length - 1 && node.children) {
          parentNodes = node.children;
        }
      });
    });

    return root;
  }

  async initialize(dashboard?: any) {
    // Store dashboard reference for API access
    if (dashboard) {
      (window as any).dashboard = dashboard;
    }
    
    // Load all filesystems
    for (const fs of this.fileSystems) {
      await this.loadFileSystem(fs);
    }
    this.render();
  }

  private render() {
    this.container.innerHTML = `
      <div class="multi-fs-explorer">
        <div class="fs-header">
          <h3 class="text-sm font-semibold">File Systems</h3>
          <button class="btn-icon" title="Add/Manage FileSystems" onclick="window.multiFSExplorer?.manageFileSystems()">
            <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
              <path d="M14 8.5H8.5V14H7.5V8.5H2V7.5H7.5V2H8.5V7.5H14V8.5Z"/>
            </svg>
          </button>
        </div>
        <div class="fs-list">
          ${this.fileSystems.map(fs => this.renderFileSystem(fs)).join('')}
        </div>
      </div>
    `;

    // Make explorer globally accessible for button handlers
    (window as any).multiFSExplorer = this;
  }

  private renderFileSystem(fs: FileSystemClient): string {
    const tree = this.fileTreesByFS.get(fs.id) || [];
    const isEmpty = tree.length === 0;
    
    return `
      <div class="fs-section" data-fs-id="${fs.id}">
        <div class="fs-header-section">
          <div class="fs-title">
            <span class="fs-icon">${fs.icon || '📁'}</span>
            <span class="fs-name">${fs.name}</span>
            ${fs.isReadOnly ? '<span class="lock-icon" title="Read-only">🔒</span>' : ''}
          </div>
          <div class="fs-actions">
            ${!fs.isReadOnly ? `
              <button class="btn-icon" title="New File" onclick="window.multiFSExplorer?.createNewFile('${fs.id}')">
                <svg width="14" height="14" viewBox="0 0 16 16" fill="currentColor">
                  <path d="M9 7h6v1H9v6H8V8H2V7h6V1h1v6z"/>
                </svg>
              </button>
              <button class="btn-icon" title="New Folder" onclick="window.multiFSExplorer?.createNewFolder('${fs.id}')">
                <svg width="14" height="14" viewBox="0 0 16 16" fill="currentColor">
                  <path d="M7 3H2v11h12V5H8V3H7zm0-1h1v2h6v9H2V2h5z"/>
                  <path d="M9 7h4v1H9v4H8V8H4V7h4V4h1v3z"/>
                </svg>
              </button>
            ` : ''}
            <button class="btn-icon" title="Refresh" onclick="window.multiFSExplorer?.refreshFileSystem('${fs.id}')">
              <svg width="14" height="14" viewBox="0 0 16 16" fill="currentColor">
                <path d="M13.451 5.609l-.579-.939-1.068.812-.076.094c-.335.415-.927 1.341-1.124 2.876l-.021.165.033.163c.071.363.224.694.456.97l.087.102c.25.282.554.514.897.683l.123.061c.404.182.852.279 1.312.279.51 0 1.003-.12 1.444-.349l.105-.059c.435-.255.785-.618 1.014-1.051l.063-.119c.185-.38.283-.8.283-1.228 0-.347-.063-.684-.183-1.003l-.056-.147-.098-.245zm-3.177 3.342c-.169 0-.331-.037-.48-.109l-.044-.023c-.122-.061-.227-.145-.313-.249l-.032-.04c-.084-.106-.144-.227-.176-.361l-.012-.056c-.03-.137-.037-.283-.01-.428l.008-.059c.088-.987.373-1.76.603-2.122.183.338.276.735.276 1.142 0 .168-.02.332-.06.491l-.023.079c-.082.268-.225.51-.417.703l-.037.035c-.189.186-.423.325-.689.413l-.064.021c-.14.042-.288.063-.44.063zm1.373-4.326l2.255-1.718 1.017 1.647-2.351 1.79-.921-1.719zm-10.296.577l1.017-1.647 2.255 1.718-.921 1.719-2.351-1.79zM6.353 9.198c-.016-.196-.047-.39-.105-.576l-.024-.076c-.209-.586-.642-1.082-1.219-1.396l-.111-.058c-.369-.194-.79-.3-1.221-.308l-.085-.002c-.456-.007-.909.106-1.309.328l-.106.061c-.44.256-.79.62-1.013 1.053l-.063.12c-.186.378-.284.798-.284 1.226 0 .523.146 1.024.42 1.446l.075.112c.291.421.701.744 1.186.934l.122.046c.34.123.705.186 1.076.186.51 0 1.003-.12 1.444-.349l.105-.059c.347-.203.633-.485.839-.827l.053-.091c.175-.315.269-.668.269-1.027 0-.196-.029-.39-.084-.575l-.042-.133-.031-.097zm-2.031 1.754c-.17 0-.332-.037-.481-.109l-.044-.023c-.122-.061-.227-.145-.313-.249l-.032-.04c-.084-.106-.144-.227-.176-.361l-.012-.056c-.03-.137-.036-.282-.009-.427.027-.148.077-.291.15-.424l.032-.056c.079-.134.181-.252.303-.348l.039-.03c.117-.089.249-.157.392-.2l.062-.017c.133-.038.274-.057.418-.057.156 0 .307.022.451.065l.056.018c.142.048.271.119.382.214l.043.038c.236.212.385.516.422.866l.004.058c.004.045.006.091.006.137 0 .167-.02.331-.06.49l-.023.079c-.082.268-.225.51-.417.703l-.037.035c-.189.186-.423.325-.689.413l-.064.021c-.14.042-.289.063-.441.063z"/>
              </svg>
            </button>
          </div>
        </div>
        <div class="fs-tree">
          ${isEmpty ? 
            '<div class="empty-fs">No files</div>' : 
            this.renderTree(tree, fs.id, 0)
          }
        </div>
      </div>
    `;
  }

  private renderTree(nodes: FileNode[], _fsId: string, level: number = 0): string {
    return nodes.map(node => this.renderNode(node, level)).join('');
  }

  private renderNode(node: FileNode, level: number): string {
    const indent = level * 16 + 8;
    const icon = node.isDirectory 
      ? (node.expanded ? '📂' : '📁')
      : '📄';
    
    const nodeKey = `${node.fsId}:${node.path}`;
    const selected = nodeKey === this.selectedFile ? 'selected' : '';
    const readOnlyClass = node.isReadOnly ? 'readonly' : '';
    
    let html = `
      <div class="file-node ${selected} ${readOnlyClass}" style="padding-left: ${indent}px" 
           data-path="${node.path}" data-fs-id="${node.fsId}"
           onclick="window.multiFSExplorer?.selectFile('${node.path}', '${node.fsId}', ${node.isDirectory})">
        <span class="file-icon">${icon}</span>
        <span class="file-name">${node.name}</span>
      </div>
    `;

    if (node.isDirectory && node.expanded && node.children) {
      html += this.renderTree(node.children, node.fsId, level + 1);
    }

    return html;
  }

  selectFile(path: string, fsId: string, isDirectory: boolean) {
    this.selectedFile = `${fsId}:${path}`;
    
    // Update UI
    document.querySelectorAll('.file-node').forEach(el => {
      const elPath = el.getAttribute('data-path');
      const elFsId = el.getAttribute('data-fs-id');
      const isSelected = elPath === path && elFsId === fsId;
      el.classList.toggle('selected', isSelected);
    });

    if (!isDirectory && this.onFileSelect) {
      this.onFileSelect(path, fsId);
    } else if (isDirectory) {
      // Toggle directory expansion
      const tree = this.fileTreesByFS.get(fsId);
      if (tree) {
        const node = this.findNode(path, tree);
        if (node) {
          node.expanded = !node.expanded;
          this.render();
        }
      }
    }
  }

  private findNode(path: string, nodes: FileNode[]): FileNode | null {
    for (const node of nodes) {
      if (node.path === path) {
        return node;
      }
      if (node.children) {
        const found = this.findNode(path, node.children);
        if (found) return found;
      }
    }
    return null;
  }

  createNewFile(fsId: string) {
    const fs = this.fileSystems.find(f => f.id === fsId);
    if (!fs || fs.isReadOnly) return;

    const name = prompt('Enter file name:');
    if (name) {
      const path = `/${name}`;
      if (this.onFileCreate) {
        this.onFileCreate(path, fsId);
      }
    }
  }

  async createNewFolder(fsId: string) {
    const fs = this.fileSystems.find(f => f.id === fsId);
    if (!fs || fs.isReadOnly) return;

    const name = prompt('Enter folder name:');
    if (name) {
      const path = `/${name}`;
      try {
        // Create directory using FileSystemClient
        await fs.createDirectory(path);
        // Refresh the filesystem to show the new folder
        await this.refreshFileSystem(fsId);
      } catch (error) {
        console.error('Failed to create folder:', error);
        alert(`Failed to create folder: ${error}`);
      }
    }
  }

  async refreshFileSystem(fsId: string) {
    const fs = this.fileSystems.find(f => f.id === fsId);
    if (fs) {
      await this.loadFileSystem(fs);
      this.render();
    }
  }

  manageFileSystems() {
    // TODO: Show dialog to add/remove filesystems
    alert('FileSystem management coming soon!\n\nYou will be able to:\n- Add IndexedDB filesystem for local storage\n- Mount GitHub repositories\n- Configure custom file sources');
  }

  addFileSystem(fs: FileSystemClient) {
    this.fileSystems.push(fs);
    this.loadFileSystem(fs).then(() => this.render());
  }

  removeFileSystem(fsId: string) {
    this.fileSystems = this.fileSystems.filter(fs => fs.id !== fsId);
    this.fileTreesByFS.delete(fsId);
    this.render();
  }

  highlightFile(path: string, fsId: string) {
    // Update selected file with composite key
    this.selectedFile = `${fsId}:${path}`;
    
    // Update UI to highlight the file
    document.querySelectorAll('.file-node').forEach(el => {
      const elPath = el.getAttribute('data-path');
      const elFsId = el.getAttribute('data-fs-id');
      const isSelected = elPath === path && elFsId === fsId;
      el.classList.toggle('selected', isSelected);
    });
    
    // Ensure parent directories are expanded
    const tree = this.fileTreesByFS.get(fsId);
    if (tree) {
      this.expandParentDirectories(path, tree);
      this.render();
      
      // After render, scroll to the selected file
      setTimeout(() => {
        const selectedElement = document.querySelector(`.file-node[data-path="${CSS.escape(path)}"][data-fs-id="${CSS.escape(fsId)}"]`);
        if (selectedElement) {
          selectedElement.scrollIntoView({ behavior: 'smooth', block: 'center' });
        }
      }, 100);
    }
  }

  private expandParentDirectories(path: string, nodes: FileNode[]) {
    const parts = path.split('/').filter(p => p);
    let currentPath = '';
    
    for (let i = 0; i < parts.length - 1; i++) {
      currentPath += '/' + parts[i];
      const node = this.findNode(currentPath, nodes);
      if (node && node.isDirectory) {
        node.expanded = true;
      }
    }
  }
}

// Add styles
const style = document.createElement('style');
style.textContent = `
  .multi-fs-explorer {
    height: 100%;
    display: flex;
    flex-direction: column;
    background-color: #1e1e1e;
  }
  
  .fs-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 8px 12px;
    border-bottom: 1px solid #3e3e3e;
    background-color: #252526;
  }
  
  .fs-list {
    flex: 1;
    overflow-y: auto;
  }
  
  .fs-section {
    border-bottom: 1px solid #3e3e3e;
  }
  
  .fs-header-section {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 6px 8px;
    background-color: #2d2d30;
    border-top: 1px solid #3e3e3e;
  }
  
  .fs-title {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 13px;
    font-weight: 500;
  }
  
  .fs-icon {
    font-size: 16px;
  }
  
  .fs-name {
    color: #cccccc;
  }
  
  .fs-actions {
    display: flex;
    gap: 2px;
  }
  
  .btn-icon {
    background: none;
    border: none;
    color: #cccccc;
    cursor: pointer;
    padding: 3px;
    border-radius: 3px;
    opacity: 0.7;
    transition: opacity 0.2s;
  }
  
  .btn-icon:hover {
    background-color: #3e3e3e;
    opacity: 1;
  }
  
  .fs-tree {
    background-color: #1e1e1e;
    min-height: 50px;
  }
  
  .empty-fs {
    padding: 16px;
    text-align: center;
    color: #6e7681;
    font-size: 12px;
    font-style: italic;
  }
  
  .file-node {
    display: flex;
    align-items: center;
    padding: 2px 8px;
    cursor: pointer;
    user-select: none;
    font-size: 13px;
  }
  
  .file-node:hover {
    background-color: #2a2a2a;
  }
  
  .file-node.selected {
    background-color: #094771;
  }
  
  .file-icon {
    margin-right: 6px;
    font-size: 14px;
  }
  
  .file-name {
    color: #cccccc;
  }
  
  .file-node.readonly .file-name {
    font-style: italic;
    color: #999999;
  }
  
  .lock-icon {
    font-size: 11px;
    opacity: 0.7;
  }
`;
document.head.appendChild(style);

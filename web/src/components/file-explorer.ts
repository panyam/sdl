export interface FileNode {
  name: string;
  path: string;
  isDirectory: boolean;
  children?: FileNode[];
  expanded?: boolean;
}

export class FileExplorer {
  private container: HTMLElement;
  private selectedFile: string | null = null;
  private onFileSelect?: (path: string) => void;
  private onFileCreate?: (path: string) => void;
  private fileTree: FileNode[] = [];

  constructor(container: HTMLElement) {
    this.container = container;
  }

  setFileSelectHandler(handler: (path: string) => void) {
    this.onFileSelect = handler;
  }

  setFileCreateHandler(handler: (path: string) => void) {
    this.onFileCreate = handler;
  }

  async loadFiles(files: string[]) {
    // Build tree structure from flat file list
    this.fileTree = this.buildFileTree(files);
    this.render();
  }

  private buildFileTree(files: string[]): FileNode[] {
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
          const isDirectory = index < parts.length - 1 || !part.includes('.');
          const node: FileNode = {
            name: part,
            path: currentPath,
            isDirectory,
            children: isDirectory ? [] : undefined,
            expanded: currentPath.startsWith('/examples') || currentPath === '/workspace'
          };
          
          nodeMap.set(currentPath, node);
          parentNodes.push(node);
        }

        const node = nodeMap.get(currentPath)!;
        if (index < parts.length - 1) {
          parentNodes = node.children!;
        }
      });
    });

    return root;
  }

  private render() {
    this.container.innerHTML = `
      <div class="file-explorer">
        <div class="file-explorer-header">
          <h3 class="text-sm font-semibold mb-2">Files</h3>
          <div class="file-explorer-actions">
            <button class="btn-icon" title="New File" onclick="window.fileExplorer?.createNewFile()">
              <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
                <path d="M9 7h6v1H9v6H8V8H2V7h6V1h1v6z"/>
              </svg>
            </button>
            <button class="btn-icon" title="New Folder" onclick="window.fileExplorer?.createNewFolder()">
              <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
                <path d="M7 3H2v11h12V5H8V3H7zm0-1h1v2h6v9H2V2h5z"/>
                <path d="M9 7h4v1H9v4H8V8H4V7h4V4h1v3z"/>
              </svg>
            </button>
            <button class="btn-icon" title="Refresh" onclick="window.fileExplorer?.refresh()">
              <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
                <path d="M13.451 5.609l-.579-.939-1.068.812-.076.094c-.335.415-.927 1.341-1.124 2.876l-.021.165.033.163c.071.363.224.694.456.97l.087.102c.25.282.554.514.897.683l.123.061c.404.182.852.279 1.312.279.51 0 1.003-.12 1.444-.349l.105-.059c.435-.255.785-.618 1.014-1.051l.063-.119c.185-.38.283-.8.283-1.228 0-.347-.063-.684-.183-1.003l-.056-.147-.098-.245zm-3.177 3.342c-.169 0-.331-.037-.48-.109l-.044-.023c-.122-.061-.227-.145-.313-.249l-.032-.04c-.084-.106-.144-.227-.176-.361l-.012-.056c-.03-.137-.037-.283-.01-.428l.008-.059c.088-.987.373-1.76.603-2.122.183.338.276.735.276 1.142 0 .168-.02.332-.06.491l-.023.079c-.082.268-.225.51-.417.703l-.037.035c-.189.186-.423.325-.689.413l-.064.021c-.14.042-.288.063-.44.063zm1.373-4.326l2.255-1.718 1.017 1.647-2.351 1.79-.921-1.719zm-10.296.577l1.017-1.647 2.255 1.718-.921 1.719-2.351-1.79zM6.353 9.198c-.016-.196-.047-.39-.105-.576l-.024-.076c-.209-.586-.642-1.082-1.219-1.396l-.111-.058c-.369-.194-.79-.3-1.221-.308l-.085-.002c-.456-.007-.909.106-1.309.328l-.106.061c-.44.256-.79.62-1.013 1.053l-.063.12c-.186.378-.284.798-.284 1.226 0 .523.146 1.024.42 1.446l.075.112c.291.421.701.744 1.186.934l.122.046c.34.123.705.186 1.076.186.51 0 1.003-.12 1.444-.349l.105-.059c.347-.203.633-.485.839-.827l.053-.091c.175-.315.269-.668.269-1.027 0-.196-.029-.39-.084-.575l-.042-.133-.031-.097zm-2.031 1.754c-.17 0-.332-.037-.481-.109l-.044-.023c-.122-.061-.227-.145-.313-.249l-.032-.04c-.084-.106-.144-.227-.176-.361l-.012-.056c-.03-.137-.036-.282-.009-.427.027-.148.077-.291.15-.424l.032-.056c.079-.134.181-.252.303-.348l.039-.03c.117-.089.249-.157.392-.2l.062-.017c.133-.038.274-.057.418-.057.156 0 .307.022.451.065l.056.018c.142.048.271.119.382.214l.043.038c.236.212.385.516.422.866l.004.058c.004.045.006.091.006.137 0 .167-.02.331-.06.49l-.023.079c-.082.268-.225.51-.417.703l-.037.035c-.189.186-.423.325-.689.413l-.064.021c-.14.042-.289.063-.441.063z"/>
              </svg>
            </button>
          </div>
        </div>
        <div class="file-tree">
          ${this.renderTree(this.fileTree)}
        </div>
      </div>
    `;

    // Make file explorer globally accessible for button handlers
    (window as any).fileExplorer = this;
  }

  private renderTree(nodes: FileNode[], level: number = 0): string {
    return nodes.map(node => this.renderNode(node, level)).join('');
  }

  private renderNode(node: FileNode, level: number): string {
    const indent = level * 16;
    const icon = node.isDirectory 
      ? (node.expanded ? '📂' : '📁')
      : '📄';
    
    const selected = node.path === this.selectedFile ? 'selected' : '';
    
    let html = `
      <div class="file-node ${selected}" style="padding-left: ${indent}px" 
           data-path="${node.path}"
           onclick="window.fileExplorer?.selectFile('${node.path}', ${node.isDirectory})">
        <span class="file-icon">${icon}</span>
        <span class="file-name">${node.name}</span>
      </div>
    `;

    if (node.isDirectory && node.expanded && node.children) {
      html += this.renderTree(node.children, level + 1);
    }

    return html;
  }

  selectFile(path: string, isDirectory: boolean) {
    this.selectedFile = path;
    
    // Update UI
    document.querySelectorAll('.file-node').forEach(el => {
      el.classList.toggle('selected', el.getAttribute('data-path') === path);
    });

    if (!isDirectory && this.onFileSelect) {
      this.onFileSelect(path);
    } else if (isDirectory) {
      // Toggle directory expansion
      const node = this.findNode(path);
      if (node) {
        node.expanded = !node.expanded;
        this.render();
      }
    }
  }

  private findNode(path: string, nodes: FileNode[] = this.fileTree): FileNode | null {
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

  createNewFile() {
    const name = prompt('Enter file name:');
    if (name) {
      const dir = this.selectedFile && this.findNode(this.selectedFile)?.isDirectory 
        ? this.selectedFile 
        : '/workspace';
      const path = `${dir}/${name}`;
      if (this.onFileCreate) {
        this.onFileCreate(path);
      }
    }
  }

  createNewFolder() {
    const name = prompt('Enter folder name:');
    if (name) {
      const dir = this.selectedFile && this.findNode(this.selectedFile)?.isDirectory 
        ? this.selectedFile 
        : '/workspace';
      const path = `${dir}/${name}`;
      // Create empty folder by adding it to the tree
      this.fileTree.push({
        name: name,
        path: path,
        isDirectory: true,
        children: [],
        expanded: true
      });
      this.render();
    }
  }

  async refresh() {
    // This should be called by the parent component to reload files
    console.log('Refresh requested');
  }
}

// Add styles
const style = document.createElement('style');
style.textContent = `
  .file-explorer {
    height: 100%;
    display: flex;
    flex-direction: column;
  }
  
  .file-explorer-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 8px;
    border-bottom: 1px solid #3e3e3e;
  }
  
  .file-explorer-actions {
    display: flex;
    gap: 4px;
  }
  
  .btn-icon {
    background: none;
    border: none;
    color: #cccccc;
    cursor: pointer;
    padding: 4px;
    border-radius: 4px;
  }
  
  .btn-icon:hover {
    background-color: #2a2a2a;
  }
  
  .file-tree {
    flex: 1;
    overflow-y: auto;
  }
  
  .file-node {
    display: flex;
    align-items: center;
    padding: 4px 8px;
    cursor: pointer;
    user-select: none;
  }
  
  .file-node:hover {
    background-color: #2a2a2a;
  }
  
  .file-node.selected {
    background-color: #094771;
  }
  
  .file-icon {
    margin-right: 8px;
  }
  
  .file-name {
    color: #cccccc;
  }
`;
document.head.appendChild(style);
import { CanvasClient } from '../canvas-client.js';

export interface RecipeCommand {
  lineNumber: number;
  rawLine: string;
  type: 'command' | 'comment' | 'echo' | 'pause' | 'empty';
  command?: string;
  args?: string[];
  description?: string;
}

export interface RecipeStep {
  index: number;
  command: RecipeCommand;
  status: 'pending' | 'running' | 'completed' | 'failed' | 'skipped';
  output?: string;
  error?: string;
  startTime?: number;
  endTime?: number;
}

export interface RecipeState {
  filePath: string;
  fileName: string;
  steps: RecipeStep[];
  currentStep: number;
  isRunning: boolean;
  isPaused: boolean;
  mode: 'step' | 'auto';
  autoDelay: number; // milliseconds between auto steps
}

export class RecipeRunner {
  private api: CanvasClient;
  private state: RecipeState | null = null;
  private onStateChange?: (state: RecipeState) => void;
  private onOutput?: (message: string, type: 'info' | 'success' | 'error' | 'warning') => void;
  private abortController?: AbortController;

  constructor(api: CanvasClient) {
    this.api = api;
  }

  setStateChangeHandler(handler: (state: RecipeState) => void) {
    this.onStateChange = handler;
  }

  setOutputHandler(handler: (message: string, type: 'info' | 'success' | 'error' | 'warning') => void) {
    this.onOutput = handler;
  }

  async loadRecipe(filePath: string, content: string): Promise<RecipeState> {
    const fileName = filePath.split('/').pop() || 'recipe';
    
    try {
      const commands = this.parseRecipe(content);
      
      this.state = {
        filePath,
        fileName,
        steps: commands.map((cmd, index) => ({
          index,
          command: cmd,
          status: 'pending'
        })),
        currentStep: 0,
        isRunning: false,
        isPaused: false,
        mode: 'step',
        autoDelay: 1000
      };

      this.notifyStateChange();
      return this.state;
    } catch (error: any) {
      // Parse error - show to user
      this.output(`❌ Recipe parse error:\n${error.message}`, 'error');
      throw error;
    }
  }

  private parseRecipe(content: string): RecipeCommand[] {
    const lines = content.split('\n');
    const commands: RecipeCommand[] = [];
    const errors: string[] = [];

    lines.forEach((line, index) => {
      const trimmed = line.trim();
      const lineNumber = index + 1;

      // Empty line
      if (trimmed === '') {
        commands.push({
          lineNumber,
          rawLine: line,
          type: 'empty'
        });
        return;
      }

      // Comment line
      if (trimmed.startsWith('#')) {
        commands.push({
          lineNumber,
          rawLine: line,
          type: 'comment',
          description: trimmed.substring(1).trim()
        });
        return;
      }

      // Echo statement (description)
      if (trimmed.startsWith('echo ')) {
        const echoContent = trimmed.substring(5).trim();
        // Validate echo syntax
        if (!echoContent) {
          errors.push(`Line ${lineNumber}: Empty echo statement`);
        }
        // Check for variable expansion
        if (echoContent.includes('$')) {
          errors.push(`Line ${lineNumber}: Variable expansion not supported in echo statements`);
        }
        commands.push({
          lineNumber,
          rawLine: line,
          type: 'echo',
          description: echoContent.replace(/^["']|["']$/g, '')
        });
        return;
      }

      // Read command (pause point)
      if (trimmed === 'read' || trimmed.startsWith('read ')) {
        // Check for read with variables
        if (trimmed.includes(' ') && trimmed !== 'read') {
          errors.push(`Line ${lineNumber}: 'read' with variables not supported. Use plain 'read' for pause points`);
        }
        commands.push({
          lineNumber,
          rawLine: line,
          type: 'pause',
          description: 'Press continue to proceed'
        });
        return;
      }

      // SDL command
      if (trimmed.startsWith('sdl ')) {
        const parts = this.parseCommandLine(trimmed);
        if (parts.length > 1) {
          // Validate SDL command
          const validCommands = ['load', 'use', 'gen', 'metrics', 'set', 'canvas'];
          const sdlCommand = parts[1];
          if (!validCommands.includes(sdlCommand)) {
            errors.push(`Line ${lineNumber}: Unknown SDL command '${sdlCommand}'. Valid commands: ${validCommands.join(', ')}`);
          }
          
          // Check for variable expansion in SDL commands
          if (trimmed.includes('$')) {
            errors.push(`Line ${lineNumber}: Variable expansion not supported in SDL commands`);
          }
          
          commands.push({
            lineNumber,
            rawLine: line,
            type: 'command',
            command: parts[0],
            args: parts.slice(1)
          });
        } else {
          errors.push(`Line ${lineNumber}: SDL command missing arguments`);
        }
        return;
      }

      // Check for unsupported shell syntax
      const unsupportedPatterns = [
        { pattern: /^\s*if\s+/, message: 'if statements not supported' },
        { pattern: /^\s*for\s+/, message: 'for loops not supported' },
        { pattern: /^\s*while\s+/, message: 'while loops not supported' },
        { pattern: /^\s*case\s+/, message: 'case statements not supported' },
        { pattern: /^\s*function\s+/, message: 'function definitions not supported' },
        { pattern: /.*\|.*/, message: 'pipes not supported' },
        { pattern: /.*>>?.*/, message: 'redirections not supported' },
        { pattern: /.*<.*/, message: 'input redirection not supported' },
        { pattern: /.*\$\(.*\)/, message: 'command substitution not supported' },
        { pattern: /.*`.*`/, message: 'backtick command substitution not supported' },
        { pattern: /^\s*export\s+/, message: 'export not supported' },
        { pattern: /^\s*source\s+/, message: 'source not supported' },
        { pattern: /^\s*\.\s+/, message: 'source (.) not supported' },
        { pattern: /.*\$\{.*\}/, message: 'variable expansion not supported' },
        { pattern: /.*\$\w+/, message: 'variables not supported' },
        { pattern: /.*&\s*$/, message: 'background jobs not supported' },
        { pattern: /^\s*\[.*\]/, message: 'test expressions not supported' },
        { pattern: /^\s*\[\[.*\]\]/, message: 'test expressions not supported' },
        { pattern: /.*\$\(\(.*\)\)/, message: 'arithmetic expansion not supported' }
      ];

      let foundUnsupported = false;
      for (const { pattern, message } of unsupportedPatterns) {
        if (pattern.test(trimmed)) {
          errors.push(`Line ${lineNumber}: ${message} - ${trimmed}`);
          foundUnsupported = true;
          break;
        }
      }

      // Check for other executable commands
      if (!foundUnsupported && !trimmed.startsWith('#')) {
        const firstWord = trimmed.split(/\s+/)[0];
        // List of shell built-ins and common commands that aren't supported
        const unsupportedCommands = [
          'cd', 'pwd', 'ls', 'mkdir', 'rm', 'cp', 'mv', 'cat', 'grep', 'sed', 'awk',
          'find', 'chmod', 'chown', 'curl', 'wget', 'git', 'npm', 'yarn', 'python',
          'node', 'bash', 'sh', 'zsh', 'exit', 'return', 'break', 'continue'
        ];
        
        if (unsupportedCommands.includes(firstWord)) {
          errors.push(`Line ${lineNumber}: Command '${firstWord}' not supported. Only 'sdl', 'echo', and 'read' commands are allowed`);
        } else if (!['echo', 'read', 'sdl'].includes(firstWord)) {
          errors.push(`Line ${lineNumber}: Unknown command '${firstWord}'. Only 'sdl', 'echo', and 'read' commands are supported`);
        }
      }

      // Add as comment for unsupported lines
      commands.push({
        lineNumber,
        rawLine: line,
        type: 'comment',
        description: `[Unsupported: ${trimmed.substring(0, 50)}${trimmed.length > 50 ? '...' : ''}]`
      });
    });

    // If there are errors, throw them
    if (errors.length > 0) {
      const errorMessage = `Recipe syntax errors:\n${errors.join('\n')}`;
      throw new Error(errorMessage);
    }

    return commands;
  }

  private parseCommandLine(line: string): string[] {
    // Simple command line parser that handles quoted strings
    const parts: string[] = [];
    let current = '';
    let inQuote = false;
    let quoteChar = '';

    for (let i = 0; i < line.length; i++) {
      const char = line[i];
      
      if (inQuote) {
        if (char === quoteChar) {
          inQuote = false;
          quoteChar = '';
        } else {
          current += char;
        }
      } else {
        if (char === '"' || char === "'") {
          inQuote = true;
          quoteChar = char;
        } else if (char === ' ' || char === '\t') {
          if (current) {
            parts.push(current);
            current = '';
          }
        } else {
          current += char;
        }
      }
    }

    if (current) {
      parts.push(current);
    }

    return parts;
  }

  async start(mode: 'step' | 'auto' = 'step') {
    if (!this.state || this.state.isRunning) return;

    this.state.mode = mode;
    this.state.isRunning = true;
    this.state.isPaused = false;
    this.abortController = new AbortController();
    
    this.notifyStateChange();

    if (mode === 'auto') {
      this.runAutoMode();
    }
  }

  async step() {
    if (!this.state || !this.state.isRunning) return;
    if (this.state.currentStep >= this.state.steps.length) return;

    await this.executeCurrentStep();
  }

  pause() {
    if (!this.state || !this.state.isRunning) return;
    
    this.state.isPaused = true;
    this.notifyStateChange();
  }

  resume() {
    if (!this.state || !this.state.isRunning || !this.state.isPaused) return;
    
    this.state.isPaused = false;
    this.notifyStateChange();

    if (this.state.mode === 'auto') {
      this.runAutoMode();
    }
  }

  stop() {
    if (!this.state) return;

    this.state.isRunning = false;
    this.state.isPaused = false;
    this.abortController?.abort();
    
    // Mark remaining steps as skipped
    for (let i = this.state.currentStep; i < this.state.steps.length; i++) {
      if (this.state.steps[i].status === 'pending') {
        this.state.steps[i].status = 'skipped';
      }
    }

    this.notifyStateChange();
  }

  restart() {
    if (!this.state) return;

    // Reset all steps
    this.state.steps.forEach(step => {
      step.status = 'pending';
      step.output = undefined;
      step.error = undefined;
      step.startTime = undefined;
      step.endTime = undefined;
    });

    this.state.currentStep = 0;
    this.state.isRunning = false;
    this.state.isPaused = false;

    this.notifyStateChange();
  }

  private async runAutoMode() {
    if (!this.state) return;

    while (this.state.isRunning && !this.state.isPaused && this.state.currentStep < this.state.steps.length) {
      await this.executeCurrentStep();
      
      // Wait for auto delay if not paused
      if (this.state.isRunning && !this.state.isPaused && this.state.currentStep < this.state.steps.length) {
        await this.delay(this.state.autoDelay);
      }
    }

    // Auto-stop when done
    if (this.state.currentStep >= this.state.steps.length) {
      this.stop();
    }
  }

  private async executeCurrentStep() {
    if (!this.state || this.state.currentStep >= this.state.steps.length) return;

    const step = this.state.steps[this.state.currentStep];
    step.status = 'running';
    step.startTime = Date.now();
    this.notifyStateChange();

    try {
      const result = await this.executeCommand(step.command);
      
      step.status = 'completed';
      step.output = result;
      step.endTime = Date.now();

      // Handle pause commands
      if (step.command.type === 'pause' && this.state.mode === 'auto') {
        this.pause();
        this.output('⏸️ Recipe paused. Click resume to continue.', 'info');
      }

    } catch (error: any) {
      step.status = 'failed';
      step.error = error.message || 'Unknown error';
      step.endTime = Date.now();
      
      this.output(`❌ Step failed: ${error.message}`, 'error');
      
      // Stop on error
      this.stop();
    }

    this.state.currentStep++;
    this.notifyStateChange();
  }

  private async executeCommand(command: RecipeCommand): Promise<string> {
    switch (command.type) {
      case 'comment':
      case 'empty':
        return '';

      case 'echo':
        this.output(command.description || '', 'info');
        return command.description || '';

      case 'pause':
        return 'Paused';

      case 'command':
        return await this.executeSDLCommand(command.command!, command.args || []);

      default:
        return '';
    }
  }

  private async executeSDLCommand(command: string, args: string[]): Promise<string> {
    // Remove 'sdl' prefix if present
    if (command === 'sdl' && args.length > 0) {
      command = args[0];
      args = args.slice(1);
    }

    try {
      switch (command) {
        case 'load':
          if (args.length < 1) throw new Error('Missing file path');
          await this.api.loadFile(args[0]);
          this.output(`✅ Loaded ${args[0]}`, 'success');
          return `Loaded ${args[0]}`;

        case 'use':
          if (args.length < 1) throw new Error('Missing system name');
          await this.api.useSystem(args[0]);
          this.output(`✅ Using system ${args[0]}`, 'success');
          return `Using system ${args[0]}`;

        case 'gen':
          return await this.executeGeneratorCommand(args);

        case 'metrics':
          return await this.executeMetricsCommand(args);

        case 'set':
          if (args.length < 2) throw new Error('Missing parameter path or value');
          await this.api.setParameter(args[0], args[1]);
          this.output(`✅ Set ${args[0]} = ${args[1]}`, 'success');
          return `Set ${args[0]} = ${args[1]}`;

        case 'canvas':
          return await this.executeCanvasCommand(args);

        default:
          throw new Error(`Unknown SDL command: ${command}`);
      }
    } catch (error: any) {
      throw new Error(`SDL command failed: ${error.message}`);
    }
  }

  private async executeGeneratorCommand(args: string[]): Promise<string> {
    if (args.length < 1) throw new Error('Missing generator subcommand');

    const subcommand = args[0];
    const subArgs = args.slice(1);

    switch (subcommand) {
      case 'add':
        if (subArgs.length < 3) throw new Error('Usage: gen add <name> <component.method> <rate>');
        const [name, target, rateStr, ...flags] = subArgs;
        const [component, method] = target.split('.');
        const rate = parseFloat(rateStr);
        
        // Check for --apply-flows flag
        const applyFlows = flags.includes('--apply-flows');
        
        await this.api.addGenerator(name, component, method, rate);
        if (applyFlows) {
          // TODO: Apply flows after adding generator
          this.output(`✅ Added generator ${name} (flows will be applied)`, 'success');
        } else {
          this.output(`✅ Added generator ${name}`, 'success');
        }
        return `Added generator ${name}`;

      case 'start':
        if (subArgs.length === 0 || subArgs[0] === '--all') {
          await this.api.startAllGenerators();
          this.output('✅ Started all generators', 'success');
          return 'Started all generators';
        } else {
          await this.api.startGenerator(subArgs[0]);
          this.output(`✅ Started generator ${subArgs[0]}`, 'success');
          return `Started generator ${subArgs[0]}`;
        }

      case 'stop':
        if (subArgs.length === 0 || subArgs[0] === '--all') {
          await this.api.stopAllGenerators();
          this.output('✅ Stopped all generators', 'success');
          return 'Stopped all generators';
        } else {
          await this.api.stopGenerator(subArgs[0]);
          this.output(`✅ Stopped generator ${subArgs[0]}`, 'success');
          return `Stopped generator ${subArgs[0]}`;
        }

      case 'update':
        if (subArgs.length < 2) throw new Error('Usage: gen update <name> <rate>');
        await this.api.updateGeneratorRate(subArgs[0], parseFloat(subArgs[1]));
        this.output(`✅ Updated generator ${subArgs[0]} rate to ${subArgs[1]}`, 'success');
        return `Updated generator ${subArgs[0]}`;

      case 'delete':
        if (subArgs.length < 1) throw new Error('Usage: gen delete <name>');
        await this.api.deleteGenerator(subArgs[0]);
        this.output(`✅ Deleted generator ${subArgs[0]}`, 'success');
        return `Deleted generator ${subArgs[0]}`;

      default:
        throw new Error(`Unknown generator subcommand: ${subcommand}`);
    }
  }

  private async executeMetricsCommand(args: string[]): Promise<string> {
    if (args.length < 1) throw new Error('Missing metrics subcommand');

    const subcommand = args[0];
    const subArgs = args.slice(1);

    switch (subcommand) {
      case 'add':
        if (subArgs.length < 3) throw new Error('Usage: metrics add <name> <component> <method> [options]');
        const [name, component, method, ...options] = subArgs;
        
        // Parse options
        let metricType = 'count';
        let aggregation = 'sum';
        let window = 1;
        
        for (let i = 0; i < options.length; i += 2) {
          const flag = options[i];
          const value = options[i + 1];
          
          switch (flag) {
            case '--type':
              metricType = value;
              break;
            case '--aggregation':
              aggregation = value;
              break;
            case '--window':
              window = parseInt(value);
              break;
          }
        }
        
        await this.api.addMetric(name, component, [method], metricType, aggregation, window);
        this.output(`✅ Added metric ${name}`, 'success');
        return `Added metric ${name}`;

      default:
        throw new Error(`Unknown metrics subcommand: ${subcommand}`);
    }
  }

  private async executeCanvasCommand(args: string[]): Promise<string> {
    if (args.length < 1) throw new Error('Missing canvas subcommand');

    const subcommand = args[0];

    switch (subcommand) {
      case 'create':
        await this.api.createCanvas();
        this.output('✅ Created canvas', 'success');
        return 'Created canvas';

      case 'reset':
        // TODO: Implement reset
        this.output('⚠️ Canvas reset not implemented yet', 'warning');
        return 'Canvas reset not implemented';

      default:
        throw new Error(`Unknown canvas subcommand: ${subcommand}`);
    }
  }

  private output(message: string, type: 'info' | 'success' | 'error' | 'warning') {
    if (this.onOutput) {
      this.onOutput(message, type);
    }
  }

  private notifyStateChange() {
    if (this.onStateChange && this.state) {
      this.onStateChange(this.state);
    }
  }

  private delay(ms: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, ms));
  }

  getState(): RecipeState | null {
    return this.state;
  }

  setAutoDelay(delay: number) {
    if (this.state) {
      this.state.autoDelay = Math.max(100, delay); // Minimum 100ms
      this.notifyStateChange();
    }
  }
}
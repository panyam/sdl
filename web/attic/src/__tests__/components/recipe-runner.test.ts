import { describe, it, expect, vi, beforeEach } from 'vitest';
import { RecipeRunner } from '../../components/recipe-runner';
import { createMockElement } from '../utils/test-helpers';

describe('RecipeRunner', () => {
  let runner: RecipeRunner;
  let mockOutput: HTMLElement;
  let mockOnStep: any;
  let mockOnComplete: any;

  beforeEach(() => {
    mockOutput = createMockElement('div');
    mockOnStep = vi.fn();
    mockOnComplete = vi.fn();
    
    runner = new RecipeRunner(mockOutput, mockOnStep, mockOnComplete);
  });

  describe('Recipe Parsing', () => {
    it('should parse valid recipe commands', () => {
      const recipe = `
        echo "Starting deployment"
        start metrics
        sdl deploy examples/basic.sdl
        stop
      `;
      
      const steps = runner.parseRecipe(recipe);
      
      expect(steps).toHaveLength(4);
      expect(steps[0]).toEqual({ type: 'command', value: 'echo "Starting deployment"' });
      expect(steps[1]).toEqual({ type: 'command', value: 'start metrics' });
      expect(steps[2]).toEqual({ type: 'command', value: 'sdl deploy examples/basic.sdl' });
      expect(steps[3]).toEqual({ type: 'command', value: 'stop' });
    });

    it('should ignore empty lines and comments', () => {
      const recipe = `
        # This is a comment
        echo "Hello"
        
        # Another comment
        echo "World"
      `;
      
      const steps = runner.parseRecipe(recipe);
      
      expect(steps).toHaveLength(2);
      expect(steps[0].value).toBe('echo "Hello"');
      expect(steps[1].value).toBe('echo "World"');
    });

    it('should parse pause commands', () => {
      const recipe = `
        echo "Step 1"
        pause
        echo "Step 2"
      `;
      
      const steps = runner.parseRecipe(recipe);
      
      expect(steps).toHaveLength(3);
      expect(steps[1]).toEqual({ type: 'pause', value: 'pause' });
    });
  });

  describe('Recipe Validation', () => {
    it('should detect unsupported shell features', () => {
      const errors = runner.validateRecipe('echo "test" | grep test');
      
      expect(errors).toHaveLength(1);
      expect(errors[0]).toMatchObject({
        lineNumber: 1,
        message: expect.stringContaining('Pipes'),
        severity: 'error'
      });
    });

    it('should detect function declarations', () => {
      const errors = runner.validateRecipe('function test() { echo "test"; }');
      
      expect(errors).toHaveLength(1);
      expect(errors[0].message).toContain('Function declarations');
    });

    it('should allow dollar signs in quoted strings', () => {
      const errors = runner.validateRecipe('echo "Price: $20/month"');
      
      expect(errors).toHaveLength(0);
    });

    it('should detect unquoted variables', () => {
      const errors = runner.validateRecipe('echo $HOME');
      
      expect(errors).toHaveLength(1);
      expect(errors[0].message).toContain('Variables');
    });

    it('should validate multiple lines', () => {
      const recipe = `
        echo "Valid line"
        echo test | grep test
        echo "Another valid line"
        for i in *; do echo $i; done
      `;
      
      const errors = runner.validateRecipe(recipe);
      
      expect(errors).toHaveLength(2);
      expect(errors[0].lineNumber).toBe(2);
      expect(errors[1].lineNumber).toBe(4);
    });
  });

  describe('Recipe Execution', () => {
    it('should execute recipe steps', async () => {
      const recipe = `
        echo "Step 1"
        echo "Step 2"
      `;
      
      await runner.runRecipe(recipe);
      
      expect(mockOnStep).toHaveBeenCalledTimes(2);
      expect(mockOnStep).toHaveBeenNthCalledWith(1, 0, 'echo "Step 1"');
      expect(mockOnStep).toHaveBeenNthCalledWith(2, 1, 'echo "Step 2"');
      expect(mockOnComplete).toHaveBeenCalled();
    });

    it('should handle pause commands in manual mode', async () => {
      const recipe = `
        echo "Before pause"
        pause
        echo "After pause"
      `;
      
      runner.runRecipe(recipe);
      
      // First step should execute
      expect(mockOnStep).toHaveBeenCalledWith(0, 'echo "Before pause"');
      
      // Should pause at pause command
      await vi.waitFor(() => {
        expect(runner.isPaused()).toBe(true);
      });
      
      // Continue execution
      runner.resume();
      
      // Remaining steps should execute
      await vi.waitFor(() => {
        expect(mockOnComplete).toHaveBeenCalled();
      });
    });

    it('should stop execution when stop is called', async () => {
      const recipe = `
        echo "Step 1"
        echo "Step 2"
        echo "Step 3"
      `;
      
      runner.runRecipe(recipe);
      
      // Stop after first step
      await vi.waitFor(() => {
        expect(mockOnStep).toHaveBeenCalledWith(0, 'echo "Step 1"');
      });
      
      runner.stop();
      
      // Should not execute remaining steps
      expect(runner.isRunning()).toBe(false);
      expect(mockOnStep).toHaveBeenCalledTimes(1);
    });

    it('should handle step mode execution', async () => {
      const recipe = `
        echo "Step 1"
        echo "Step 2"
      `;
      
      runner.runRecipe(recipe);
      runner.setStepMode(true);
      
      // Should execute first step and pause
      await vi.waitFor(() => {
        expect(mockOnStep).toHaveBeenCalledWith(0, 'echo "Step 1"');
        expect(runner.isPaused()).toBe(true);
      });
      
      // Step to next
      runner.step();
      
      await vi.waitFor(() => {
        expect(mockOnStep).toHaveBeenCalledWith(1, 'echo "Step 2"');
      });
    });
  });

  describe('Output Handling', () => {
    it('should update output element with command results', async () => {
      const recipe = 'echo "Hello World"';
      
      await runner.runRecipe(recipe);
      
      expect(mockOutput.innerHTML).toContain('Hello World');
    });

    it('should show command being executed', async () => {
      const recipe = 'sdl deploy test.sdl';
      
      runner.runRecipe(recipe);
      
      await vi.waitFor(() => {
        expect(mockOutput.innerHTML).toContain('$ sdl deploy test.sdl');
      });
    });

    it('should handle errors gracefully', async () => {
      const recipe = 'invalid-command';
      
      await runner.runRecipe(recipe);
      
      expect(mockOutput.innerHTML).toContain('Command not found');
      expect(mockOnComplete).toHaveBeenCalled();
    });
  });

  describe('State Management', () => {
    it('should track running state correctly', () => {
      expect(runner.isRunning()).toBe(false);
      
      runner.runRecipe('echo "test"');
      expect(runner.isRunning()).toBe(true);
      
      runner.stop();
      expect(runner.isRunning()).toBe(false);
    });

    it('should track current step', async () => {
      const recipe = `
        echo "Step 1"
        echo "Step 2"
        echo "Step 3"
      `;
      
      runner.runRecipe(recipe);
      
      await vi.waitFor(() => {
        expect(runner.getCurrentStep()).toBeGreaterThanOrEqual(0);
      });
    });

    it('should reset state on stop', () => {
      runner.runRecipe('echo "test"');
      runner.stop();
      
      expect(runner.getCurrentStep()).toBe(0);
      expect(runner.isPaused()).toBe(false);
      expect(runner.isRunning()).toBe(false);
    });
  });
});
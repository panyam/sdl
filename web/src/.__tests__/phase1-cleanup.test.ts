/**
 * Phase 1 Verification Tests: Clean Foundation
 *
 * These tests verify that the dead code cleanup from Phase 1 was done correctly:
 * - Dead code files have been moved to attic (not importable from active code)
 * - Active code has no stale imports referencing moved files
 * - The main entry point (main.ts) only references active page types
 * - DockView theme class is applied to the container
 * - The CanvasViewerPage template sets window.sdlPageData
 */

import { describe, it, expect } from 'vitest';
import * as fs from 'fs';
import * as path from 'path';

// __dirname is web/src/.__tests__, so ../.. is web/src and ../../.. is web/
const SRC_DIR = path.resolve(__dirname, '..');
const WEB_DIR = path.resolve(__dirname, '../..');

/**
 * Recursively collect all .ts files in a directory, excluding node_modules, gen, attic, and .__tests__
 */
function collectTsFiles(dir: string, files: string[] = []): string[] {
  if (!fs.existsSync(dir)) return files;
  for (const entry of fs.readdirSync(dir, { withFileTypes: true })) {
    const fullPath = path.join(dir, entry.name);
    if (entry.isDirectory()) {
      if (['node_modules', 'gen', 'attic', 'dist', '.__tests__'].includes(entry.name)) continue;
      collectTsFiles(fullPath, files);
    } else if (entry.name.endsWith('.ts') && !entry.name.endsWith('.d.ts')) {
      files.push(fullPath);
    }
  }
  return files;
}

describe('Phase 1: Dead Code Cleanup', () => {
  const DEAD_MODULES = [
    'system-details-page',
    'wasm-system-detail-tool',
    'wasm-dashboard',
    'canvas-api',
    'dashboard/dashboard-coordinator',
    'core/app-state-manager',
    'panels/base-panel',
    'panels/panel-factory',
    'panels/system-architecture-panel',
    'panels/traffic-generation-panel',
    'panels/live-metrics-panel',
    'panels/recipe-editor-panel',
    'panels/sdl-editor-panel',
  ];

  /**
   * Verifies that dead code files have been moved to the attic directory
   * and no longer exist in the active source tree.
   */
  it('should have moved dead code files to attic', () => {
    for (const mod of DEAD_MODULES) {
      const activePath = path.join(SRC_DIR, `${mod}.ts`);
      expect(fs.existsSync(activePath), `${mod}.ts should not exist in active src/`).toBe(false);
    }
  });

  /**
   * Verifies that dead code files exist in the attic directory,
   * confirming they were moved rather than deleted.
   */
  it('should have dead code files preserved in attic', () => {
    for (const mod of DEAD_MODULES) {
      const atticPath = path.join(WEB_DIR, 'attic', 'src', `${mod}.ts`);
      expect(fs.existsSync(atticPath), `${mod}.ts should exist in attic/src/`).toBe(true);
    }
  });

  /**
   * Scans all active TypeScript source files for imports that reference
   * dead modules. Any such import would cause build failures or indicate
   * incomplete cleanup.
   */
  it('should have no stale imports in active source files', () => {
    const tsFiles = collectTsFiles(SRC_DIR);
    const staleImports: string[] = [];

    for (const file of tsFiles) {
      const content = fs.readFileSync(file, 'utf-8');
      for (const mod of DEAD_MODULES) {
        const baseName = path.basename(mod);
        // Check for import from './dead-module' or '../dead-module' patterns
        if (content.includes(`from '`) && content.includes(baseName)) {
          // More precise check: look for actual import statements
          const importRegex = new RegExp(`from\\s+['"][^'"]*${baseName}['"]`, 'g');
          if (importRegex.test(content)) {
            staleImports.push(`${path.relative(SRC_DIR, file)} imports ${baseName}`);
          }
        }
      }
    }

    expect(staleImports, `Found stale imports:\n${staleImports.join('\n')}`).toEqual([]);
  });
});

describe('Phase 1: Main Entry Point', () => {
  /**
   * Verifies that main.ts does not import any dead modules.
   * The main entry point should only reference active code paths.
   */
  it('should not import dead modules in main.ts', () => {
    const mainPath = path.join(SRC_DIR, 'main.ts');
    const content = fs.readFileSync(mainPath, 'utf-8');

    // Should NOT import system-details-page, dashboard.ts, dashboard-coordinator
    expect(content).not.toContain('system-details-page');
    expect(content).not.toContain('./dashboard');
    expect(content).not.toContain('dashboard-coordinator');
  });

  /**
   * Verifies that main.ts imports the active CanvasViewerPage component
   * which is the target architecture for the workspace IDE.
   */
  it('should import CanvasViewerPageDockView', () => {
    const mainPath = path.join(SRC_DIR, 'main.ts');
    const content = fs.readFileSync(mainPath, 'utf-8');

    expect(content).toContain('CanvasViewerPageDockView');
  });
});

describe('Phase 1: DockView Theme', () => {
  /**
   * Verifies that the CanvasViewerPageDockView applies a dockview theme class
   * to the container element. Without this, dockview panels render unstyled.
   * Reference: system-details-page.ts (now in attic) had this correctly.
   */
  it('should apply dockview theme class in CanvasViewerPageDockView', () => {
    const dockviewPath = path.join(SRC_DIR, 'pages/CanvasViewerPage/CanvasViewerPageDockView.ts');
    const content = fs.readFileSync(dockviewPath, 'utf-8');

    expect(content).toContain('dockview-theme-dark');
    expect(content).toContain('dockview-theme-light');
  });

  /**
   * Verifies that the global CSS file defines dockview theme variables.
   * These were previously only in SystemDetailsPage.html (inline styles)
   * and need to be in the global CSS for all pages to use.
   */
  it('should have dockview theme CSS variables in style.css', () => {
    const stylePath = path.join(SRC_DIR, 'style.css');
    const content = fs.readFileSync(stylePath, 'utf-8');

    expect(content).toContain('.dockview-theme-dark');
    expect(content).toContain('.dockview-theme-light');
    expect(content).toContain('--dv-background-color');
    expect(content).toContain('--dv-activeTab-background');
  });
});

describe('Phase 1: Template Integrity', () => {
  /**
   * Verifies that the CanvasViewerPage template sets window.sdlPageData
   * in the PostBodySection block. This is required for main.ts to detect
   * the page type and initialize the correct JavaScript component.
   */
  it('should set window.sdlPageData in CanvasViewerPage template', () => {
    const templatePath = path.join(WEB_DIR, 'templates/canvases/CanvasViewerPage.html');
    const content = fs.readFileSync(templatePath, 'utf-8');

    expect(content).toContain('window.sdlPageData');
    expect(content).toContain('canvas-dashboard');
    expect(content).toContain('PostBodySection');
  });

  /**
   * Verifies that SystemDetailsPage.html has been moved to attic.
   * This template is no longer served — the unified workspace page
   * replaces it.
   */
  it('should have SystemDetailsPage.html in attic', () => {
    const activePath = path.join(WEB_DIR, 'templates/systems/SystemDetailsPage.html');
    const atticPath = path.join(WEB_DIR, 'attic/templates/systems/SystemDetailsPage.html');

    expect(fs.existsSync(activePath)).toBe(false);
    expect(fs.existsSync(atticPath)).toBe(true);
  });
});

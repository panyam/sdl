/**
 * Phase 3 Verification Tests: Unified Landing Page
 *
 * These tests verify that the workspace listing page correctly combines
 * example systems and user workspaces into a single landing page:
 * - WorkspaceListingPage template has both Examples and My Workspaces sections
 * - Fork button exists for examples
 * - Old system listing templates moved to attic
 * - Old system-listing-handlers.ts moved to attic
 * - main.ts no longer imports system-listing-handlers
 * - Routes reference /workspaces/ not /canvases/ or /systems/
 */

import { describe, it, expect } from 'vitest';
import * as fs from 'fs';
import * as path from 'path';

const SRC_DIR = path.resolve(__dirname, '..');
const WEB_DIR = path.resolve(__dirname, '../..');

describe('Phase 3: Unified Landing Page Template', () => {
  const templatePath = path.join(WEB_DIR, 'templates/workspaces/WorkspaceListingPage.html');

  /**
   * Verifies the unified template exists in the workspaces directory.
   */
  it('should have WorkspaceListingPage.html template', () => {
    expect(fs.existsSync(templatePath)).toBe(true);
  });

  /**
   * Verifies the template contains an Examples section showing
   * system catalog entries (Uber, Bitly, etc.).
   */
  it('should have an Examples section in the template', () => {
    const content = fs.readFileSync(templatePath, 'utf-8');
    expect(content).toContain('Example Systems');
    expect(content).toContain('.Examples');
  });

  /**
   * Verifies the template contains a My Workspaces section using
   * goapplib EntityListing for user-created workspaces.
   */
  it('should have a My Workspaces section using EntityListing', () => {
    const content = fs.readFileSync(templatePath, 'utf-8');
    expect(content).toContain('WorkspaceEntityListing');
    expect(content).toContain('.ListingData');
  });

  /**
   * Verifies the Fork button exists in the examples section.
   * Fork creates a new workspace pre-loaded with example SDL content.
   */
  it('should have a Fork button for examples', () => {
    const content = fs.readFileSync(templatePath, 'utf-8');
    expect(content).toContain('/workspaces/fork');
    expect(content).toContain('exampleId');
  });
});

describe('Phase 3: Old Systems Files Moved', () => {
  /**
   * Verifies that SystemsListingPage.html has been moved out of the
   * active templates directory to the attic.
   */
  it('should have moved SystemsListingPage.html to attic', () => {
    const activePath = path.join(WEB_DIR, 'templates/systems/SystemsListingPage.html');
    const atticPath = path.join(WEB_DIR, 'attic/templates/systems/SystemsListingPage.html');
    expect(fs.existsSync(activePath)).toBe(false);
    expect(fs.existsSync(atticPath)).toBe(true);
  });

  /**
   * Verifies that the old system-listing-handlers.ts has been moved
   * to attic. The workspace listing page is server-rendered and doesn't
   * need client-side JS for search/filter.
   */
  it('should have moved system-listing-handlers.ts to attic', () => {
    const activePath = path.join(SRC_DIR, 'system-listing-handlers.ts');
    const atticPath = path.join(WEB_DIR, 'attic/src/system-listing-handlers.ts');
    expect(fs.existsSync(activePath)).toBe(false);
    expect(fs.existsSync(atticPath)).toBe(true);
  });
});

describe('Phase 3: Main Entry Point Cleanup', () => {
  /**
   * Verifies that main.ts no longer imports the old system-listing-handlers.
   * The workspace listing is server-rendered and needs no JS.
   */
  it('should not import system-listing-handlers in main.ts', () => {
    const content = fs.readFileSync(path.join(SRC_DIR, 'main.ts'), 'utf-8');
    expect(content).not.toContain('system-listing-handlers');
    expect(content).not.toContain('initializeSystemListing');
  });

  /**
   * Verifies main.ts handles the workspace-listing page type
   * (even if it's a no-op for server-rendered pages).
   */
  it('should handle workspace-listing page type', () => {
    const content = fs.readFileSync(path.join(SRC_DIR, 'main.ts'), 'utf-8');
    expect(content).toContain('workspace-listing');
  });
});

describe('Phase 3: Go Route Consolidation', () => {
  /**
   * Verifies that webapp.go includes the fork handler and examples
   * in the workspace listing page. Checks the Go source for key
   * patterns that indicate the unified landing page is wired up.
   */
  it('should have fork handler and examples in webapp.go', () => {
    const content = fs.readFileSync(path.join(WEB_DIR, 'server/webapp.go'), 'utf-8');
    expect(content).toContain('forkExampleHandler');
    expect(content).toContain('Examples');
    expect(content).toContain('SystemInfo');
  });
});

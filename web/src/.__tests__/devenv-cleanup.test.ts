/**
 * DevEnv cleanup verification tests (#34 Phase 3).
 *
 * These tests verify that the old CanvasDashboardPage callback pattern has been
 * fully removed from active code. The WASM pipeline now uses DevEnvPage exclusively
 * for push updates from DevEnv to the browser.
 *
 * If any of these tests fail, it means dead CanvasDashboardPage code has been
 * reintroduced. The correct pattern is:
 * - DevEnv pushes via DevEnvPageHandler interface
 * - BrowserDevEnvPage (cmd/wasm/browser.go) forwards to DevEnvPageClient
 * - Browser registers as 'DevEnvPage' service, not 'CanvasDashboardPage'
 */

import * as fs from 'fs';
import * as path from 'path';

const WEB_DIR = path.join(__dirname, '..', '..');
const SRC_DIR = path.join(WEB_DIR, 'src');
const CMD_WASM_DIR = path.join(WEB_DIR, '..', 'cmd', 'wasm');

describe('DevEnv Cleanup: No CanvasDashboardPage in active code', () => {
  /**
   * Verifies that WorkspaceViewerPageBase does not import CanvasDashboardPageMethods.
   * The old interface is replaced by WorkspacePageMethods which uses CRUD-by-name
   * instead of bulk SetGeneratorList/SetMetricsList.
   */
  it('should not import CanvasDashboardPageMethods in WorkspaceViewerPageBase', () => {
    const basePath = path.join(SRC_DIR, 'pages/WorkspaceViewerPage/WorkspaceViewerPageBase.ts');
    const content = fs.readFileSync(basePath, 'utf-8');

    expect(content).not.toContain('CanvasDashboardPageMethods');
    expect(content).not.toContain('canvasDashboardPageClient');
  });

  /**
   * Verifies that WorkspaceViewerPageBase implements WorkspacePageMethods (not
   * CanvasDashboardPageMethods). This ensures the browser receives push
   * updates through the new typed panel interface.
   */
  it('should implement WorkspacePageMethods in WorkspaceViewerPageBase', () => {
    const basePath = path.join(SRC_DIR, 'pages/WorkspaceViewerPage/WorkspaceViewerPageBase.ts');
    const content = fs.readFileSync(basePath, 'utf-8');

    expect(content).toContain('WorkspacePageMethods');
    expect(content).toContain("registerBrowserService('DevEnvPage'");
  });

  /**
   * Verifies that WorkspaceViewerPageBase does not register as CanvasDashboardPage
   * browser service. Only DevEnvPage registration should exist.
   */
  it('should not register as CanvasDashboardPage browser service', () => {
    const basePath = path.join(SRC_DIR, 'pages/WorkspaceViewerPage/WorkspaceViewerPageBase.ts');
    const content = fs.readFileSync(basePath, 'utf-8');

    expect(content).not.toContain("registerBrowserService('CanvasDashboardPage'");
  });

  /**
   * Verifies that WorkspaceViewerPageBase does not use CanvasServiceClient.
   * All operations go through CanvasViewPresenterClient which delegates to DevEnv.
   */
  it('should not import CanvasServiceClient in WorkspaceViewerPageBase', () => {
    const basePath = path.join(SRC_DIR, 'pages/WorkspaceViewerPage/WorkspaceViewerPageBase.ts');
    const content = fs.readFileSync(basePath, 'utf-8');

    expect(content).not.toContain('CanvasServiceClient');
  });

  /**
   * Verifies that cmd/wasm/main.go does not create a CanvasDashboardPageClient.
   * The WASM pipeline uses DevEnvPageClient exclusively.
   */
  it('should not create CanvasDashboardPageClient in WASM main.go', () => {
    const mainPath = path.join(CMD_WASM_DIR, 'main.go');
    const content = fs.readFileSync(mainPath, 'utf-8');

    expect(content).not.toContain('CanvasDashboardPageClient');
    expect(content).not.toContain('CanvasDashboardPage');
  });

  /**
   * Verifies that cmd/wasm/browser.go uses the BrowserDevEnvPage naming convention
   * (matching lilbattle's Browser*Panel pattern) rather than the old Forwarder name.
   */
  it('should use BrowserDevEnvPage naming in browser.go', () => {
    const browserPath = path.join(CMD_WASM_DIR, 'browser.go');
    const content = fs.readFileSync(browserPath, 'utf-8');

    expect(content).toContain('BrowserWorkspacePage');
    expect(content).not.toContain('DevEnvPageForwarder');
  });
});

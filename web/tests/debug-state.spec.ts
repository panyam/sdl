import { test, expect } from '@playwright/test';

test.describe('Debug Dashboard State', () => {
  test('check current dashboard state', async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');
    
    // Take screenshot first
    await page.screenshot({ 
      path: 'tests/screenshots/debug-current-state.png', 
      fullPage: true 
    });
    
    // Check connection status
    const connectionStatus = await page.locator('.text-sm.text-gray-400').first().textContent();
    console.log('Connection status:', connectionStatus);
    
    // Check what's in the system architecture panel
    const systemPanel = page.locator('.panel').filter({ hasText: 'System Architecture' });
    const systemContent = await systemPanel.locator('div').nth(1).textContent();
    console.log('System architecture content:', systemContent);
    
    // Check if there's already a system loaded
    const systemName = await page.locator('text=System:').locator('..').textContent();
    console.log('Current system info:', systemName);
    
    // Test the API directly
    const canvasState = await page.evaluate(async () => {
      const response = await fetch('/api/canvas/state');
      return response.json();
    });
    console.log('Canvas state:', JSON.stringify(canvasState, null, 2));
    
    const diagramResponse = await page.evaluate(async () => {
      const response = await fetch('/api/canvas/diagram');
      return response.json();
    });
    console.log('Diagram response:', JSON.stringify(diagramResponse, null, 2));
  });
});
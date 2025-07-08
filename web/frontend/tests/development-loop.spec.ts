import { test, expect } from '@playwright/test';

test.describe('Development Inner Loop', () => {
  test('dashboard development workflow', async ({ page }) => {
    // Navigate to dashboard
    await page.goto('/');
    await page.waitForLoadState('networkidle');
    
    // Take initial screenshot
    await page.screenshot({ 
      path: 'tests/screenshots/dev-loop-initial.png', 
      fullPage: true 
    });
    
    // Verify initial empty state
    await expect(page.locator('text=No System Loaded')).toBeVisible();
    await expect(page.locator('text=No Parameters Available')).toBeVisible();
    await expect(page.locator('text=No Traffic Generators')).toBeVisible();
    await expect(page.locator('text=No Metrics Available')).toBeVisible();
    
    console.log('✅ Initial state validated');
    
    // Load contacts system via API
    const loadResponse = await page.evaluate(async () => {
      const response = await fetch('/api/load', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ filePath: 'examples/contacts/contacts.sdl' })
      });
      return response.json();
    });
    expect(loadResponse.success).toBe(true);
    console.log('✅ System loaded successfully');
    
    // Activate system
    const useResponse = await page.evaluate(async () => {
      const response = await fetch('/api/use', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ systemName: 'ContactsSystem' })
      });
      return response.json();
    });
    expect(useResponse.success).toBe(true);
    console.log('✅ System activated successfully');
    
    // Wait for WebSocket updates and system to be displayed
    await page.waitForTimeout(2000); // Allow time for WebSocket messages
    
    // Check system is displayed in header
    await expect(page.locator('text=ContactsSystem')).toBeVisible();
    console.log('✅ System name appears in header');
    
    // Check system architecture panel is populated
    const systemArchPanel = page.locator('.panel').filter({ hasText: 'System Architecture' });
    await expect(systemArchPanel).toBeVisible();
    
    // Look for component names in the system architecture
    await expect(page.locator('text=server')).toBeVisible();
    await expect(page.locator('text=database')).toBeVisible();
    console.log('✅ System components visible');
    
    // Check that "No System Loaded" is gone
    await expect(page.locator('text=No System Loaded')).not.toBeVisible();
    console.log('✅ Empty state messages cleared');
    
    // Take final screenshot with system loaded
    await page.screenshot({ 
      path: 'tests/screenshots/dev-loop-with-system.png', 
      fullPage: true 
    });
    
    // Test API endpoints work
    const diagramResponse = await page.evaluate(async () => {
      const response = await fetch('/api/canvas/diagram', {
        method: 'GET',
        headers: { 'Content-Type': 'application/json' }
      });
      return response.json();
    });
    expect(diagramResponse.success).toBe(true);
    expect(diagramResponse.data.systemName).toBe('ContactsSystem');
    console.log('✅ Diagram API returns system data');
    
    // Test parameter setting
    const setResponse = await page.evaluate(async () => {
      const response = await fetch('/api/set', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ path: 'server.pool.ArrivalRate', value: 5.0 })
      });
      return response.json();
    });
    expect(setResponse.success).toBe(true);
    console.log('✅ Parameter setting works');
    
    console.log('🎉 Development loop test completed successfully!');
  });
  
  test('quick dashboard validation', async ({ page }) => {
    // Minimal test for rapid iteration
    await page.goto('/');
    
    // Just check basic structure loads
    await expect(page.locator('h1')).toContainText('SDL Canvas Dashboard');
    await expect(page.locator('.panel').first()).toBeVisible();
    
    // Take a quick screenshot
    await page.screenshot({ 
      path: 'tests/screenshots/dev-quick-check.png'
    });
    
    console.log('✅ Quick validation passed');
  });
});
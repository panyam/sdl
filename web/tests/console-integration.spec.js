import { test, expect } from '@playwright/test';

test.describe('SDL Console and Dashboard Integration', () => {
  test('dashboard loads with empty state initially', async ({ page }) => {
    await page.goto('/');
    
    // Check connection status
    await expect(page.locator('text=Connected')).toBeVisible();
    
    // Check empty state messages
    await expect(page.locator('text=No System Loaded')).toBeVisible();
    await expect(page.locator('text=No Parameters Available')).toBeVisible();
    await expect(page.locator('text=No Traffic Generators')).toBeVisible();
    await expect(page.locator('text=No Metrics Available')).toBeVisible();
  });

  test('loading system via API updates dashboard', async ({ page }) => {
    await page.goto('/');
    
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
    
    // Wait for system to load and check dashboard updates
    await expect(page.locator('text=ContactsSystem')).toBeVisible();
    
    // Check that component boxes are visible
    await expect(page.locator('text=server')).toBeVisible();
    await expect(page.locator('text=database')).toBeVisible();
    await expect(page.locator('text=idx')).toBeVisible();
    
    // Check component types
    await expect(page.locator('text=ContactAppServer')).toBeVisible();
    await expect(page.locator('text=ContactDatabase')).toBeVisible();
    await expect(page.locator('text=HashIndex')).toBeVisible();
  });

  test('setting parameters via API updates dashboard', async ({ page }) => {
    await page.goto('/');
    
    // Load and activate system
    await page.evaluate(async () => {
      await fetch('/api/load', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ filePath: 'examples/contacts/contacts.sdl' })
      });
      await fetch('/api/use', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ systemName: 'ContactsSystem' })
      });
    });
    
    // Wait for system to load
    await expect(page.locator('text=ContactsSystem')).toBeVisible();
    
    // Set a parameter
    const setResponse = await page.evaluate(async () => {
      const response = await fetch('/api/set', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ 
          path: 'server.pool.ArrivalRate', 
          value: 15.5 
        })
      });
      return response.json();
    });
    expect(setResponse.success).toBe(true);
    
    // Check that Canvas state includes the parameter change
    const stateResponse = await page.evaluate(async () => {
      const response = await fetch('/api/canvas/state');
      return response.json();
    });
    expect(stateResponse.success).toBe(true);
    expect(stateResponse.data.systemParameters).toEqual({
      'server.pool.ArrivalRate': 15.5
    });
  });

  test('canvas state persists across page reloads', async ({ page }) => {
    await page.goto('/');
    
    // Load system and set parameters
    await page.evaluate(async () => {
      await fetch('/api/load', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ filePath: 'examples/contacts/contacts.sdl' })
      });
      await fetch('/api/use', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ systemName: 'ContactsSystem' })
      });
      await fetch('/api/set', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ 
          path: 'server.pool.ArrivalRate', 
          value: 25 
        })
      });
    });
    
    // Wait for system to load
    await expect(page.locator('text=ContactsSystem')).toBeVisible();
    
    // Reload the page
    await page.reload();
    
    // Check that system is still loaded
    await expect(page.locator('text=ContactsSystem')).toBeVisible();
    await expect(page.locator('text=server')).toBeVisible();
    
    // Check that Canvas state still includes the parameter
    const stateResponse = await page.evaluate(async () => {
      const response = await fetch('/api/canvas/state');
      return response.json();
    });
    expect(stateResponse.success).toBe(true);
    expect(stateResponse.data.activeSystem).toBe('ContactsSystem');
    expect(stateResponse.data.systemParameters['server.pool.ArrivalRate']).toBe(25);
  });

  test('diagram API returns correct system topology', async ({ page }) => {
    await page.goto('/');
    
    // Load and activate system
    await page.evaluate(async () => {
      await fetch('/api/load', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ filePath: 'examples/contacts/contacts.sdl' })
      });
      await fetch('/api/use', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ systemName: 'ContactsSystem' })
      });
    });
    
    // Get diagram
    const diagramResponse = await page.evaluate(async () => {
      const response = await fetch('/api/canvas/diagram');
      return response.json();
    });
    
    expect(diagramResponse.success).toBe(true);
    expect(diagramResponse.data.systemName).toBe('ContactsSystem');
    expect(diagramResponse.data.nodes).toHaveLength(3);
    expect(diagramResponse.data.edges).toHaveLength(2);
    
    // Check specific nodes
    const nodeNames = diagramResponse.data.nodes.map(n => n.Name);
    expect(nodeNames).toContain('server');
    expect(nodeNames).toContain('database');
    expect(nodeNames).toContain('idx');
    
    // Check node types
    const nodeTypes = diagramResponse.data.nodes.map(n => n.Type);
    expect(nodeTypes).toContain('ContactAppServer');
    expect(nodeTypes).toContain('ContactDatabase');
    expect(nodeTypes).toContain('HashIndex');
  });

  test('websocket connection broadcasts events', async ({ page }) => {
    await page.goto('/');
    
    // Listen for WebSocket messages
    const messages = [];
    page.on('console', msg => {
      if (msg.text().includes('📡 WebSocket message:')) {
        messages.push(msg.text());
      }
    });
    
    // Load system via API
    await page.evaluate(async () => {
      await fetch('/api/load', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ filePath: 'examples/contacts/contacts.sdl' })
      });
    });
    
    // Wait for WebSocket message
    await page.waitForTimeout(1000);
    
    // Check that fileLoaded event was received
    const fileLoadedMsg = messages.find(msg => msg.includes('fileLoaded'));
    expect(fileLoadedMsg).toBeTruthy();
  });
});
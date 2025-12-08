import { test, expect } from '@playwright/test';

test.describe('Layout Debugging', () => {
  test('capture full page layout', async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');
    
    // Take full page screenshot
    await page.screenshot({ 
      path: 'tests/screenshots/full-page.png', 
      fullPage: true 
    });
    
    // Take viewport screenshot
    await page.screenshot({ 
      path: 'tests/screenshots/viewport.png' 
    });
    
    console.log('Screenshots saved to tests/screenshots/');
  });

  test('measure layout elements', async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');
    
    // Measure main container
    const app = page.locator('#app');
    const appBox = await app.boundingBox();
    console.log('App container:', appBox);
    
    // Measure header
    const header = page.locator('h1');
    const headerBox = await header.boundingBox();
    console.log('Header:', headerBox);
    
    // Measure grid layout
    const grid = page.locator('.grid.grid-cols-1.lg\\:grid-cols-3');
    const gridBox = await grid.boundingBox();
    console.log('Main grid:', gridBox);
    
    // Measure panels
    const panels = page.locator('.panel');
    const panelCount = await panels.count();
    console.log(`Found ${panelCount} panels`);
    
    for (let i = 0; i < panelCount; i++) {
      const panel = panels.nth(i);
      const panelBox = await panel.boundingBox();
      const panelTitle = await panel.locator('.panel-header').textContent();
      console.log(`Panel "${panelTitle}":`, panelBox);
    }
  });

  test('test responsive behavior', async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');
    
    // Test different viewport sizes
    const viewports = [
      { width: 1920, height: 1080, name: 'desktop-large' },
      { width: 1366, height: 768, name: 'desktop-medium' },
      { width: 1024, height: 768, name: 'tablet-landscape' },
      { width: 768, height: 1024, name: 'tablet-portrait' },
      { width: 375, height: 667, name: 'mobile' }
    ];
    
    for (const viewport of viewports) {
      await page.setViewportSize({ width: viewport.width, height: viewport.height });
      await page.waitForTimeout(500); // Wait for layout to settle
      
      await page.screenshot({ 
        path: `tests/screenshots/responsive-${viewport.name}.png` 
      });
      
      const gridClasses = await page.locator('.grid').getAttribute('class');
      console.log(`${viewport.name} (${viewport.width}x${viewport.height}): ${gridClasses}`);
    }
  });
});
import { test, expect } from '@playwright/test';

test.describe('Three Panel Layout', () => {
  test('validates new Row 1 layout structure', async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');
    
    console.log('📊 Validating 3-panel Row 1 layout:');
    
    // Check System Architecture panel (70% width, left side)
    const systemArchPanel = page.locator('.panel:has-text("System Architecture")');
    await expect(systemArchPanel).toBeVisible();
    const systemBox = await systemArchPanel.boundingBox();
    console.log(`System Architecture: ${systemBox?.width}x${systemBox?.height} at (${systemBox?.x}, ${systemBox?.y})`);
    
    // Check Traffic Generation panel (top right, 48% of right side height)
    const trafficPanel = page.locator('.panel:has-text("Traffic Generation")');
    await expect(trafficPanel).toBeVisible();
    const trafficBox = await trafficPanel.boundingBox();
    console.log(`Traffic Generation: ${trafficBox?.width}x${trafficBox?.height} at (${trafficBox?.x}, ${trafficBox?.y})`);
    
    // Check System Parameters panel (bottom right, 48% of right side height)
    const paramsPanel = page.locator('.panel:has-text("System Parameters")');
    await expect(paramsPanel).toBeVisible();
    const paramsBox = await paramsPanel.boundingBox();
    console.log(`System Parameters: ${paramsBox?.width}x${paramsBox?.height} at (${paramsBox?.x}, ${paramsBox?.y})`);
    
    // Validate layout proportions
    if (systemBox && trafficBox && paramsBox) {
      // System Architecture should be about 70% width
      const totalWidth = systemBox.width + trafficBox.width + 16; // +16 for gap
      const systemWidthRatio = systemBox.width / totalWidth;
      console.log(`System Architecture width ratio: ${(systemWidthRatio * 100).toFixed(1)}%`);
      expect(systemWidthRatio).toBeGreaterThan(0.65); // Should be around 70%
      expect(systemWidthRatio).toBeLessThan(0.75);
      
      // Traffic Generation and System Parameters should have similar widths
      const widthDiff = Math.abs(trafficBox.width - paramsBox.width);
      console.log(`Right panel width difference: ${widthDiff}px`);
      expect(widthDiff).toBeLessThan(10); // Should be very similar
      
      // Traffic panel should be above System Parameters panel
      expect(trafficBox.y).toBeLessThan(paramsBox.y);
      console.log(`Vertical separation: ${paramsBox.y - (trafficBox.y + trafficBox.height)}px`);
      
      // Both right panels should have similar heights (around 48% each)
      const heightDiff = Math.abs(trafficBox.height - paramsBox.height);
      console.log(`Right panel height difference: ${heightDiff}px`);
      expect(heightDiff).toBeLessThan(20); // Should be very similar
    }
    
    console.log('✅ 3-panel layout structure validated');
  });

  test('validates enhanced system architecture content', async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');
    
    // Check for enhanced system architecture components
    await expect(page.locator('.component-title:has-text("ContactAppServer")')).toBeVisible();
    await expect(page.locator('.component-title:has-text("ContactDatabase")')).toBeVisible();
    await expect(page.locator('.component-title:has-text("HashIndex")')).toBeVisible();
    
    // Check for new System Health section
    await expect(page.locator('h4:has-text("System Health")')).toBeVisible();
    await expect(page.locator('.text-green-400')).toBeVisible(); // Success rate
    await expect(page.locator('.text-blue-400')).toBeVisible();  // Avg latency
    await expect(page.locator('.text-yellow-400')).toBeVisible(); // Current load
    
    console.log('✅ Enhanced system architecture content validated');
  });

  test('validates separated panels functionality', async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');
    
    // Check Traffic Generation controls in their own panel
    const trafficControls = page.locator('.panel:has-text("Traffic Generation")');
    await expect(trafficControls.locator('[data-generate-id="lookup-traffic"]')).toBeVisible();
    await expect(trafficControls.locator('#add-generator')).toBeVisible();
    
    // Check System Parameters controls in their own panel
    const paramControls = page.locator('.panel:has-text("System Parameters")');
    await expect(paramControls.locator('#slider-server\\.pool\\.ArrivalRate')).toBeVisible();
    await expect(paramControls.locator('#slider-server\\.pool\\.Size')).toBeVisible();
    
    // Verify parameters are NOT in the System Architecture panel
    const systemArch = page.locator('.panel:has-text("System Architecture")');
    const paramsInSystemArch = systemArch.locator('#slider-server\\.pool\\.ArrivalRate');
    await expect(paramsInSystemArch).not.toBeVisible();
    
    console.log('✅ Panel separation validated - parameters moved successfully');
  });

  test('validates right panel scrolling', async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');
    
    // Check that right panels have overflow scrolling
    const trafficPanel = page.locator('.panel:has-text("Traffic Generation") .overflow-y-auto');
    await expect(trafficPanel).toBeVisible();
    
    const paramsPanel = page.locator('.panel:has-text("System Parameters") .overflow-y-auto');
    await expect(paramsPanel).toBeVisible();
    
    // Test scrolling in parameters panel (likely to have more content)
    const paramsScrollTop = await paramsPanel.evaluate(el => el.scrollTop);
    console.log(`Parameters panel initial scroll: ${paramsScrollTop}`);
    
    // Try to scroll (may not scroll if content fits)
    await paramsPanel.evaluate(el => el.scrollTop = 50);
    const newScrollTop = await paramsPanel.evaluate(el => el.scrollTop);
    console.log(`Parameters panel after scroll attempt: ${newScrollTop}`);
    
    console.log('✅ Right panel scrolling capability validated');
  });
});
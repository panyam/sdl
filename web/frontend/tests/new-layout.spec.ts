import { test, expect } from '@playwright/test';

test.describe('New 2-Row Layout', () => {
  test('validates new layout structure', async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');
    
    // Verify header exists
    await expect(page.locator('h1:has-text("SDL Canvas Dashboard")')).toBeVisible();
    
    // Row 1: System Architecture + Traffic Generation (50% height)
    const systemArchPanel = page.locator('.panel:has-text("System Architecture")');
    await expect(systemArchPanel).toBeVisible();
    
    const trafficPanel = page.locator('.panel:has-text("Traffic Generation")');  
    await expect(trafficPanel).toBeVisible();
    
    // Row 2: Live Metrics with dynamic charts (50% height)
    const metricsPanel = page.locator('.panel:has-text("Live Metrics")');
    await expect(metricsPanel).toBeVisible();
    
    // Verify dynamic charts are present
    await expect(page.locator('#chart-server-latency')).toBeVisible();
    await expect(page.locator('#chart-server-qps')).toBeVisible(); 
    await expect(page.locator('#chart-db-latency')).toBeVisible();
    
    console.log('✅ New 2-row layout validated successfully');
  });

  test('verifies traffic generation controls', async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');
    
    // Check for traffic generator checkboxes
    const lookupCheckbox = page.locator('[data-generate-id="lookup-traffic"]');
    await expect(lookupCheckbox).toBeVisible();
    
    const bulkCheckbox = page.locator('[data-generate-id="bulk-traffic"]');
    await expect(bulkCheckbox).toBeVisible();
    
    // Check for rate sliders  
    const lookupSlider = page.locator('[data-rate-id="lookup-traffic"]');
    await expect(lookupSlider).toBeVisible();
    
    // Check add generator button
    await expect(page.locator('#add-generator')).toBeVisible();
    
    console.log('✅ Traffic generation controls validated');
  });

  test('verifies enhanced system architecture', async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');
    
    // Enhanced system components should be visible
    await expect(page.locator('.component-box:has-text("ContactAppServer")')).toBeVisible();
    await expect(page.locator('.component-box:has-text("ContactDatabase")')).toBeVisible();
    await expect(page.locator('.component-box:has-text("HashIndex")')).toBeVisible();
    
    // Parameter controls should be integrated
    await expect(page.locator('h3:has-text("System Parameters")')).toBeVisible();
    await expect(page.locator('#slider-server\\.pool\\.ArrivalRate')).toBeVisible();
    
    console.log('✅ Enhanced system architecture validated');
  });
});
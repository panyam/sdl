import { test, expect } from '@playwright/test';

test.describe('SDL Dashboard', () => {
  test('loads dashboard successfully', async ({ page }) => {
    await page.goto('/');
    await expect(page).toHaveTitle(/SDL Canvas/);
  });

  test('displays main dashboard components', async ({ page }) => {
    await page.goto('/');
    
    // Wait for page to load
    await page.waitForLoadState('networkidle');
    
    // Wait for main dashboard elements to be visible
    await expect(page.locator('h1')).toBeVisible();
    await expect(page.locator('#load-btn')).toBeVisible();
    await expect(page.locator('#run-btn')).toBeVisible();
    await expect(page.locator('#latency-chart')).toBeVisible();
    
    // Check for main panels using class selectors
    await expect(page.locator('.panel').first()).toBeVisible();
    await expect(page.locator('.component-box').first()).toBeVisible();
  });

  test('interactive layout testing', async ({ page }) => {
    await page.goto('/');
    
    // Take a screenshot for layout inspection
    await page.screenshot({ path: 'tests/screenshots/dashboard-layout.png', fullPage: true });
    
    // Get viewport dimensions
    const viewportSize = page.viewportSize();
    console.log('Viewport:', viewportSize);
    
    // Get main container dimensions
    const container = page.locator('body');
    const containerBox = await container.boundingBox();
    console.log('Container dimensions:', containerBox);
  });
});
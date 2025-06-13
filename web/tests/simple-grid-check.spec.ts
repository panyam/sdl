import { test, expect } from '@playwright/test';

test.describe('Simple Grid Check', () => {
  test('check grid structure', async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');
    
    // Check the metrics grid specifically
    const metricsGrid = page.locator('.panel:has-text("Live Metrics") .grid');
    await expect(metricsGrid).toBeVisible();
    
    const gridBox = await metricsGrid.boundingBox();
    console.log(`Grid dimensions: ${gridBox?.width}x${gridBox?.height}`);
    
    // Count direct children of the grid
    const gridChildren = metricsGrid.locator('> div');
    const childCount = await gridChildren.count();
    console.log(`Direct grid children: ${childCount}`);
    
    // Check each chart specifically by ID
    const expectedCharts = [
      'chart-server-latency',
      'chart-server-qps', 
      'chart-db-latency',
      'chart-server-errors',
      'chart-db-qps',
      'chart-cache-hit',
      'chart-server-p99',
      'chart-network-util',
      'chart-memory-heap'
    ];
    
    console.log('📊 Checking specific charts:');
    let visibleCharts = 0;
    for (const chartId of expectedCharts) {
      const chart = page.locator(`#${chartId}`);
      const isVisible = await chart.isVisible();
      if (isVisible) {
        const box = await chart.boundingBox();
        console.log(`✅ ${chartId}: (${box?.x}, ${box?.y}) ${box?.width}x${box?.height}`);
        visibleCharts++;
      } else {
        console.log(`❌ ${chartId}: NOT VISIBLE`);
      }
    }
    
    console.log(`Total visible charts: ${visibleCharts}/${expectedCharts.length}`);
    
    // Take a screenshot to inspect visually
    await page.screenshot({ 
      path: 'tests/screenshots/simple-grid-check.png',
      fullPage: true 
    });
    
    // Check if scrolling is working by scrolling down a bit
    const scrollContainer = page.locator('.panel:has-text("Live Metrics") .overflow-y-auto');
    await scrollContainer.evaluate(el => el.scrollTop = 100);
    
    await page.screenshot({ 
      path: 'tests/screenshots/simple-grid-scrolled.png',
      fullPage: true 
    });
    
    console.log('✅ Grid structure check completed');
  });
});
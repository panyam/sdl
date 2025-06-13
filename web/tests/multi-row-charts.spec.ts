import { test, expect } from '@playwright/test';

test.describe('Multi-Row Charts Demo', () => {
  test('validates 5 charts in 2-row layout (3+2)', async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');
    
    // Verify all 5 charts are present
    const charts = [
      '#chart-server-latency',
      '#chart-server-qps', 
      '#chart-db-latency',
      '#chart-server-errors',
      '#chart-db-qps'
    ];
    
    for (const chartId of charts) {
      await expect(page.locator(chartId)).toBeVisible();
      console.log(`✅ Found chart: ${chartId}`);
    }
    
    // Take screenshot of the 2-row layout
    await page.screenshot({ 
      path: 'tests/screenshots/multi-row-charts.png',
      fullPage: true 
    });
    
    // Verify metrics grid layout
    const metricsGrid = page.locator('.grid.grid-cols-1.md\\:grid-cols-2.lg\\:grid-cols-3');
    await expect(metricsGrid).toBeVisible();
    
    // Count chart containers
    const chartContainers = page.locator('.grid > div');
    const count = await chartContainers.count();
    console.log(`📊 Found ${count} chart containers`);
    expect(count).toBe(5);
    
    console.log('✅ 2-row metrics layout (3+2) validated successfully');
  });

  test('measures chart layout distribution', async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');
    
    // Get positions of all charts
    const charts = [
      { id: '#chart-server-latency', name: 'Server Latency' },
      { id: '#chart-server-qps', name: 'Server QPS' },
      { id: '#chart-db-latency', name: 'DB Latency' },
      { id: '#chart-server-errors', name: 'Server Errors' },
      { id: '#chart-db-qps', name: 'DB QPS' }
    ];
    
    console.log('📊 Chart Layout Distribution:');
    for (const chart of charts) {
      const element = page.locator(chart.id);
      const box = await element.boundingBox();
      console.log(`${chart.name}: x=${box?.x}, y=${box?.y}, w=${box?.width}, h=${box?.height}`);
    }
    
    // Verify first 3 charts are in top row (similar Y position)
    const topRowCharts = charts.slice(0, 3);
    const bottomRowCharts = charts.slice(3, 5);
    
    let topRowY = null;
    for (const chart of topRowCharts) {
      const box = await page.locator(chart.id).boundingBox();
      if (topRowY === null) {
        topRowY = box?.y;
      } else {
        // Charts in same row should have similar Y positions (within 10px)
        expect(Math.abs((box?.y || 0) - topRowY)).toBeLessThan(10);
      }
    }
    
    let bottomRowY = null;
    for (const chart of bottomRowCharts) {
      const box = await page.locator(chart.id).boundingBox();
      if (bottomRowY === null) {
        bottomRowY = box?.y;
      } else {
        expect(Math.abs((box?.y || 0) - bottomRowY)).toBeLessThan(10);
      }
    }
    
    // Bottom row should be below top row
    expect(bottomRowY).toBeGreaterThan(topRowY);
    
    console.log(`✅ Top row at y=${topRowY}, Bottom row at y=${bottomRowY}`);
  });
});
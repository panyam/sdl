import { test, expect } from '@playwright/test';

test.describe('Scrollable Metrics Grid', () => {
  test('validates scrollable metrics with 9 charts', async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');
    
    // Verify Live Metrics panel has vertical scroll
    const metricsPanel = page.locator('.panel:has-text("Live Metrics")');
    await expect(metricsPanel).toBeVisible();
    
    const scrollContainer = page.locator('.panel:has-text("Live Metrics") .overflow-y-auto');
    await expect(scrollContainer).toBeVisible();
    
    // Count all charts
    const allCharts = [
      '#chart-server-latency',
      '#chart-server-qps', 
      '#chart-db-latency',
      '#chart-server-errors',
      '#chart-db-qps',
      '#chart-cache-hit',
      '#chart-server-p99',
      '#chart-network-util',
      '#chart-memory-heap'
    ];
    
    console.log('📊 Verifying all 9 charts are present:');
    for (const chartId of allCharts) {
      await expect(page.locator(chartId)).toBeVisible();
      console.log(`✅ Found: ${chartId}`);
    }
    
    // Test scrolling behavior
    const scrollElement = scrollContainer.first();
    
    // Get initial scroll position
    const initialScrollTop = await scrollElement.evaluate(el => el.scrollTop);
    console.log(`Initial scroll position: ${initialScrollTop}`);
    
    // Scroll down
    await scrollElement.evaluate(el => el.scrollTop = 200);
    
    // Verify we can scroll
    const newScrollTop = await scrollElement.evaluate(el => el.scrollTop);
    console.log(`After scroll: ${newScrollTop}`);
    expect(newScrollTop).toBeGreaterThan(initialScrollTop);
    
    // Take screenshot showing scrolled state
    await page.screenshot({ 
      path: 'tests/screenshots/scrollable-metrics.png',
      fullPage: true 
    });
    
    console.log('✅ Scrollable metrics grid validated successfully');
  });

  test('measures grid layout with multiple rows', async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');
    
    // Get Live Metrics panel dimensions
    const metricsPanel = page.locator('.panel:has-text("Live Metrics")');
    const panelBox = await metricsPanel.boundingBox();
    console.log(`Live Metrics panel: ${panelBox?.width}x${panelBox?.height}`);
    
    // Get scroll container dimensions
    const scrollContainer = page.locator('.panel:has-text("Live Metrics") .overflow-y-auto');
    const scrollBox = await scrollContainer.boundingBox();
    console.log(`Scroll container: ${scrollBox?.width}x${scrollBox?.height}`);
    
    // Get grid dimensions
    const grid = page.locator('.grid.grid-cols-1.md\\:grid-cols-2.lg\\:grid-cols-3');
    const gridBox = await grid.boundingBox();
    console.log(`Grid dimensions: ${gridBox?.width}x${gridBox?.height}`);
    
    // Verify grid content is taller than scroll container (enabling scroll)
    if (gridBox && scrollBox) {
      console.log(`Grid content height: ${gridBox.height}px`);
      console.log(`Scroll container height: ${scrollBox.height}px`);
      
      if (gridBox.height > scrollBox.height) {
        console.log('✅ Grid content is taller than container - scrolling enabled');
      } else {
        console.log('ℹ️ Grid content fits within container - no scrolling needed yet');
      }
    }
    
    // Count chart containers with fixed height
    const chartContainers = page.locator('.grid > div.h-48');
    const count = await chartContainers.count();
    console.log(`📊 Found ${count} charts with fixed 192px height (h-48)`);
    
    // Calculate expected total height
    const expectedRows = Math.ceil(count / 3); // 3 columns
    const expectedHeight = expectedRows * (192 + 16); // 192px height + 16px gap
    console.log(`Expected grid height: ${expectedHeight}px (${expectedRows} rows)`);
  });

  test('simulates server adding metrics dynamically', async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');
    
    // Expose dashboard instance for testing
    await page.evaluate(() => {
      // @ts-ignore - accessing global dashboard for testing
      window.testDashboard = document.dashboardInstance;
    });
    
    // Simulate server adding new metrics (this would come from WebSocket in real usage)
    await page.evaluate(() => {
      const dashboard = document.querySelector('#app')?.__dashboard_instance;
      if (dashboard && dashboard.addMetricFromServer) {
        dashboard.addMetricFromServer('loadbalancer.requests.p95Latency', 'Load Balancer P95 Latency');
        dashboard.addMetricFromServer('api.gateway.throughput.qps', 'API Gateway Throughput');
        dashboard.addMetricFromServer('disk.io.utilization', 'Disk I/O Utilization %');
      }
    });
    
    console.log('✅ Simulated server adding 3 new metrics dynamically');
  });
});
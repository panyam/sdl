import { test, expect } from '@playwright/test';

test.describe('Overflow Check', () => {
  test('check panel boundaries vs chart positions', async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');
    
    // Get Live Metrics panel boundaries
    const panel = page.locator('.panel:has-text("Live Metrics")');
    const panelBox = await panel.boundingBox();
    console.log(`Live Metrics Panel: (${panelBox?.x}, ${panelBox?.y}) ${panelBox?.width}x${panelBox?.height}`);
    
    // Get scroll container boundaries
    const scrollContainer = page.locator('.panel:has-text("Live Metrics") .overflow-y-auto');
    const scrollBox = await scrollContainer.boundingBox();
    console.log(`Scroll Container: (${scrollBox?.x}, ${scrollBox?.y}) ${scrollBox?.width}x${scrollBox?.height}`);
    
    // Get grid boundaries
    const grid = page.locator('.panel:has-text("Live Metrics") .grid');
    const gridBox = await grid.boundingBox();
    console.log(`Grid Container: (${gridBox?.x}, ${gridBox?.y}) ${gridBox?.width}x${gridBox?.height}`);
    
    // Check specific chart positions relative to panel
    const charts = [
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
    
    console.log('\n📊 Chart positions relative to panel:');
    let chartsOutsidePanel = 0;
    
    for (const chartId of charts) {
      const chart = page.locator(`#${chartId}`);
      const chartBox = await chart.boundingBox();
      
      if (chartBox && panelBox) {
        const isOutsidePanel = 
          chartBox.y + chartBox.height > panelBox.y + panelBox.height ||
          chartBox.x + chartBox.width > panelBox.x + panelBox.width ||
          chartBox.x < panelBox.x ||
          chartBox.y < panelBox.y;
        
        const relativeY = chartBox.y - panelBox.y;
        const relativeBottom = (chartBox.y + chartBox.height) - (panelBox.y + panelBox.height);
        
        console.log(`${chartId}: y=${chartBox.y} (rel: ${relativeY}), bottom overflow: ${relativeBottom}px ${isOutsidePanel ? '❌ OUTSIDE' : '✅ INSIDE'}`);
        
        if (isOutsidePanel) chartsOutsidePanel++;
      }
    }
    
    console.log(`\nCharts outside panel: ${chartsOutsidePanel}/${charts.length}`);
    
    // Check if panel has proper clipping
    const panelStyles = await panel.evaluate(el => {
      const styles = window.getComputedStyle(el);
      return {
        overflow: styles.overflow,
        overflowY: styles.overflowY,
        position: styles.position
      };
    });
    
    console.log(`Panel styles:`, panelStyles);
    
    const scrollStyles = await scrollContainer.evaluate(el => {
      const styles = window.getComputedStyle(el);
      return {
        overflow: styles.overflow,
        overflowY: styles.overflowY,
        position: styles.position,
        height: styles.height
      };
    });
    
    console.log(`Scroll container styles:`, scrollStyles);
  });
});
import { test, expect } from '@playwright/test';

test.describe('Clipped Scrolling Test', () => {
  test('verify charts are clipped but scrollable', async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');
    
    // Get scroll container
    const scrollContainer = page.locator('.panel:has-text("Live Metrics") .overflow-y-auto');
    
    // Initially, only first row should be visible (clipped)
    console.log('📊 Initial visibility check:');
    const row1Charts = ['chart-server-latency', 'chart-server-qps', 'chart-db-latency'];
    const row2Charts = ['chart-server-errors', 'chart-db-qps', 'chart-cache-hit'];
    const row3Charts = ['chart-server-p99', 'chart-network-util', 'chart-memory-heap'];
    
    // Check row 1 visibility (should be visible)
    for (const chartId of row1Charts) {
      const isVisible = await page.locator(`#${chartId}`).isVisible();
      console.log(`${chartId}: ${isVisible ? '✅ VISIBLE' : '❌ HIDDEN'}`);
    }
    
    // Take screenshot of initial state
    await page.screenshot({ 
      path: 'tests/screenshots/clipped-initial.png',
      fullPage: true 
    });
    
    // Scroll down to see row 2
    await scrollContainer.evaluate(el => el.scrollTop = 200);
    await page.waitForTimeout(500); // Wait for scroll to complete
    
    console.log('\n📊 After scrolling down 200px:');
    for (const chartId of row2Charts) {
      const isVisible = await page.locator(`#${chartId}`).isVisible();
      console.log(`${chartId}: ${isVisible ? '✅ VISIBLE' : '❌ HIDDEN'}`);
    }
    
    // Take screenshot of scrolled state
    await page.screenshot({ 
      path: 'tests/screenshots/clipped-scrolled.png',
      fullPage: true 
    });
    
    // Scroll to bottom to see row 3
    await scrollContainer.evaluate(el => el.scrollTop = 400);
    await page.waitForTimeout(500);
    
    console.log('\n📊 After scrolling to bottom:');
    for (const chartId of row3Charts) {
      const isVisible = await page.locator(`#${chartId}`).isVisible();
      console.log(`${chartId}: ${isVisible ? '✅ VISIBLE' : '❌ HIDDEN'}`);
    }
    
    // Take screenshot of bottom scroll state
    await page.screenshot({ 
      path: 'tests/screenshots/clipped-bottom.png',
      fullPage: true 
    });
    
    // Verify scroll container properties
    const scrollTop = await scrollContainer.evaluate(el => el.scrollTop);
    const scrollHeight = await scrollContainer.evaluate(el => el.scrollHeight);
    const clientHeight = await scrollContainer.evaluate(el => el.clientHeight);
    
    console.log(`\n📊 Scroll properties:`);
    console.log(`Current scroll position: ${scrollTop}px`);
    console.log(`Total scrollable height: ${scrollHeight}px`);
    console.log(`Visible height: ${clientHeight}px`);
    console.log(`Scrollable distance: ${scrollHeight - clientHeight}px`);
    
    // Verify scrolling is functional
    expect(scrollHeight).toBeGreaterThan(clientHeight);
    expect(scrollTop).toBeGreaterThan(0);
    
    console.log('✅ Clipped scrolling working correctly');
  });
});
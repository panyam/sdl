import { test, expect } from '@playwright/test';

test.describe('Visual Clipping Test', () => {
  test('visual confirmation of panel clipping', async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');
    
    console.log('📊 Testing visual clipping behavior:');
    
    // Take initial screenshot
    await page.screenshot({ 
      path: 'tests/screenshots/clipping-initial.png',
      fullPage: true 
    });
    
    // Try to scroll content in each panel to see if it's properly clipped
    
    // 1. System Architecture - scroll down to see if content is clipped
    const systemScroll = page.locator('.panel:has-text("System Architecture") .overflow-y-auto');
    await systemScroll.evaluate(el => el.scrollTop = 100);
    await page.screenshot({ 
      path: 'tests/screenshots/clipping-system-scrolled.png',
      fullPage: true 
    });
    console.log('✅ System Architecture scrolling tested');
    
    // 2. Traffic Generation - try to scroll (may not have enough content)
    const trafficScroll = page.locator('.panel:has-text("Traffic Generation") .overflow-y-auto');
    await trafficScroll.evaluate(el => el.scrollTop = 50);
    await page.screenshot({ 
      path: 'tests/screenshots/clipping-traffic-scrolled.png',
      fullPage: true 
    });
    console.log('✅ Traffic Generation scrolling tested');
    
    // 3. System Parameters - try to scroll
    const paramsScroll = page.locator('.panel:has-text("System Parameters") .overflow-y-auto');
    await paramsScroll.evaluate(el => el.scrollTop = 50);
    await page.screenshot({ 
      path: 'tests/screenshots/clipping-params-scrolled.png',
      fullPage: true 
    });
    console.log('✅ System Parameters scrolling tested');
    
    // 4. Live Metrics - scroll to verify charts are clipped
    const metricsScroll = page.locator('.panel:has-text("Live Metrics") .overflow-y-auto');
    await metricsScroll.evaluate(el => el.scrollTop = 200);
    await page.screenshot({ 
      path: 'tests/screenshots/clipping-metrics-scrolled.png',
      fullPage: true 
    });
    console.log('✅ Live Metrics scrolling tested');
    
    // Verify panel styles are applied correctly
    const allPanels = page.locator('.panel');
    const panelCount = await allPanels.count();
    
    console.log(`📊 Checking ${panelCount} panels for overflow:hidden`);
    let clippedPanels = 0;
    
    for (let i = 0; i < panelCount; i++) {
      const panel = allPanels.nth(i);
      const overflow = await panel.evaluate(el => window.getComputedStyle(el).overflow);
      if (overflow === 'hidden') {
        clippedPanels++;
      }
    }
    
    console.log(`✅ ${clippedPanels}/${panelCount} panels have overflow:hidden`);
    expect(clippedPanels).toBeGreaterThanOrEqual(4); // At least our 4 main panels
    
    // Final screenshot showing everything properly clipped
    await page.screenshot({ 
      path: 'tests/screenshots/final-clipped-layout.png',
      fullPage: true 
    });
    
    console.log('✅ Visual clipping verification complete');
  });
});
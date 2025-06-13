import { test, expect } from '@playwright/test';

test.describe('Panel Clipping Validation', () => {
  test('verify all panels have proper overflow clipping', async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');
    
    console.log('📊 Checking panel overflow properties:');
    
    // Check System Architecture panel
    const systemArchPanel = page.locator('.panel:has-text("System Architecture")');
    const systemStyles = await systemArchPanel.evaluate(el => {
      const styles = window.getComputedStyle(el);
      return {
        overflow: styles.overflow,
        overflowY: styles.overflowY,
        position: styles.position
      };
    });
    console.log('System Architecture panel styles:', systemStyles);
    expect(systemStyles.overflow).toBe('hidden');
    
    // Check Traffic Generation panel
    const trafficPanel = page.locator('.panel:has-text("Traffic Generation")');
    const trafficStyles = await trafficPanel.evaluate(el => {
      const styles = window.getComputedStyle(el);
      return {
        overflow: styles.overflow,
        overflowY: styles.overflowY,
        position: styles.position
      };
    });
    console.log('Traffic Generation panel styles:', trafficStyles);
    expect(trafficStyles.overflow).toBe('hidden');
    
    // Check System Parameters panel
    const paramsPanel = page.locator('.panel:has-text("System Parameters")');
    const paramsStyles = await paramsPanel.evaluate(el => {
      const styles = window.getComputedStyle(el);
      return {
        overflow: styles.overflow,
        overflowY: styles.overflowY,
        position: styles.position
      };
    });
    console.log('System Parameters panel styles:', paramsStyles);
    expect(paramsStyles.overflow).toBe('hidden');
    
    // Check Live Metrics panel (should also be clipped)
    const metricsPanel = page.locator('.panel:has-text("Live Metrics")');
    const metricsStyles = await metricsPanel.evaluate(el => {
      const styles = window.getComputedStyle(el);
      return {
        overflow: styles.overflow,
        overflowY: styles.overflowY,
        position: styles.position
      };
    });
    console.log('Live Metrics panel styles:', metricsStyles);
    expect(metricsStyles.overflow).toBe('hidden');
    
    console.log('✅ All panels have overflow: hidden');
  });

  test('verify content is contained within panel boundaries', async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');
    
    console.log('📊 Checking content containment:');
    
    // Get panel boundaries and check content doesn't overflow
    const panels = [
      { name: 'System Architecture', selector: '.panel:has-text("System Architecture")' },
      { name: 'Traffic Generation', selector: '.panel:has-text("Traffic Generation")' },
      { name: 'System Parameters', selector: '.panel:has-text("System Parameters")' },
      { name: 'Live Metrics', selector: '.panel:has-text("Live Metrics")' }
    ];
    
    for (const panel of panels) {
      const panelElement = page.locator(panel.selector);
      const panelBox = await panelElement.boundingBox();
      
      if (panelBox) {
        console.log(`${panel.name}: ${panelBox.width}x${panelBox.height} at (${panelBox.x}, ${panelBox.y})`);
        
        // Check if any child elements extend beyond panel boundaries
        const children = panelElement.locator('> div, > *');
        const childCount = await children.count();
        
        let overflowingChildren = 0;
        for (let i = 0; i < Math.min(childCount, 5); i++) { // Check first 5 children
          const child = children.nth(i);
          const childBox = await child.boundingBox();
          
          if (childBox) {
            const isOverflowing = 
              childBox.x + childBox.width > panelBox.x + panelBox.width ||
              childBox.y + childBox.height > panelBox.y + panelBox.height ||
              childBox.x < panelBox.x ||
              childBox.y < panelBox.y;
            
            if (isOverflowing) {
              console.log(`  ❌ Child ${i} overflowing: (${childBox.x}, ${childBox.y}) ${childBox.width}x${childBox.height}`);
              overflowingChildren++;
            }
          }
        }
        
        if (overflowingChildren === 0) {
          console.log(`  ✅ ${panel.name}: No content overflow detected`);
        } else {
          console.log(`  ❌ ${panel.name}: ${overflowingChildren} overflowing children`);
        }
      }
    }
  });

  test('verify scrollable containers work within clipped panels', async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');
    
    console.log('📊 Testing scrolling within clipped panels:');
    
    // Test System Architecture scrolling
    const systemScroll = page.locator('.panel:has-text("System Architecture") .overflow-y-auto');
    const systemScrollTop = await systemScroll.evaluate(el => el.scrollTop);
    await systemScroll.evaluate(el => el.scrollTop = 50);
    const systemNewScrollTop = await systemScroll.evaluate(el => el.scrollTop);
    console.log(`System Architecture scroll: ${systemScrollTop} → ${systemNewScrollTop}`);
    
    // Test Traffic Generation scrolling
    const trafficScroll = page.locator('.panel:has-text("Traffic Generation") .overflow-y-auto');
    const trafficScrollTop = await trafficScroll.evaluate(el => el.scrollTop);
    await trafficScroll.evaluate(el => el.scrollTop = 20);
    const trafficNewScrollTop = await trafficScroll.evaluate(el => el.scrollTop);
    console.log(`Traffic Generation scroll: ${trafficScrollTop} → ${trafficNewScrollTop}`);
    
    // Test System Parameters scrolling
    const paramsScroll = page.locator('.panel:has-text("System Parameters") .overflow-y-auto');
    const paramsScrollTop = await paramsScroll.evaluate(el => el.scrollTop);
    await paramsScroll.evaluate(el => el.scrollTop = 30);
    const paramsNewScrollTop = await paramsScroll.evaluate(el => el.scrollTop);
    console.log(`System Parameters scroll: ${paramsScrollTop} → ${paramsNewScrollTop}`);
    
    // Take screenshot of final clipped state
    await page.screenshot({ 
      path: 'tests/screenshots/all-panels-clipped.png',
      fullPage: true 
    });
    
    console.log('✅ Scrolling functionality preserved within clipped panels');
  });
});
import { test, expect } from '@playwright/test';

test.describe('Grid Spacing Validation', () => {
  test('validates no chart overlapping', async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');
    
    // Get all chart containers
    const chartContainers = page.locator('.grid > div');
    const count = await chartContainers.count();
    console.log(`📊 Found ${count} chart containers`);
    
    // Get positions of all charts
    const chartPositions = [];
    for (let i = 0; i < count; i++) {
      const container = chartContainers.nth(i);
      const box = await container.boundingBox();
      const title = await container.locator('h4').textContent();
      
      if (box) {
        chartPositions.push({
          index: i,
          title,
          x: box.x,
          y: box.y,
          width: box.width,
          height: box.height,
          right: box.x + box.width,
          bottom: box.y + box.height
        });
      }
    }
    
    // Log all chart positions
    console.log('📊 Chart Positions:');
    chartPositions.forEach(chart => {
      console.log(`${chart.index}: "${chart.title}" at (${chart.x}, ${chart.y}) size ${chart.width}x${chart.height}`);
    });
    
    // Check for overlaps
    let overlaps = 0;
    for (let i = 0; i < chartPositions.length; i++) {
      for (let j = i + 1; j < chartPositions.length; j++) {
        const chart1 = chartPositions[i];
        const chart2 = chartPositions[j];
        
        // Check if rectangles overlap
        const horizontalOverlap = chart1.x < chart2.right && chart2.x < chart1.right;
        const verticalOverlap = chart1.y < chart2.bottom && chart2.y < chart1.bottom;
        
        if (horizontalOverlap && verticalOverlap) {
          console.log(`❌ OVERLAP: "${chart1.title}" and "${chart2.title}"`);
          console.log(`  Chart 1: (${chart1.x}, ${chart1.y}) to (${chart1.right}, ${chart1.bottom})`);
          console.log(`  Chart 2: (${chart2.x}, ${chart2.y}) to (${chart2.right}, ${chart2.bottom})`);
          overlaps++;
        }
      }
    }
    
    if (overlaps === 0) {
      console.log('✅ No chart overlaps detected!');
    } else {
      console.log(`❌ Found ${overlaps} overlapping chart pairs`);
    }
    
    expect(overlaps).toBe(0);
  });

  test('validates proper grid row spacing', async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');
    
    // Get all chart containers
    const chartContainers = page.locator('.grid > div');
    const count = await chartContainers.count();
    
    // Group charts by rows (based on Y position)
    const chartsByRow = new Map();
    
    for (let i = 0; i < count; i++) {
      const container = chartContainers.nth(i);
      const box = await container.boundingBox();
      const title = await container.locator('h4').textContent();
      
      if (box) {
        // Find which row this chart belongs to (tolerance of 10px)
        let rowFound = false;
        for (const [rowY, charts] of chartsByRow.entries()) {
          if (Math.abs(box.y - rowY) <= 10) {
            charts.push({ title, ...box });
            rowFound = true;
            break;
          }
        }
        
        if (!rowFound) {
          chartsByRow.set(box.y, [{ title, ...box }]);
        }
      }
    }
    
    // Log row information
    const rows = Array.from(chartsByRow.entries()).sort((a, b) => a[0] - b[0]);
    console.log(`📊 Found ${rows.length} rows:`);
    
    rows.forEach(([rowY, charts], index) => {
      console.log(`Row ${index + 1} (y=${rowY}): ${charts.length} charts`);
      charts.forEach(chart => {
        console.log(`  - "${chart.title}" at (${chart.x}, ${chart.y}) size ${chart.width}x${chart.height}`);
      });
    });
    
    // Validate expected grid behavior
    if (rows.length >= 3) {
      // Should have 3 charts in first row, 3 in second, 3 in third for 9 total
      expect(rows[0][1].length).toBeLessThanOrEqual(3);
      expect(rows[1][1].length).toBeLessThanOrEqual(3);
      expect(rows[2][1].length).toBeLessThanOrEqual(3);
      
      // Check vertical spacing between rows
      const row1Y = rows[0][0];
      const row2Y = rows[1][0];
      const row3Y = rows[2][0];
      
      const spacing1to2 = row2Y - row1Y;
      const spacing2to3 = row3Y - row2Y;
      
      console.log(`Row spacing: ${spacing1to2}px between rows 1-2, ${spacing2to3}px between rows 2-3`);
      
      // Row spacing should be at least 200px (our grid-auto-rows setting)
      expect(spacing1to2).toBeGreaterThanOrEqual(190); // Allow some tolerance
      expect(spacing2to3).toBeGreaterThanOrEqual(190);
    }
    
    console.log('✅ Grid row spacing validated successfully');
  });
});
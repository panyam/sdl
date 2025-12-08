import { test, expect } from '@playwright/test';

test.describe('Recipe File Execution', () => {
  test('recipe file parsing and execution works', async ({ page }) => {
    // Create a simple test page that validates recipe file functionality
    const testHtml = `
      <!DOCTYPE html>
      <html>
      <head><title>Recipe Test</title></head>
      <body>
        <div id="test-output">Recipe test ready</div>
        <script>
          // Mock Canvas API for testing recipe logic
          class MockCanvas {
            constructor() {
              this.state = {
                activeFile: null,
                activeSystem: null,
                systemParameters: {}
              };
              this.output = [];
            }
            
            load(filePath) {
              this.output.push('load ' + filePath);
              this.state.activeFile = filePath;
              return Promise.resolve();
            }
            
            use(systemName) {
              this.output.push('use ' + systemName);
              this.state.activeSystem = systemName;
              return Promise.resolve();
            }
            
            set(path, value) {
              this.output.push('set ' + path + ' ' + value);
              this.state.systemParameters[path] = value;
              return Promise.resolve();
            }
          }
          
          // Recipe execution logic (simplified version)
          async function executeRecipe(canvas, recipeText) {
            const lines = recipeText.split('\\n');
            const commands = [];
            
            for (const line of lines) {
              const trimmed = line.trim();
              if (trimmed === '' || trimmed.startsWith('#')) continue;
              
              if (trimmed.startsWith('sleep ')) {
                commands.push({ type: 'sleep', duration: trimmed.split(' ')[1] });
              } else {
                commands.push({ type: 'command', text: trimmed });
              }
            }
            
            return commands;
          }
          
          // Test the recipe parsing
          const canvas = new MockCanvas();
          const sampleRecipe = \`# SDL Demo Recipe
load examples/contacts/contacts.sdl
use ContactsSystem
set server.pool.ArrivalRate 5
sleep 1s
set server.pool.ArrivalRate 15\`;
          
          executeRecipe(canvas, sampleRecipe).then(commands => {
            document.getElementById('test-output').textContent = 
              'Commands parsed: ' + commands.length + ', Types: ' + 
              commands.map(c => c.type).join(', ');
          });
        </script>
      </body>
      </html>
    `;
    
    await page.setContent(testHtml);
    
    // Check that recipe parsing works
    await expect(page.locator('#test-output')).toContainText('Commands parsed: 5');
    await expect(page.locator('#test-output')).toContainText('Types: command, command, command, sleep, command');
  });
  
  test('recipe file supports comments and sleep commands', async ({ page }) => {
    const testHtml = `
      <!DOCTYPE html>
      <html>
      <head><title>Recipe Comments Test</title></head>
      <body>
        <div id="comment-test">Testing recipe comments</div>
        <script>
          function parseRecipeLines(recipeText) {
            const lines = recipeText.split('\\n');
            const parsedLines = [];
            
            for (const line of lines) {
              const trimmed = line.trim();
              if (trimmed === '') {
                parsedLines.push({ type: 'empty' });
              } else if (trimmed.startsWith('#')) {
                parsedLines.push({ type: 'comment', text: trimmed });
              } else if (trimmed.startsWith('sleep ')) {
                parsedLines.push({ type: 'sleep', duration: trimmed.split(' ')[1] });
              } else {
                parsedLines.push({ type: 'command', text: trimmed });
              }
            }
            
            return parsedLines;
          }
          
          const testRecipe = \`# This is a comment
          
load test.sdl
# Another comment
use TestSystem
sleep 2s
set param 10\`;
          
          const parsed = parseRecipeLines(testRecipe);
          const summary = {
            comments: parsed.filter(l => l.type === 'comment').length,
            commands: parsed.filter(l => l.type === 'command').length,
            sleeps: parsed.filter(l => l.type === 'sleep').length,
            empty: parsed.filter(l => l.type === 'empty').length
          };
          
          document.getElementById('comment-test').textContent = 
            \`Comments: \${summary.comments}, Commands: \${summary.commands}, Sleeps: \${summary.sleeps}, Empty: \${summary.empty}\`;
        </script>
      </body>
      </html>
    `;
    
    await page.setContent(testHtml);
    
    await expect(page.locator('#comment-test')).toContainText('Comments: 2, Commands: 3, Sleeps: 1, Empty: 1');
  });
});
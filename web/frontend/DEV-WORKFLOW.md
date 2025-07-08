# SDL Dashboard Development Workflow

## Quick Inner Loop for Dashboard Development

This setup provides a fast feedback loop for testing dashboard changes against a running SDL server.

### Prerequisites
- SDL server running on port 8080 (`sdl serve --port 8080`)
- Node dependencies installed (`npm install`)

### Development Commands

#### 1. Full Development Test Loop
```bash
# From project root
make dev-test
```
This will:
- Build the SDL binary
- Build the web frontend  
- Run comprehensive dashboard tests
- Take screenshots for visual validation
- Test API integration

#### 2. Quick Validation
```bash
# From project root  
make dev-quick
```
- Fast build and basic validation
- Just checks dashboard loads correctly
- Takes a quick screenshot

#### 3. Individual Test Commands
```bash
# From web/ directory
npm run dev-test        # Full development test
npm run dev-quick       # Quick validation only
npm run dev-screenshot  # Just take screenshot
./dev-test.sh          # Manual script execution
```

### Test Scenarios Covered

#### Initial State Validation
- Dashboard loads successfully
- Empty state messages display correctly:
  - "No System Loaded" in architecture panel
  - "No Parameters Available" in parameters panel
  - "No Traffic Generators" in traffic panel
  - "No Metrics Available" in metrics panel

#### System Loading Workflow
- Load SDL file via API (`examples/contacts/contacts.sdl`)
- Activate system via API (`ContactsSystem`)
- Verify WebSocket updates dashboard
- Validate system architecture appears
- Check component topology displays correctly

#### API Integration Testing
- Canvas state management
- System diagram endpoint
- Parameter setting functionality
- Generator management
- Measurement data retrieval

#### Visual Validation
- Screenshots saved to `tests/screenshots/`:
  - `dev-loop-initial.png` - Empty dashboard state
  - `dev-loop-with-system.png` - Dashboard with loaded system
  - `dev-quick-check.png` - Quick validation screenshot

### Troubleshooting the Architecture Panel

The system architecture panel shows "No System Loaded" when:

1. **No system is loaded**: Use the API to load and activate a system:
   ```bash
   curl -X POST http://localhost:8080/api/load \
     -H "Content-Type: application/json" \
     -d '{"filePath": "examples/contacts/contacts.sdl"}'
   
   curl -X POST http://localhost:8080/api/use \
     -H "Content-Type: application/json" \
     -d '{"systemName": "ContactsSystem"}'
   ```

2. **WebSocket not connected**: Check browser console for WebSocket errors

3. **Diagram API failing**: Test diagram endpoint:
   ```bash
   curl http://localhost:8080/api/canvas/diagram
   ```

### Development Iteration Process

1. **Make changes** to dashboard code (`web/src/dashboard.ts`) or backend (`console/canvas_web.go`)
2. **Build and test** with `make dev-test` 
3. **Check screenshots** in `tests/screenshots/` for visual validation
4. **Review console output** for any API/WebSocket errors
5. **Repeat** until desired functionality is working

### Example Outputs

Successful test run shows:
```
✅ Initial state validated
✅ System loaded successfully  
✅ System activated successfully
✅ System name appears in header
✅ System components visible
✅ Empty state messages cleared
✅ Diagram API returns system data
✅ Parameter setting works
🎉 Development loop test completed successfully!
```

### Advanced Testing

For more complex scenarios, extend `development-loop.spec.ts`:
- Traffic generator functionality
- Measurement data visualization  
- Parameter adjustment workflows
- Multi-system scenarios
- Error handling edge cases

This workflow enables rapid iteration on dashboard features while ensuring integration with the SDL Canvas API works correctly.
# Dynamic System Suggestions Test

## Overview

The SDL console now dynamically extracts system names from loaded SDL files instead of using hardcoded suggestions.

## Implementation

### Changes Made

1. **Replaced hardcoded system list** in `getSystemSuggestions()`
2. **Added `getSystemNamesFromCanvas()`** function that:
   - Uses reflection to access Canvas's private `loadedFiles` field
   - Iterates through all loaded `*loader.FileStatus` objects
   - Calls `FileDecl.GetSystems()` on each loaded file
   - Extracts system names using `system.Name.Value`

### Code Details

```go
func getSystemNamesFromCanvas(canvas *console.Canvas) []string {
    var systemNames []string
    
    // Access Canvas's private loadedFiles field using reflection
    canvasValue := reflect.ValueOf(canvas).Elem()
    loadedFilesField := canvasValue.FieldByName("loadedFiles")
    
    // loadedFiles is map[string]*loader.FileStatus
    loadedFilesMap := loadedFilesField.Interface().(map[string]*loader.FileStatus)
    
    for _, fileStatus := range loadedFilesMap {
        if fileStatus == nil || fileStatus.FileDecl == nil {
            continue
        }
        
        // Get systems from this file
        systems, err := fileStatus.FileDecl.GetSystems()
        if err != nil {
            continue // Skip files with errors
        }
        
        // Add system names to our list
        for _, system := range systems {
            systemNames = append(systemNames, system.Name.Value)
        }
    }
    
    return systemNames
}
```

## Testing the Feature

### Test Scenario 1: Empty Console
```bash
./sdl console --port 8080
SDL> use <TAB>
# Should show no suggestions since no files loaded
```

### Test Scenario 2: Load SDL File
```bash
SDL> load examples/contacts/contacts.sdl
✅ Loaded: examples/contacts/contacts.sdl

SDL> use <TAB>
# Should now show "ContactsSystem" (from the actual SDL file)
ContactsSystem    System definition from loaded SDL files
```

### Test Scenario 3: Multiple Files
```bash
SDL> load examples/kafka/kafka.sdl  # (if it exists)
✅ Loaded: examples/kafka/kafka.sdl

SDL> use <TAB>
# Should show systems from both files:
ContactsSystem    System definition from loaded SDL files
KafkaSystem       System definition from loaded SDL files
```

### Test Scenario 4: Invalid File
```bash
SDL> load nonexistent.sdl
❌ Error: [load error]

SDL> use <TAB>
# Should still show previously loaded systems only
ContactsSystem    System definition from loaded SDL files
```

## Benefits

1. **Accurate suggestions**: Only shows systems that actually exist in loaded files
2. **Dynamic behavior**: Suggestions update as files are loaded/unloaded
3. **Real-time feedback**: No outdated or incorrect system names
4. **Extensible**: Works with any SDL file structure
5. **Error resilient**: Skips files with parsing errors gracefully

## Implementation Notes

### Why Reflection?
- Canvas's `loadedFiles` field is private (lowercase)
- No public getter method exists for accessing loaded file declarations
- Reflection provides access to the internal state without modifying Canvas API

### Alternative Approaches
1. **Add public method to Canvas**: `GetLoadedSystems() []string`
2. **Export loadedFiles field**: Change to `LoadedFiles`
3. **Add to Canvas state**: Include system names in `CanvasState`

### API Compatibility
- Uses existing `FileDecl.GetSystems()` method from the decl package
- Compatible with current SDL parsing infrastructure
- No breaking changes to existing Canvas API

## Verification

To verify this is working correctly:

1. **Start console and check no suggestions**:
   ```
   SDL> use <TAB>  # Empty list
   ```

2. **Load a file and verify system appears**:
   ```
   SDL> load examples/contacts/contacts.sdl
   SDL> use <TAB>  # Shows "ContactsSystem"
   ```

3. **Check system name matches the actual SDL file**:
   ```bash
   cat examples/contacts/contacts.sdl | grep "system "
   # Should show: system ContactsSystem { ... }
   ```

The dynamic system suggestions ensure the console auto-completion stays in sync with the actual loaded SDL declarations!
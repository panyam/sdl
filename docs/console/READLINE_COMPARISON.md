# Go Readline Libraries Comparison for SDL Console REPL

## Overview

This document compares the main Go readline libraries for implementing an enhanced REPL console for the SDL project. The current implementation uses basic `bufio.Scanner` which lacks features like command history, arrow key navigation, and tab completion.

## Library Comparison Table

| Library | Stars | Last Update | License | Key Features | Maintenance Status |
|---------|-------|-------------|---------|--------------|-------------------|
| **c-bata/go-prompt** | 5.3k | May 2024 | MIT | • Rich auto-completion with descriptions<br>• Live prefix changes<br>• Multi-line support<br>• Emacs-like shortcuts<br>• Cross-platform | Active |
| **chzyer/readline** | 2.1k | April 2022 | MIT | • GNU Readline compatible<br>• Vi/Emacs modes<br>• History search<br>• Password input<br>• Cross-platform | Inactive |
| **peterh/liner** | 1.1k | Moderate | MIT | • Simple API<br>• Basic completion<br>• History support<br>• Lightweight<br>• Thread-safe history | Moderate |
| **eiannone/keyboard** | ~600 | Active | MIT | • Simple keystroke capture<br>• Channel-based input<br>• Lightweight<br>• No readline features | Active |
| **rivo/tview** | 10k+ | Active | MIT | • Full TUI framework<br>• Input widgets<br>• Built on tcell<br>• Complex for simple REPL | Very Active |

## Detailed Analysis

### 1. **c-bata/go-prompt** ⭐ Recommended

**Pros:**
- Most popular and actively maintained readline library
- Rich completion system with descriptions (perfect for SDL commands)
- Built-in support for live prefix changes (useful for showing current context)
- Excellent documentation and examples
- Inspired by python-prompt-toolkit (proven design)
- Cross-platform with proper Windows support
- Easy integration with existing code

**Cons:**
- Larger dependency than simpler alternatives
- More complex API for advanced features
- Some open issues with multi-line editing

**Example Integration:**
```go
func completer(d prompt.Document) []prompt.Suggest {
    s := []prompt.Suggest{
        {Text: "load", Description: "Load an SDL file"},
        {Text: "use", Description: "Activate a system"},
        {Text: "set", Description: "Set parameter value"},
        {Text: "run", Description: "Run simulation"},
        {Text: "state", Description: "Show Canvas state"},
        {Text: "execute", Description: "Execute recipe file"},
    }
    return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
}

func executor(s string) {
    executeCommand(canvas, s)
}

p := prompt.New(
    executor,
    completer,
    prompt.OptionPrefix("SDL> "),
    prompt.OptionTitle("SDL Console"),
)
p.Run()
```

### 2. **chzyer/readline**

**Pros:**
- GNU Readline compatible behavior
- Vi and Emacs editing modes
- Built-in history search (Ctrl+R)
- Password input support
- Good for users familiar with bash/zsh

**Cons:**
- No longer actively maintained (last update 2022)
- Considered "discontinued" by package analysis tools
- Less modern API compared to go-prompt
- Completion system less flexible

**Example Integration:**
```go
config := &readline.Config{
    Prompt:          "SDL> ",
    HistoryFile:     "/tmp/sdl_history.tmp",
    AutoComplete:    createCompleter(),
    InterruptPrompt: "^C",
    EOFPrompt:       "exit",
}

rl, err := readline.NewEx(config)
if err != nil {
    panic(err)
}
defer rl.Close()

for {
    line, err := rl.Readline()
    if err != nil {
        break
    }
    executeCommand(canvas, line)
}
```

### 3. **peterh/liner**

**Pros:**
- Simple, clean API
- Lightweight with minimal dependencies
- Thread-safe history operations
- Good middle ground between features and simplicity
- Used by several popular Go projects

**Cons:**
- Basic completion (no descriptions)
- Less feature-rich than go-prompt
- Limited customization options
- No multi-line support

**Example Integration:**
```go
line := liner.NewLiner()
defer line.Close()

line.SetCtrlCAborts(true)
line.SetCompleter(func(line string) (c []string) {
    commands := []string{"load", "use", "set", "run", "state", "execute"}
    for _, cmd := range commands {
        if strings.HasPrefix(cmd, strings.ToLower(line)) {
            c = append(c, cmd)
        }
    }
    return
})

for {
    if input, err := line.Prompt("SDL> "); err == nil {
        line.AppendHistory(input)
        executeCommand(canvas, input)
    } else {
        break
    }
}
```

### 4. **Alternative Approaches**

**eiannone/keyboard:**
- Too low-level for readline functionality
- Would require building all features from scratch
- Good for capturing special keys but not for REPL

**rivo/tview:**
- Overkill for a simple REPL
- Better suited for full-screen TUI applications
- Would require significant refactoring

## Integration Complexity

| Library | Integration Effort | Code Changes Required | Learning Curve |
|---------|-------------------|----------------------|----------------|
| **go-prompt** | Medium | Moderate refactor of console.go | Medium - Rich API |
| **readline** | Low | Minimal changes | Low - Simple API |
| **liner** | Very Low | Drop-in replacement | Very Low |
| **keyboard** | High | Complete rewrite | High |
| **tview** | Very High | Major refactor | High |

## Recommendation for SDL Console

**Primary Choice: c-bata/go-prompt**

Reasons:
1. **Active maintenance** - Critical for long-term project health
2. **Rich completions** - Can show command descriptions and parameter hints
3. **Modern design** - Better UX for conference demonstrations
4. **Extensibility** - Can add features like syntax highlighting later
5. **Community** - Largest user base means better support

**Implementation Plan:**
1. Add `github.com/c-bata/go-prompt` dependency
2. Refactor `startREPL()` to use prompt.New()
3. Implement rich completer with command descriptions
4. Add context-aware completions (e.g., system names after "use")
5. Maintain recipe file execution compatibility

**Fallback Option: peterh/liner**

If go-prompt proves too complex or has issues, liner provides a simple drop-in replacement that would immediately improve the current experience with minimal effort.

## Conclusion

While the current `bufio.Scanner` implementation works, upgrading to a proper readline library would significantly improve the user experience for SDL console. The go-prompt library offers the best balance of features, maintenance, and user experience for creating a professional REPL that will impress at conference demonstrations.
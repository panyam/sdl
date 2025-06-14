# SDL Console Shell Command Feature Test

## Overview
The SDL console now supports executing shell commands by prefixing them with `!`.

## Test Commands

### Basic File Operations
```
SDL> !pwd
SDL> !ls
SDL> !ls -la
```

### Git Operations
```
SDL> !git status
SDL> !git log --oneline -5
SDL> !git branch
```

### Process and System Information
```
SDL> !ps aux | grep sdl
SDL> !echo "Hello from shell!"
SDL> !date
```

### File Content Viewing
```
SDL> !cat examples/contacts/contacts.sdl
SDL> !head -10 go.mod
```

### Development Commands
```
SDL> !go version
SDL> !make
SDL> !docker ps
```

## Tab Completion Tests

### Test Shell Command Suggestions
1. Type `!` and press Tab - should show common shell commands
2. Type `!l` and press Tab - should suggest `ls`
3. Type `!git ` and press Tab - should suggest git subcommands (if implemented)
4. Type `!cat ` and press Tab - should suggest files

## Features

1. **Command Execution**: Any shell command prefixed with `!` is executed
2. **Tab Completion**: Common shell commands are suggested with descriptions
3. **File Completion**: For file-related commands, files are suggested
4. **Full Integration**: Shell commands are part of the command history
5. **Error Handling**: Shell command errors are displayed clearly

## Example Session

```
SDL> help
[... help output includes !<shell_command> ...]

SDL> !<TAB>
!ls        List directory contents
!pwd       Print working directory
!git       Git version control
...

SDL> !pwd
🐚 Running: pwd
/Users/sri/personal/golang/sdl

SDL> !ls -la examples/
🐚 Running: ls -la examples/
total 0
drwxr-xr-x  4 sri  staff  128 Jun 14 12:00 .
drwxr-xr-x 15 sri  staff  480 Jun 14 12:00 ..
drwxr-xr-x  3 sri  staff   96 Jun 14 12:00 contacts
...

SDL> load examples/contacts/contacts.sdl
✅ Loaded: examples/contacts/contacts.sdl

SDL> !git status
🐚 Running: git status
On branch main
Your branch is ahead of 'origin/main' by 1 commit.
...
```

This feature makes the SDL console much more versatile for development workflows!

## Understand the Project First
- I am continuing with a previous project.  You will find the summaries in SUMMARY.md files located in the top level as various sub folders.  NEXTSTEPS.md is used to note what has been completed and what are next steps in our roadmap.  Thorougly understand it and give me a recap so we can continue where we left off.

## Coding Style and Conservativeness
- Be conservative on how many comments are you are adding or modifying unless it is absolutely necessary (for example a comment could be contradicting what is going on - in which case it is prudent to modify it).  
- When modifying files just focus on areas where the change is required instead of diving into a full fledged refactor.
- Make sure you ignore 'gen' and 'node_modules' as it has a lot of files you wont need for most things and are either auto generated or just package dependencies
- When updating .md files and in commit messages use emojis and flowerly languages sparingly.  We dont want to be too grandios or overpromising.
- Make sure the playwright tool is setup so you can inspect the browser when we are implementing and testing the Dashboard features.
- Do not refer to claude or anthropic or gemini in your commit messages
- Do not rebuild the server - it will be continuosly be rebuilt and run by the air configs.  Output of the server will be written to /tmp/sdlserver.log.  Build errors will also be shown in this log file.
- Find the root cause of an issue before figuring out a solution.  Fix problems.
- Do not create workarounds for issues without asking.  Always find the root cause of an issue and fix it.
- The web module automatically builds when files are changed - DO NOT run npm build or npm run build commands.
- Proto files are automatically regenerated when changed - DO NOT run buf generate commands.

## WASM Build Memory
- You can build the wasm binary by simply doing `make wasmbin` in the top level directory but this is being built as files are changed.

## Continuous Builds

Builds for frontend, wasm, backend are all running continuously and can be queried using the `devloop` cli tool.   You have the following devloop commands:
- `devloop config` - Get configuration from running devloop server
- `devloop paths` - List all file patterns being watched
- `devloop trigger <rulename>` - Trigger execution of a specific rule
- `devloop logs <rulename>`  - Stream logs from running devloop server
- `devloop status <rulename>` - Get status of rules from running devloop server

## Summary instructions

- When you are using compact, please focus on test output and code changes

- For the ROADMAP.md always use the top-level ./ROADMAP.md so we have a global view of the roadmap instead of being fragemented in various folders.

## SDL Demo Guidelines
- Make sure when you create SDL demos they are not as markdown but as .recipe files that are executable with pause points that print out what is going to be come next before the SDL command is executed.

## Session Workflow Memories
- When you checkpoint update all relevant .md files with our latest understanding, statuses and progress in the current session and then commit.


## SDL System Declaration Notes
- In SDL system declaration you can declare the components in any order. There are no "set" statements. You pass the dependencies in the constructor of a "use" keyword.  For example:
```system Twitter {
    use app AppServer(db = database)
    use db Database
}```
- Here the AppServer component has a "db" dependency that is set by the "database" component declared in the next line.

## Available commands

- `buf generate`- To generate protos
- `make` - To generate all binaries
- `make dash` - To rebuild the web dashboard


## SDL Demo Guidelines
- Make sure when you create SDL demos they are not as markdown but as .recipe files that are executable with pause points that print out what is going to be come next before the SDL command is executed.

**Session Workflow Memories:**
- When you checkpoint update all relevant .md files with our latest understanding, statuses and progress in the current session and then commit.

## Latest Session Progress (July 8, 2025)

### Go Recipe Parser Implementation ✅ COMPLETED
**Objective**: Create Go version of TypeScript recipe parser for use in SystemDetailTool and WASM mode

**Key Accomplishments**:
- **Complete Parser Port**: Created `tools/shared/recipe/` package with full TypeScript parity
- **Comprehensive Testing**: 100% test coverage including real Bitly recipe validation (115 steps)
- **Security Model**: Extensive validation preventing unsupported shell syntax and filesystem access
- **Command Line Parser**: Handles quoted strings, complex arguments, and edge cases correctly

**Files Created/Modified**:
- `tools/shared/recipe/command.go` - Command types and structures
- `tools/shared/recipe/validator.go` - Validation patterns and security checks  
- `tools/shared/recipe/parser.go` - Core parsing logic with command line parsing
- `tools/shared/recipe/parser_test.go` - Comprehensive test suite
- `tools/shared/recipe/bitly_test.go` - Real-world recipe testing

### @stdlib Import Support ✅ COMPLETED
**Objective**: Enable SystemDetailTool to compile SDL files with @stdlib imports like Bitly example

**Key Accomplishments**:
- **Memory Filesystem**: Loads stdlib files from `examples/stdlib/` into memory
- **Custom Resolver**: Created StdlibPrefixFS to handle @stdlib/ prefix stripping
- **Path Resolution**: Robust path finding for different runtime environments (tests vs main)
- **Complete Integration**: Bitly SDL now compiles successfully with all imports resolved

**Technical Solution**:
- Fixed CompositeFS mount prefix handling issue
- Created wrapper filesystem that strips @stdlib/ prefix before file lookup
- Integrated with existing MemoryResolver in SystemDetailTool
- Added comprehensive test coverage for @stdlib functionality

**Files Modified**:
- `tools/systemdetail/sdl.go` - Added stdlib support and StdlibPrefixFS
- `tools/systemdetail/tool_test.go` - Added @stdlib integration tests
- `cmd/debug-bitly/main.go` - Debug program for standalone testing

### SystemDetailTool Enhancement ✅ COMPLETED
**Objective**: Integrate recipe parser and enable full Bitly example compilation

**Key Accomplishments**:
- **Recipe Integration**: SystemDetailTool now uses shared Go recipe parser
- **Environment Agnostic**: Works in CLI, WASM, and test environments
- **Error Handling**: Proper validation errors with line numbers
- **Debug Infrastructure**: Standalone testing capabilities

**Current Status**:
- Bitly SDL compiles successfully (1 system: [Bitly])
- Recipe parsing works (115 executable steps from 216 total lines)
- All tests pass with comprehensive coverage
- Ready for WASM integration

### Next Critical Steps
1. **WASM Bindings**: Create WASM bindings for SystemDetailTool
2. **Browser Integration**: Update System details page to use WASM SystemDetailTool  
3. **Recipe Execution**: Complete step-by-step recipe execution in browser
4. **UI Enhancement**: Show recipe progress and integrate with existing panels

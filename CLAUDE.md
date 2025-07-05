
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

Builds for frontend, wasm, backend are all running continuously and can be queried against the remote `devloop` mcp server with project ID - "sdl".  You can use it to get the results of the latest build for the various components being watched and live loaded.

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

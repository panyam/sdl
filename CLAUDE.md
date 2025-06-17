Read and learn all about this project by looking at all the SUMMARY.md and NEXTSTEPS.md.  

I am continuing with a previous project.  You will find the summaries in SUMMARY.md files located in the top level as various sub folders.  NEXTSTEPS.md is used to note what has been completed and what are next steps in our roadmap.
Thorougly understand it and give me a recap so we can continue where we left off.

## FlowEval Runtime Migration (June 2025)
When continuing work on FlowEval, note that we're in the middle of migrating from string-based to runtime-based flow analysis:
- **New code location**: runtime/flowrteval.go (runtime-based) replacing runtime/floweval.go (string-based)
- **Key types**: RateMap (runtime/ratemap.go), FlowScope (runtime/flowscope.go), GeneratorEntryPointRuntime
- **Architecture**: Uses actual ComponentInstance objects from SimpleEval, no duplicate instances
- **Pattern**: NWBase wrapper provides smart defaults for non-flow-analyzable components
- **Status**: Steps 1-7 complete, need to finish migration (steps 8-9) and update Canvas integration
- **Test with**: `go test -v ./runtime -run "TestFlowEvalRuntime|TestSolveSystemFlowsRuntime"`   
Also be conservative on how many comments are you are adding or modifying unless it is absolutely necessary (for example a comment could be contradicting what is going on - in which case it is prudent to modify it).  
When modifying files just focus on areas where the change is required instead of diving into a full fledged refactor.
Make sure you ignore node_modules as it has a lot of files you wont need for most things
When updating .md files and in commit messages use emojis and flowerly languages sparingly.  We dont want to be too grandios or overpromising.
Make sure the playwright tool is setup so you can inspect the browser when we are implementing and testing the Dashboard features.
Do not refer to claude or anthropic in your commit messages
Do not rebuild the server - it will be continuosly be rebuilt and run by the air configs.  Output of the server will be written to /tmp/sdlserver.log.  Build errors will also be shown in this log file.
Find the root cause of an issue before figuring out a solution.  Fix problems.
Do not create workarounds for issues without asking.  Always find the root cause of an issue and fix it.

# Summary instructions

When you are using compact, please focus on test output and code changes

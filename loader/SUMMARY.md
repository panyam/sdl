
## DSL Loader Package Summary

1. Decoupling: The loader package depends on decl (for AST) and uses interfaces (Parser, FileResolver) to interact with parsing and file system specifics. decl and parser remain unaware of loader.
2. FileResolver: Abstracting file access allows loading from sources other than the local filesystem later (e.g., embedded resources, network). The canonicalPath return value is crucial for reliable caching and cycle detection.
3. Parser Interface: Makes the loader independent of the specific goyacc parser implementation.
4. Cycle Detection: The pending map tracks files currently in the recursion stack for a single LoadRootFile call. If a path is encountered that's already in pending, it's a cycle.
5. Caching: The loadedFiles map acts as a cache. Once a file (identified by its canonical path) is parsed, it's stored and reused if imported again during the same LoadRootFile operation.
6. Max Depth: A simple counter prevents infinite recursion in non-cyclic but deeply nested imports.
7. Type Inference Timing: Inference (decl.InferTypesForFile) is explicitly called after all recursive loading is complete. This ensures that all component/enum definitions from all loaded files are available in their respective FileDecls' resolved maps before type checking begins, allowing cross-file references to be resolved correctly during inference.
8. Error Handling: The LoadResult struct collects all errors (parsing, resolution, cycle, depth, inference). The main LoadRootFile function returns the first critical error encountered during loading/parsing, but the caller can inspect result.Errors for the full list, including type errors.
9. State Management: The Loader resets its internal state (loadedFiles, pending) for each LoadRootFile call, making it reusable. Concurrency protection is added via a mutex, although the initial implementation is sequential.



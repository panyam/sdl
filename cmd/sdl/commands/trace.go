package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/panyam/sdl/core"
	"github.com/panyam/sdl/decl"
	"github.com/panyam/sdl/loader"
	"github.com/panyam/sdl/runtime"
	"github.com/spf13/cobra"
)

var traceCmd = &cobra.Command{
	Use:   "trace <system_name> <method_call_string>",
	Short: "Traces the execution of a specific operation",
	Long: `Executes a specific method call within a system (e.g., "myService.doWork")
and outputs a detailed execution trace. The output is a JSON file that can be
used by other commands like 'diagram dynamic' to generate visualizations.`,
	Args: cobra.ExactArgs(2), // system_name and method_call_string
	Run: func(cmd *cobra.Command, args []string) {
		systemName := args[0]
		methodCallString := args[1]

		outputFile, _ := cmd.Flags().GetString("out")
		if outputFile == "" {
			fmt.Fprintln(os.Stderr, "Error: Output file must be specified with --out or -o.")
			os.Exit(1)
		}

		if dslFilePath == "" {
			fmt.Fprintln(os.Stderr, "Error: DSL file path must be specified with -f or --file.")
			os.Exit(1)
		}

		// 1. Load and validate the SDL file
		sdlLoader := loader.NewLoader(nil, nil, 10)
		fileStatus, err := sdlLoader.LoadFile(dslFilePath, "", 0)
		if err != nil || fileStatus.HasErrors() {
			fmt.Fprintf(os.Stderr, "Error loading or parsing SDL file '%s': %v\n", dslFilePath, err)
			fileStatus.PrintErrors()
			os.Exit(1)
		}
		if !sdlLoader.Validate(fileStatus) {
			fmt.Fprintf(os.Stderr, "Validation failed for file '%s':\n", dslFilePath)
			fileStatus.PrintErrors()
			os.Exit(1)
		}

		// 2. Initialize the runtime and the target system
		rt := runtime.NewRuntime(sdlLoader)
		fileInstance := rt.LoadFile(dslFilePath)
		system := fileInstance.NewSystem(systemName)
		if system == nil {
			fmt.Fprintf(os.Stderr, "System '%s' not found in file '%s'.\n", systemName, dslFilePath)
			os.Exit(1)
		}

		// 3. Create the tracer and the instrumented evaluator
		tracer := runtime.NewExecutionTracer()
		tracer.SetRuntime(rt)
		eval := runtime.NewSimpleEval(fileInstance, tracer)

		// 4. Prepare the system environment
		var totalSimTime core.Duration
		env := fileInstance.Env()
		eval.EvalInitSystem(system, env, &totalSimTime)

		// 5. Parse the method call string (simplified)
		parts := strings.Split(methodCallString, ".")
		if len(parts) != 2 {
			fmt.Fprintf(os.Stderr, "Error: Invalid method call string format. Expected 'instance.method', got '%s'\n", methodCallString)
			os.Exit(1)
		}
		instanceName := parts[0]
		methodName := parts[1]

		// 6. Create the call expression AST node to be evaluated
		callExpr := &decl.CallExpr{
			Function: &decl.MemberAccessExpr{
				Receiver: &decl.IdentifierExpr{Value: instanceName},
				Member:   &decl.IdentifierExpr{Value: methodName},
			},
			// Note: No arguments are parsed or passed in this simplified version
		}

		fmt.Printf("Tracing method '%s' in system '%s'...\n", methodCallString, systemName)

		// 7. Execute the trace
		_, _ = eval.Eval(callExpr, env, &totalSimTime)

		// 8. Process and save the results
		traceData := &runtime.TraceData{
			System:     systemName,
			EntryPoint: methodCallString,
			Events:     tracer.Events,
		}

		jsonData, err := json.MarshalIndent(traceData, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error marshalling trace data to JSON: %v\n", err)
			os.Exit(1)
		}

		err = os.WriteFile(outputFile, jsonData, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing trace data to %s: %v\n", outputFile, err)
			os.Exit(1)
		}

		fmt.Printf("Trace data successfully written to %s\n", outputFile)
	},
}

func init() {
	AddCommand(traceCmd)
	traceCmd.Flags().StringP("out", "o", "", "Output detailed trace data to a JSON file (required)")
	traceCmd.Flags().Int("depth", 0, "Limit trace depth (0 for unlimited)")
}

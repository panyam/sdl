package commands

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/panyam/sdl/loader"
	"github.com/panyam/sdl/runtime"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run <system_name> <instance_name> <method_name>",
	Short: "Runs a simulation for a specific system method",
	Long: `Executes a method on a component instance within a system a specified number of times 
to gather performance and result data. This command is designed for statistical 
analysis of a system's behavior under simulated load.

The results, including latency, return values, and errors for each run, are
saved to a JSON file for further analysis by commands like 'sdl plot'.`,
	Args: cobra.ExactArgs(3), // Now requires system, instance, and method
	Run: func(cmd *cobra.Command, args []string) {
		systemName := args[0]
		instanceName := args[1]
		methodName := args[2]

		totalRuns, _ := cmd.Flags().GetInt("runs")
		numWorkers, _ := cmd.Flags().GetInt("workers")
		outputFile, _ := cmd.Flags().GetString("out")

		if dslFilePath == "" {
			fmt.Fprintln(os.Stderr, "Error: DSL file path must be specified with -f or --file.")
			os.Exit(1)
		}
		if outputFile == "" {
			fmt.Fprintln(os.Stderr, "Error: Output file must be specified with --out or -o.")
			os.Exit(1)
		}

		fmt.Printf("Starting simulation for %s.%s.%s...\n", systemName, instanceName, methodName)
		fmt.Printf("Total Runs: %d, Concurrent Workers: %d\n", totalRuns, numWorkers)

		// 1. Load and initialize the system
		sdlLoader := loader.NewLoader(nil, nil, 10)
		fileStatus, err := sdlLoader.LoadFile(dslFilePath, "", 0)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading SDL file '%s': %v\n", dslFilePath, err)
			os.Exit(1)
		}

		if !sdlLoader.Validate(fileStatus) {
			fmt.Fprintf(os.Stderr, "Validation failed for file '%s':\n", dslFilePath)
			fileStatus.PrintErrors()
			os.Exit(1)
		}

		rt := runtime.NewRuntime(sdlLoader)
		fileInstance := rt.LoadFile(dslFilePath) // This will be fast as it's cached
		system := fileInstance.NewSystem(systemName)
		if system == nil {
			fmt.Fprintf(os.Stderr, "System '%s' not found in file '%s'.\n", systemName, dslFilePath)
			os.Exit(1)
		}

		// 2. Setup for concurrent execution
		batchSize := totalRuns / 100
		if batchSize == 0 {
			batchSize = 1
		}
		if batchSize > 1000 {
			batchSize = 1000
		}
		numBatches := (totalRuns + batchSize - 1) / batchSize

		allResults := make([]RunResult, 0, totalRuns)
		resultsChan := make(chan []RunResult, numWorkers)
		var wg sync.WaitGroup
		var resultMutex sync.Mutex

		wg.Add(1)
		go func() {
			defer wg.Done()
			for batchResults := range resultsChan {
				resultMutex.Lock()
				allResults = append(allResults, batchResults...)
				resultMutex.Unlock()
			}
		}()

		fmt.Println("Simulation in progress...")
		startTime := time.Now()

		onBatch := func(batch int, batchVals []runtime.Value) {
			batchResults := make([]RunResult, len(batchVals))
			now := time.Now().UnixMilli()
			for i, val := range batchVals {
				batchResults[i] = RunResult{
					Timestamp:   now,
					Latency:     val.Time * 1000,
					ResultValue: val.String(),
					IsError:     false, // Placeholder
				}
			}
			resultsChan <- batchResults
			if (batch+1)%10 == 0 || batch == numBatches-1 {
				log.Printf("  ... completed %d / %d batches\n", batch+1, numBatches)
			}
		}

		// The obj argument is the *instance* name.
		runtime.RunCallInBatches(system, instanceName, methodName, numBatches, batchSize, numWorkers, onBatch)

		close(resultsChan)
		wg.Wait()

		duration := time.Since(startTime)
		fmt.Printf("Simulation finished in %v.\n", duration)
		fmt.Printf("Collected %d results.\n", len(allResults))

		jsonData, err := json.MarshalIndent(allResults, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error marshalling results to JSON: %v\n", err)
			os.Exit(1)
		}
		err = os.WriteFile(outputFile, jsonData, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing JSON results to %s: %v\n", outputFile, err)
			os.Exit(1)
		}
		fmt.Printf("Results successfully written to %s\n", outputFile)
	},
}

func init() {
	AddCommand(runCmd)
	runCmd.Flags().Int("runs", 1000, "Total number of simulation runs to execute.")
	runCmd.Flags().Int("workers", 50, "Number of concurrent workers to run the simulation.")
	runCmd.Flags().StringP("out", "o", "", "Output file path for the detailed JSON results (required).")
}

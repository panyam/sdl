package commands

import (
	// "github.com/panyam/sdl/decl"
	// "gonum.org/v1/plot" // For actual plotting
	// "gonum.org/v1/plot/plotter"
	// "gonum.org/v1/plot/vg"

	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/panyam/sdl/decl"
	"github.com/panyam/sdl/loader"
	"github.com/panyam/sdl/runtime"
	"github.com/spf13/cobra"
)

func plotCmd() *cobra.Command {
	plotCmd := &cobra.Command{
		Use:   "plot <system> <component> <method>",
		Short: "Generates plots of running System.Component.Method calls.",
		Long: `Generates latency and count plots for running method calls multiple times.

	The call is run <num_batches> times and in each batch it is run <batchsize> times.
  `,
		Args: cobra.ExactArgs(3), // metric_type, system_name, analysis_name
		Run: func(cmd *cobra.Command, args []string) {
			systemName := args[0]
			componentName := args[1]
			methodName := args[2]

			numWorkers, _ := cmd.Flags().GetInt("numworkers")
			numBatches, _ := cmd.Flags().GetInt("numbatches")
			batchSize, _ := cmd.Flags().GetInt("batchsize")

			outputFile, _ := cmd.Flags().GetString("output")
			// resultsFile, _ := cmd.Flags().GetString("results")

			/*
				title, _ := cmd.Flags().GetString("title")
				xLabel, _ := cmd.Flags().GetString("xlabel")
				yLabel, _ := cmd.Flags().GetString("ylabel")
				percentilesStr, _ := cmd.Flags().GetString("percentiles")
			*/
			// --- End Placeholder ---

			l := loader.NewLoader(nil, nil, 10) // Max depth 10
			rt := runtime.NewRuntime(l)
			fi := rt.LoadFile(dslFilePath)
			system := fi.NewSystem(systemName)

			avgVals := make([]DataPoint, numBatches, numBatches)
			p50Vals := make([]DataPoint, numBatches, numBatches)
			p90Vals := make([]DataPoint, numBatches, numBatches)
			p99Vals := make([]DataPoint, numBatches, numBatches)
			now := time.Now()
			timeDelta := time.Second * 1
			var lastReported time.Time
			runtime.RunCallInBatches(system, componentName, methodName, numBatches, batchSize, numWorkers, func(batch int, batchVals []decl.Value) {
				log.Println("Generating batch: ", batch)
				if time.Now().Sub(lastReported) < time.Second*5 {
					lastReported = time.Now()
				}
				// sort batch vals so we can pick percentile values
				sort.Slice(batchVals, func(i, j int) bool {
					return batchVals[i].Time < batchVals[j].Time
				})

				t := now.Add(time.Duration(batch) * timeDelta)

				p50Idx := int(float64(batchSize) * 0.5)
				p50Vals[batch].Y = batchVals[p50Idx].Time // FloatVal()

				p90Idx := int(float64(batchSize) * 0.9)
				p90Vals[batch].Y = batchVals[p90Idx].Time // FloatVal()

				p99Idx := int(float64(batchSize) * 0.99)
				p99Vals[batch].Y = batchVals[p99Idx].Time // FloatVal()
				for _, bv := range batchVals {
					avgVals[batch].Y += bv.Time // FloatVal()
				}
				avgVals[batch].Y /= float64(batchSize)

				avgVals[batch].X = t.UnixMilli()
				p50Vals[batch].X = t.UnixMilli()
				p90Vals[batch].X = t.UnixMilli()
				p99Vals[batch].X = t.UnixMilli()
			})

			// Now plot them!
			pctFileName := func(prefix string) string {
				if strings.HasSuffix(outputFile, ".svg") {
					return strings.Replace(outputFile, ".svg", fmt.Sprintf("%s.svg", prefix), 1)
				}
				return fmt.Sprintf("%s.%s.svg", outputFile, prefix)
			}
			plot(pctFileName("Avg"), avgVals, "Time", "Latency Avg (ms)", "Avg API Latency over Time")
			plot(pctFileName("p50"), p50Vals, "Time", "Latency p50 (ms)", "p50 API Latency over Time")
			plot(pctFileName("p90"), p90Vals, "Time", "Latency p90 (ms)", "p90 API Latency over Time")
			plot(pctFileName("p99"), p99Vals, "Time", "Latency p99 (ms)", "p99 API Latency over Time")
		},
	}

	plotCmd.Flags().StringP("output", "o", "output.svg", "Output file path for the plot (e.g., plot.svg)")
	plotCmd.Flags().String("results", "", "Path to a JSON results file (from 'sdl plot --json-results')")
	plotCmd.Flags().String("title", "Latency vs Time", "Title for the plot")
	plotCmd.Flags().String("xlabel", "Time", "Label for the X-axis")
	plotCmd.Flags().String("ylabel", "Latency", "Label for the Y-axis")
	plotCmd.Flags().Int("numbatches", 1000, "Size of each batch")
	plotCmd.Flags().Int("numworkers", 50, "Number of parallel workers")
	plotCmd.Flags().Int("batchsize", 1000, "Size of each batch")
	plotCmd.Flags().String("percentiles", "0.5,0.9,0.99", "Comma-separated percentiles to mark on CDF (e.g., '0.5,0.99').  Each percentile will be shown in its own graph.")
	return plotCmd
}

func init() {
	// Add 'plot' as a subcommand of 'visualize', or directly if 'visualize' is just a namespace.
	// If we had a visualizeCmd:
	// visualizeCmd.AddCommand(plotCmd)
	// Else, add directly:
	AddCommand(plotCmd())
}

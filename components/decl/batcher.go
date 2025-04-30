// sdl/decl/components.go (or separate files like decl/disk.go, decl/index.go etc.)
package decl

import (
	"github.com/panyam/leetcoach/sdl/dsl"
	// Assume core ExecutionMode is accessible or redefine here
	// "github.com/panyam/leetcoach/sdl/core"
)

// --- Batcher ---

// BatchProcessor interface for declarative world
type BatchProcessor interface {
	// Name identifies the specific processor implementation for the VM
	ProcessorName() string
	// ProcessBatch returns the  for processing a batch.
	// The batchSize argument is an Expr yielding an INT.
	ProcessBatch(batchSize dsl.Expr) dsl.Expr
}

type Batcher struct {
	Name string
	// Config as Expr
	Policy         dsl.Expr       // LiteralExpr "SizeBased" or "TimeBased" ? Or Enum? Use string.
	BatchSize      dsl.Expr       // INT
	Timeout        dsl.Expr       // Duration
	ArrivalRate    dsl.Expr       // FLOAT (requests/sec)
	DownstreamProc BatchProcessor // Reference to the downstream processor
}

func NewBatcher(name string, policy, batchSize, timeout, arrivalRate dsl.Expr, proc BatchProcessor) *Batcher {
	return &Batcher{Name: name, Policy: policy, BatchSize: batchSize, Timeout: timeout, ArrivalRate: arrivalRate, DownstreamProc: proc}
}

// Submit generates  for submitting one item.
func (b *Batcher) Submit() dsl.Expr {
	// 1. Calculate Wait Time (Internal VM call)
	avgWaitTime := &dsl.InternalCallExpr{
		FuncName: "CalculateBatcherWaitTime",
		Args:     []dsl.Expr{b.Policy, b.BatchSize, b.Timeout, b.ArrivalRate},
	}
	// VM needs to know this returns Outcomes[Duration] and convert to AccessResult
	waitAsAccessResult := &dsl.InternalCallExpr{
		FuncName: "DurationToAccessResult", // VM helper
		Args:     []dsl.Expr{avgWaitTime},
	}

	// 2. Calculate Avg Batch Size (Internal VM call)
	avgBatchSize := &dsl.InternalCallExpr{
		FuncName: "CalculateBatcherAvgSize",
		Args:     []dsl.Expr{b.Policy, b.BatchSize, b.Timeout, b.ArrivalRate},
	}
	// Ceil it? VM handles this internal calc
	effectiveBatchSize := &dsl.InternalCallExpr{
		FuncName: "CeilInt",
		Args:     []dsl.Expr{avgBatchSize},
	}

	// 3. Get Downstream Processing AST
	downstream := b.DownstreamProc.ProcessBatch(effectiveBatchSize)

	// 4. Combine Wait + Downstream
	return &dsl.AndExpr{Left: waitAsAccessResult, Right: downstream}
}

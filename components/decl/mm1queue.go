package decl

import "github.com/panyam/sdl/dsl"

// --- MM1Queue ---
type MM1Queue struct {
	Name           string
	ArrivalRate    dsl.Expr // Lambda (FLOAT)
	AvgServiceTime dsl.Expr // Ts (Duration Expr)
}

func NewMM1Queue(name string, lambda, ts dsl.Expr) *MM1Queue {
	return &MM1Queue{Name: name, ArrivalRate: lambda, AvgServiceTime: ts}
}

// Enqueue - Simple, often near-zero cost
func (q *MM1Queue) Enqueue() dsl.Expr {
	// VM can return a fixed near-zero latency success outcome
	return &dsl.InternalCallExpr{FuncName: "MM1Enqueue", Args: []dsl.Expr{}}
}

// Dequeue - Returns Outcomes[Duration] representing wait time Wq
func (q *MM1Queue) Dequeue() dsl.Expr {
	// VM calculates Wq based on lambda/Ts and returns distribution
	return &dsl.InternalCallExpr{
		FuncName: "MM1CalculateWq", // VM knows this returns Outcomes[Duration]
		Args: []dsl.Expr{
			q.ArrivalRate,
			q.AvgServiceTime,
		},
	}
}

package decl

import "github.com/panyam/sdl/dsl"

// --- Queue (MMCK) ---
type Queue struct {
	Name           string
	ArrivalRate    dsl.Expr // Lambda (FLOAT)
	AvgServiceTime dsl.Expr // Ts (Duration Expr)
	Servers        dsl.Expr // c (INT)
	Capacity       dsl.Expr // K (INT, 0 for infinite)
}

func NewQueue(name string, lambda, ts, servers, capacity dsl.Expr) *Queue {
	return &Queue{Name: name, ArrivalRate: lambda, AvgServiceTime: ts, Servers: servers, Capacity: capacity}
}

// Enqueue - Returns Outcomes[AccessResult] with Success=false if blocked (Pk > 0)
func (q *Queue) Enqueue() dsl.Expr {
	// VM calculates Pk based on MMc(K) params and returns probabilistic outcome
	return &dsl.InternalCallExpr{
		FuncName: "MMCKEnqueue", // VM knows this returns Outcomes[AccessResult]
		Args: []dsl.Expr{
			q.ArrivalRate,
			q.AvgServiceTime,
			q.Servers,
			q.Capacity,
		},
	}
}

// Dequeue - Returns Outcomes[Duration] representing MMc(K) wait time Wq
func (q *Queue) Dequeue() dsl.Expr {
	// VM calculates Wq based on MMc(K) params and returns distribution
	return &dsl.InternalCallExpr{
		FuncName: "MMCKCalculateWq", // VM knows this returns Outcomes[Duration]
		Args: []dsl.Expr{
			q.ArrivalRate,
			q.AvgServiceTime,
			q.Servers,
			q.Capacity,
		},
	}
}

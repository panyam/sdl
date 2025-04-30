package decl

import "github.com/panyam/leetcoach/sdl/dsl"

// --- ResourcePool (Stateless) ---
type ResourcePool struct {
	Name        string
	Size        dsl.Expr // c (INT)
	ArrivalRate dsl.Expr // Lambda (FLOAT)
	AvgHoldTime dsl.Expr // Ts (Duration Expr)
}

func NewResourcePool(name string, size, lambda, ts dsl.Expr) *ResourcePool {
	return &ResourcePool{Name: name, Size: size, ArrivalRate: lambda, AvgHoldTime: ts}
}

// Acquire predicts queueing delay based on MMc model.
func (rp *ResourcePool) Acquire() dsl.Expr {
	// VM calculates Wq distribution or rejection based on params
	return &dsl.InternalCallExpr{
		FuncName: "PoolAcquireMMc", // Returns Outcomes[AccessResult]
		Args: []dsl.Expr{
			rp.Size,
			rp.ArrivalRate,
			rp.AvgHoldTime,
		},
	}
}

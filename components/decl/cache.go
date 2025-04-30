package decl

import "github.com/panyam/leetcoach/sdl/dsl"

// --- Cache ---
type Cache struct {
	Name string
	// Config parameters as expressions
	HitRate        dsl.Expr // Float (0-1)
	HitLatency     dsl.Expr // Outcomes[Duration]
	MissLatency    dsl.Expr // Outcomes[Duration]
	WriteLatency   dsl.Expr // Outcomes[Duration]
	FailureProb    dsl.Expr // Float (0-1)
	FailureLatency dsl.Expr // Outcomes[Duration]
}

func NewCache(name string, hr, hLat, mLat, wLat, fp, fLat dsl.Expr) *Cache {
	return &Cache{Name: name, HitRate: hr, HitLatency: hLat, MissLatency: mLat, WriteLatency: wLat, FailureProb: fp, FailureLatency: fLat}
}

// Read describes the probabilistic read operation. VM handles branching.
func (c *Cache) Read() dsl.Expr {
	// VM interprets this call, using HitRate, HitLatency, MissLatency etc.
	// to construct the final combined AccessResult distribution.
	return &dsl.InternalCallExpr{
		FuncName: "CalculateCacheRead",
		Args: []dsl.Expr{
			c.HitRate,
			c.HitLatency,
			c.MissLatency,
			c.FailureProb,
			c.FailureLatency,
		},
	}
}

// Write describes the probabilistic write operation. VM handles branching.
func (c *Cache) Write() dsl.Expr {
	// VM interprets this call, using WriteLatency, FailureProb etc.
	return &dsl.InternalCallExpr{
		FuncName: "CalculateCacheWrite",
		Args: []dsl.Expr{
			c.WriteLatency,
			c.FailureProb,
			c.FailureLatency,
		},
	}
}

package decl

import (
	"strconv"

	"github.com/panyam/sdl/dsl"
)

// --- NetworkLink ---

type NetworkLink struct {
	Name           string
	BaseLatency    dsl.Expr // Expr yielding Duration
	MaxJitter      dsl.Expr // Expr yielding Duration
	PacketLossProb dsl.Expr // Expr yielding Float (0-1)
	LatencyBuckets int      // Fixed number for now
}

func NewNetworkLink(name string, baseLat, maxJitter, packetLoss dsl.Expr, buckets int) *NetworkLink {
	if buckets <= 0 {
		buckets = 5
	} // Default
	return &NetworkLink{
		Name:           name,
		BaseLatency:    baseLat,
		MaxJitter:      maxJitter,
		PacketLossProb: packetLoss,
		LatencyBuckets: buckets,
	}
}

// Transfer returns an  expression instructing the VM to generate
// the network transfer outcome based on the link's parameters.
func (nl *NetworkLink) Transfer() dsl.Expr {
	return &dsl.InternalCallExpr{
		FuncName: "GenerateNetworkOutcome",
		Args: []dsl.Expr{
			nl.BaseLatency,
			nl.MaxJitter,
			nl.PacketLossProb,
			&dsl.LiteralExpr{Kind: "INT", Value: strconv.Itoa(nl.LatencyBuckets)},
		},
	}
}

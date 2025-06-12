package decl

import (
	"github.com/panyam/sdl/components"
	"github.com/panyam/sdl/core"
	"github.com/panyam/sdl/decl"
)

// --- ResourcePool (Stateless) ---
type ResourcePool struct {
	NWBase[*components.ResourcePool]
}

func NewResourcePool(name string) *ResourcePool {
	return &ResourcePool{NWBase: NewNWBase(name, components.NewResourcePool(name))}
}

// Acquire predicts queueing delay based on MMc model.
// Returns Outcomes[Bool] with queuing delay embedded in the Time field.
func (b *ResourcePool) Acquire() decl.Value {
	
	outcomes := b.Wrapped.Acquire()
	
	// Convert AccessResult outcomes to boolean outcomes with embedded latency
	boolOutcomes := &core.Outcomes[decl.Value]{}
	
	for _, bucket := range outcomes.Buckets {
		boolVal := decl.BoolValue(bucket.Value.Success)
		boolVal.Time = bucket.Value.Latency // Embed queuing delay in Time field
		boolOutcomes.Add(bucket.Weight, boolVal)
		
		// Debug removed for cleaner output
	}
	
	outType := decl.OutcomesType(decl.BoolType)
	v, e := decl.NewValue(outType, boolOutcomes)
	if e != nil {
		panic(e)
	}
	return v
}

package decl

import "github.com/panyam/leetcoach/sdl/dsl"

// --- LSMTree ---
type LSMTree struct {
	IndexBase
	// LSM Specific Params (as Expr)
	MemtableHitProb      dsl.Expr
	Level0HitProb        dsl.Expr
	Levels               dsl.Expr // INT
	ReadAmpFactor        dsl.Expr // FLOAT
	WriteAmpFactor       dsl.Expr // FLOAT
	CompactionImpactProb dsl.Expr // FLOAT
	CompactionSlowdown   dsl.Expr // Outcomes[Duration]
}

func NewLSMTree(base IndexBase, mtHit, l0Hit, levels, rAmp, wAmp, compProb, compSlow dsl.Expr) *LSMTree {
	return &LSMTree{IndexBase: base, MemtableHitProb: mtHit, Level0HitProb: l0Hit, Levels: levels, ReadAmpFactor: rAmp, WriteAmpFactor: wAmp, CompactionImpactProb: compProb, CompactionSlowdown: compSlow}
}

// Read for LSMTree - complex path, represent as internal call
func (lsm *LSMTree) Read() dsl.Expr {
	// VM interprets this, performing the probabilistic path combination internally
	return &dsl.InternalCallExpr{
		FuncName: "CalculateLSMRead",
		Args: []dsl.Expr{
			lsm.diskRead(), // Base disk read profile needed
			lsm.RecordProcessingTime,
			lsm.MemtableHitProb,
			lsm.Level0HitProb,
			lsm.Levels,
			lsm.ReadAmpFactor,
			lsm.CompactionImpactProb,
			lsm.CompactionSlowdown,
		},
	}
}

// Write for LSMTree - complex path, represent as internal call
func (lsm *LSMTree) Write() dsl.Expr {
	// VM interprets this, calculating base cost, WA, compaction impact
	return &dsl.InternalCallExpr{
		FuncName: "CalculateLSMWrite",
		Args: []dsl.Expr{
			lsm.diskWrite(), // Base disk write profile needed
			lsm.RecordProcessingTime,
			lsm.WriteAmpFactor,
			lsm.CompactionImpactProb,
			lsm.CompactionSlowdown,
		},
	}
}

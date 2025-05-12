package bitly

import (
	"fmt"

	sdlc "github.com/panyam/sdl/components"
	sdl "github.com/panyam/sdl/core"
)

// "log"

// --- Result Types (Simplified for now) ---
// In a real model, these might carry more specific info (URL, error code)
// We map them to AccessResult where Success indicates the primary goal achieved.

// type ShortenResult struct { Success bool; Latency Duration; ShortURL string }
// type RedirectResult struct { Success bool; Latency Duration; LongURL string }

// BitlyService orchestrates the URL shortening and redirection logic.
type BitlyService struct {
	Name string

	// Dependencies
	IDGen *IDGenerator       // Dependency on ID Generator
	Cache *sdlc.Cache        // Dependency on Cache
	DB    *DatabaseComponent // Dependency on Database

	// Optional: Resource Pool for internal processing/concurrency limiting
	// WorkerPool *ResourcePool
}

// Init initializes the BitlyService with its dependencies.
func (bs *BitlyService) Init(name string, idGen *IDGenerator, cache *sdlc.Cache, db *DatabaseComponent) *BitlyService {
	bs.Name = name
	if idGen == nil || cache == nil || db == nil {
		panic(fmt.Sprintf("BitlyService '%s' initialized with nil dependencies", name))
	}
	bs.IDGen = idGen
	bs.Cache = cache
	bs.DB = db
	return bs
}

// --- Redirect Operation ---
func (bs *BitlyService) Redirect(shortCode string, lambda float64 /* unused for now */) *sdl.Outcomes[sdl.AccessResult] {
	// This implements the logic: Cache Read -> If Miss -> DB Read -> Optional Cache Write

	// --- Revised Manual Composition for Redirect ---
	cacheReadOutcomes := bs.Cache.Read()
	cacheHits, cacheMissesFailures := cacheReadOutcomes.Split(sdl.AccessResult.IsSuccess)

	// Path 1: Cache Hit -> Final Success=true, Latency=CacheHitLatency
	finalHits := sdl.Map(cacheHits, func(cr sdl.AccessResult) sdl.AccessResult {
		return sdl.AccessResult{Success: true, Latency: cr.Latency} // Map cache hit to overall success
	})

	// Path 2: Cache Miss/Failure -> DB Read
	dbReadOutcomes := bs.DB.GetLongURL(shortCode)
	// Combine miss/fail latency with db read latency/success
	finalMissPathOutcomes := sdl.And(cacheMissesFailures, dbReadOutcomes,
		func(cacheMissFailRes sdl.AccessResult, dbReadRes sdl.AccessResult) sdl.AccessResult {
			// Final success depends ONLY on DB read success in this path.
			// Final latency is Cache Miss/Fail Latency + DB Read Latency.
			return sdl.AccessResult{
				Success: dbReadRes.Success,
				Latency: cacheMissFailRes.Latency + dbReadRes.Latency,
			}
		})

	// Combine results from both paths
	// Initialize empty with the correct reducer, then append results from the two paths
	combinedRedirectOutcomes := &sdl.Outcomes[sdl.AccessResult]{And: sdl.AndAccessResults} // Ensure the final combined outcome has the correct 'And' func if it were used later
	combinedRedirectOutcomes.Append(finalHits)
	combinedRedirectOutcomes.Append(finalMissPathOutcomes)

	// Apply Reduction
	maxLen := 10 // TODO: Get from config?
	trimmer := sdl.TrimToSize(100, maxLen)
	successes, failures := combinedRedirectOutcomes.Split(sdl.AccessResult.IsSuccess)
	trimmedSuccesses := trimmer(successes)
	trimmedFailures := trimmer(failures)
	finalTrimmedOutcomes := (&sdl.Outcomes[sdl.AccessResult]{}).Append(trimmedSuccesses, trimmedFailures)

	return finalTrimmedOutcomes
}

// --- ShortenURL Operation ---
func (bs *BitlyService) ShortenURL(longURL string, lambda float64 /* unused */) *sdl.Outcomes[sdl.AccessResult] {
	// Simplified V1: Generate ID -> Save to DB
	// Assumes ID generation is reliable and doesn't collide for now.
	// Ignores optional check for existing longURL, optional cache write.

	idGenOutcome := bs.IDGen.GenerateID()
	// Map ID Gen failure to overall failure
	idGenMapped := sdl.Map(idGenOutcome, func(idRes sdl.AccessResult) sdl.AccessResult { return idRes }) // Pass through

	dbSaveOutcome := bs.DB.SaveMapping("short", longURL) // Pass dummy short code

	// Combine: ID Gen -> DB Save
	// If ID Gen fails, the whole op fails with ID gen latency.
	// If ID Gen succeeds, proceed to DB Save. Success/Latency determined by DB Save.
	finalOutcome := sdl.And(idGenMapped, dbSaveOutcome,
		func(idRes sdl.AccessResult, dbRes sdl.AccessResult) sdl.AccessResult {
			// Final success requires BOTH to succeed (although we assume ID gen mostly succeeds)
			// Final latency is sum.
			return sdl.AccessResult{
				Success: idRes.Success && dbRes.Success,
				Latency: idRes.Latency + dbRes.Latency,
			}
		})

	// Apply Reduction
	maxLen := 10
	trimmer := sdl.TrimToSize(100, maxLen)
	successes, failures := finalOutcome.Split(sdl.AccessResult.IsSuccess)
	trimmedSuccesses := trimmer(successes)
	trimmedFailures := trimmer(failures)
	finalTrimmedOutcomes := (&sdl.Outcomes[sdl.AccessResult]{}).Append(trimmedSuccesses, trimmedFailures)

	return finalTrimmedOutcomes
}

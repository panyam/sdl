package bitly

import (
	"fmt"

	sdlc "github.com/panyam/leetcoach/sdl/components"
	sdl "github.com/panyam/leetcoach/sdl/core"
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

	/*
		cacheReadOutcome := bs.Cache.Read()

		// Define the "Cache Miss" path: Read DB, potentially write cache
		dbReadOutcome := bs.DB.GetLongURL(shortCode) // Assume GetLongURL returns sdl.AccessResult

		// Optional: Model cache write-back after DB read success
		// For simplicity, let's assume write-back is fast/async and doesn't add
		// significantly to the latency *of this read path*. We just use the DB result.
		// More complex: cacheWriteOutcome := bs.Cache.Write()
		//              dbReadAndCacheWrite := And(dbReadOutcome, cacheWriteOutcome, ...)

		// Use If to combine paths
		// If CacheRead Success (Hit) -> return CacheRead outcome latency
		// If CacheRead Failure (Miss/Error) -> return DBRead outcome
			finalOutcome := cacheReadOutcome.If(
				func(cr sdl.AccessResult) bool { return cr.Success }, // Condition: Cache Hit?
				// THEN (Cache Hit): Return the cache hit outcome directly
				// Need to map Cache Hit (Success=true) to overall Redirect Success (Success=true)
				sdl.Map(cacheReadOutcome, func(cr sdl.AccessResult) sdl.AccessResult {
					// Assuming cache hit means overall success for redirect
					if cr.Success {
						return sdl.AccessResult{Success: true, Latency: cr.Latency}
					}
					// This path shouldn't be taken if condition is Success=true, but handle defensively
					return sdl.AccessResult{Success: false, Latency: cr.Latency}
				}),

				// OTHERWISE (Cache Miss or Cache Failure): Execute the DB Read path
				// The outcome here represents the result after reading DB (and potentially failing)
				sdl.Map(dbReadOutcome, func(dr sdl.AccessResult) sdl.AccessResult {
					// The success of the redirect now depends on the DB read success
					return dr // Return DB result as is (Success, Latency)
				}),

				// Reducer for combining outcomes from the initial step (cacheRead)
				// with the outcomes of the chosen branch (cache hit or db read path).
				// The key is how latency adds up and how success is determined.
				func(cacheRes sdl.AccessResult, branchRes sdl.AccessResult) sdl.AccessResult {
					// If cacheRes indicated HIT (cacheRes.Success=true), the branchRes IS the cache hit result.
					// If cacheRes indicated MISS/FAIL (cacheRes.Success=false), the branchRes IS the db read result.
					// In both cases, the branchRes already holds the correct final Success/Latency
					// for that specific path through the If statement. We just need to pass it through.
					// However, the latency of the initial cache check (cacheRes.Latency) *always* happened.
					// So, total latency = cacheRes.Latency + branchRes.Latency (where branch latency is 0 for cache hit)
					// But wait - the Map functions *already* produced outcomes with the correct total latency for each path...
					// Let's rethink the 'If' composition slightly.

					// --- Revised 'If' Logic ---
					// Let 'If' handle the branching probability based on cacheRes.Success.
					// The THEN branch outcome IS the cache hit result.
					// The OTHERWISE branch outcome IS the DB read result.
					// We need to combine the *initial* cache read latency with *both* branches.

					// Let's simplify: Model the sequential flow more directly.
					// We'll manually combine outcomes based on cache hit/miss probability.

					// This composition logic is getting tricky without a DSL or better combinators!
					// Let's try manual construction.

					// Alternative Manual Composition:
					// 1. Get Cache Read outcomes.
					// 2. Split into Hits (Success=true) and Misses/Fails (Success=false).
					// 3. For Misses/Fails path: AND the Miss/Fail outcome with DB Read outcome.
					// 4. Append the Hits path outcomes and the (Miss/Fail -> DB Read) path outcomes.

					// See revised implementation below...

					// This placeholder reducer isn't quite right for this flow.
					return branchRes // Placeholder
				},
			)
	*/

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

package sdl

import (
	"testing"
	// Ensure metrics helpers are accessible if needed later
)

func TestResourcePool_Init(t *testing.T) {
	rp := NewResourcePool("TestPool", 5)
	if rp.Name != "TestPool" {
		t.Error("Name mismatch")
	}
	if rp.Size != 5 {
		t.Errorf("Size mismatch: expected 5, got %d", rp.Size)
	}
	if rp.Used != 0 {
		t.Errorf("Initial Used mismatch: expected 0, got %d", rp.Used)
	}

	rpZero := NewResourcePool("ZeroPool", 0)
	if rpZero.Size != 0 {
		t.Errorf("Zero Size mismatch: expected 0, got %d", rpZero.Size)
	}
}

func TestResourcePool_Acquire_Basic(t *testing.T) {
	rp := NewResourcePool("TestPool", 2)

	// Attempt 1: Should succeed
	out1 := rp.Acquire()
	if out1 == nil || out1.Len() != 1 {
		t.Fatal("Acquire(1) failed to return outcomes")
	}
	res1 := out1.Buckets[0].Value
	if !res1.Success || res1.Pool != rp {
		t.Errorf("Acquire(1) failed: Success=%v, Pool=%p", res1.Success, res1.Pool)
	}
	// --- State update (Manual for test) ---
	// In a real simulation this is the tricky part. Manually update for test validation.
	if res1.Success {
		rp.Used++
	}

	// Attempt 2: Should succeed
	out2 := rp.Acquire()
	if out2 == nil || out2.Len() != 1 {
		t.Fatal("Acquire(2) failed to return outcomes")
	}
	res2 := out2.Buckets[0].Value
	if !res2.Success || res2.Pool != rp {
		t.Errorf("Acquire(2) failed: Success=%v, Pool=%p", res2.Success, res2.Pool)
	}
	// --- State update (Manual for test) ---
	if res2.Success {
		rp.Used++
	}
	if rp.Used != 2 {
		t.Errorf("Expected Used=2 after 2 acquires, got %d", rp.Used)
	}

	// Attempt 3: Should fail (Pool full)
	out3 := rp.Acquire()
	if out3 == nil || out3.Len() != 1 {
		t.Fatal("Acquire(3) failed to return outcomes")
	}
	res3 := out3.Buckets[0].Value
	if res3.Success {
		t.Errorf("Acquire(3) should have failed (Success=true)")
	}
	if res3.Pool != rp {
		t.Errorf("Acquire(3) failed Pool pointer mismatch")
	}
	// --- State update (Manual for test) ---
	// Used count doesn't change on failure
	if rp.Used != 2 {
		t.Errorf("Expected Used=2 after failed acquire, got %d", rp.Used)
	}

	// Attempt 4: Release one, then acquire should succeed
	rp.Release() // Manually release one
	if rp.Used != 1 {
		t.Errorf("Expected Used=1 after release, got %d", rp.Used)
	}

	out4 := rp.Acquire()
	if out4 == nil || out4.Len() != 1 {
		t.Fatal("Acquire(4) failed to return outcomes")
	}
	res4 := out4.Buckets[0].Value
	if !res4.Success || res4.Pool != rp {
		t.Errorf("Acquire(4) after release failed: Success=%v, Pool=%p", res4.Success, res4.Pool)
	}
	// --- State update (Manual for test) ---
	if res4.Success {
		rp.Used++
	}
	if rp.Used != 2 {
		t.Errorf("Expected Used=2 after acquire post-release, got %d", rp.Used)
	}

}

func TestResourcePool_Release_Empty(t *testing.T) {
	rp := NewResourcePool("TestPool", 2)
	// Should not go below zero or panic
	rp.Release()
	if rp.Used != 0 {
		t.Errorf("Release on empty pool should leave Used=0, got %d", rp.Used)
	}
	rp.Release() // Call again
	if rp.Used != 0 {
		t.Errorf("Second Release on empty pool should leave Used=0, got %d", rp.Used)
	}
}

// Test using the simplified composition (ignoring Release state management)
func TestResourcePool_SimplifiedUsage(t *testing.T) {
	pool := NewResourcePool("TestPool", 1) // Pool size 1
	// Define some simple work with its own outcomes
	workOutcomes := (&Outcomes[AccessResult]{}).
		Add(0.9, AccessResult{true, Millis(10)}).
		Add(0.1, AccessResult{false, Millis(5)})

	workFunc := func() *Outcomes[AccessResult] { return workOutcomes }

	// --- Simulate Op 1 ---
	// Simplified Acquire mapping
	acq1Out := Map(pool.Acquire(), func(aq AcquireAttemptResult) AccessResult {
		return AccessResult{Success: aq.Success, Latency: 0}
	})
	// Manually update state *after* checking outcome for test validation
	if acq1Out.Buckets[0].Value.Success {
		pool.Used++
	}

	res1 := And(acq1Out, workFunc(), AndAccessResults)
	// Note: Release is NOT called here

	t.Logf("Op1 Result (Pool Size 1, Used=1): %d buckets", res1.Len())
	// Expected: Should match workOutcomes as acquire succeeded
	if !approxEqualTest(Availability(res1), Availability(workOutcomes), 1e-9) {
		t.Errorf("Op1 Avail mismatch")
	}

	// --- Simulate Op 2 (Pool should be full) ---
	acq2Out := Map(pool.Acquire(), func(aq AcquireAttemptResult) AccessResult {
		return AccessResult{Success: aq.Success, Latency: 0}
	})
	// Manually check state for test validation
	if acq2Out.Buckets[0].Value.Success {
		t.Error("Acquire(2) should have failed as pool is full (Used=1, Size=1)")
		pool.Used++ // Update anyway if test failed here
	}

	res2 := And(acq2Out, workFunc(), AndAccessResults)
	t.Logf("Op2 Result (Pool Size 1, Used=1): %d buckets", res2.Len())
	// Expected: Should be 100% failure as acquire failed
	if !approxEqualTest(Availability(res2), 0.0, 1e-9) {
		t.Errorf("Op2 Avail mismatch: Expected 0.0, got %.4f", Availability(res2))
	}

	// Test assumes sequential calls; pool state persists between ops in the test.
}

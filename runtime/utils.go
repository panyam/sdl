package runtime

import (
	"log"
	"sync"
	"time"

	"github.com/panyam/sdl/core"
)

func RunCallInBatches(system *SystemInstance, obj, method string, nbatches, batchsize int, numworkers int, onBatch func(batch int, batchVals []Value)) (results [][]Value) {
	fi := system.File
	se := NewSimpleEval(fi, nil)
	var totalSimTime core.Duration
	var simTimeMutex sync.Mutex

	// Use the existing system environment if available, otherwise create new one
	var env *Env[Value]
	if system.Env != nil {
		env = system.Env
	} else {
		env = fi.Env()
		se.EvalInitSystem(system, env, &totalSimTime)
	}

	startTime := time.Now()
	ncalls := nbatches * batchsize
	defer func() {
		log.Printf("Total Simulation Time: %v", totalSimTime)
		log.Printf("Wall Clock Time for %d calls: %v", ncalls, time.Since(startTime))
	}()

	if nbatches < numworkers {
		numworkers = nbatches
	}

	var wg sync.WaitGroup
	batchesPerWorker := (nbatches + numworkers - 1) / numworkers

	for i := range numworkers {
		wg.Add(1)
		go func(workerIndex int) {
			defer wg.Done()
			workerEnv := env.Push() // Each worker gets its own environment to avoid data races
			workerSE := NewSimpleEval(fi, nil)
			var workerSimTime core.Duration

			startBatch := workerIndex * batchesPerWorker
			endBatch := min((workerIndex+1)*batchesPerWorker, nbatches)
			// log.Printf("Starting worker %d, Batch Range: %d -> %d", workerIndex, startBatch, endBatch)

			for batch := startBatch; batch < endBatch; batch++ {
				var batchVals []Value
				// For simulations, we don't advance a single shared clock.
				// Each run is independent. We capture the latency of each run.
				for range batchsize {
					var runLatency core.Duration
					ce := &CallExpr{Function: &MemberAccessExpr{Receiver: &IdentifierExpr{Value: obj}, Member: &IdentifierExpr{Value: method}}}
					res, _ := workerSE.Eval(ce, workerEnv, &runLatency) // a fresh runLatency for each call
					res.Time = runLatency                               // The latency is the duration of this single run
					workerSimTime += runLatency                          // Accumulate worker's simulation time
					batchVals = append(batchVals, res)
				}
				results = append(results, batchVals)
				if onBatch != nil {
					onBatch(batch, batchVals)
				}
			}

			// Add worker's total simulation time to the global total
			simTimeMutex.Lock()
			totalSimTime += workerSimTime
			simTimeMutex.Unlock()
		}(i)
	}

	wg.Wait()
	return
}

func RunTestCall(system *SystemInstance, env *Env[Value], obj, method string, ncalls int) {
	startTime := time.Now()
	defer func() {
		log.Printf("Time taken for %d calls: %v", ncalls, time.Now().Sub(startTime))
	}()
	se := NewSimpleEval(system.File, nil)
	log.Printf("Now Running %s.%s.%s:", system.System.Name.Value, obj, method)
	for i := range ncalls {
		var currTime core.Duration
		ce := &CallExpr{Function: &MemberAccessExpr{Receiver: &IdentifierExpr{Value: obj}, Member: &IdentifierExpr{Value: method}}}
		res2, ret2 := se.Eval(ce, env, &currTime) // reuse env to continue
		if i == ncalls-1 {
			log.Printf("Running %d, Result: %v, Ret: %v, Time: %v", i, res2, ret2, currTime)
		}
	}
}

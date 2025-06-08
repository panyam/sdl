package runtime

import (
	"log"
	"sync"
	"time"

	"github.com/panyam/sdl/core"
)

func RunCallInBatches(system *SystemInstance, obj, method string, nbatches, batchsize int, numworkers int, onBatch func(batch int, batchVals []Value)) (results [][]Value) {
	fi := system.File
	se := NewSimpleEval(fi)
	var currTime core.Duration
	env := fi.Env.Push()
	se.EvalInitSystem(system, env, &currTime)

	startTime := time.Now()
	ncalls := nbatches * batchsize
	defer func() {
		log.Printf("Time taken for %d calls: %v", ncalls, time.Now().Sub(startTime))
	}()
	if nbatches < numworkers {
		numworkers = nbatches
	}

	var wg sync.WaitGroup
	batchesPerWorker := int(nbatches / numworkers)
	if nbatches%numworkers != 0 {
		batchesPerWorker += 1
	}
	for i := range numworkers {
		wg.Add(1)
		go func(workerIndex int) {
			startBatch := workerIndex * batchesPerWorker
			endBatch := min((workerIndex+1)*batchesPerWorker, nbatches)
			log.Printf("Starting worker %d, Batch Range: %d -> %d", workerIndex, startBatch, endBatch)
			for batch := startBatch; batch < endBatch; batch += 1 {
				var batchVals []Value
				for range batchsize {
					var currTime core.Duration
					ce := &CallExpr{Function: &MemberAccessExpr{Receiver: &IdentifierExpr{Value: obj}, Member: &IdentifierExpr{Value: method}}}
					res, _ := se.Eval(ce, env, &currTime) // reuse env to continue
					res.Time = currTime
					batchVals = append(batchVals, res)
				}
				results = append(results, batchVals)
				if onBatch != nil {
					onBatch(batch, batchVals)
				}
			}
			wg.Done()
		}(i)
	}

	// Wait for all workers to finish
	wg.Wait()
	return
}

func RunTestCall(system *SystemInstance, env *Env[Value], obj, method string, ncalls int) {
	startTime := time.Now()
	defer func() {
		log.Printf("Time taken for %d calls: %v", ncalls, time.Now().Sub(startTime))
	}()
	se := NewSimpleEval(system.File)
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

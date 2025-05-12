// sdl/decl/components_test.go
package decl

import (
	"strconv"
	"strings"
	"testing"

	"github.com/panyam/sdl/dsl"
)

// --- Test Helper Functions ---

// Helper to compare  string representations
func assertASTMatches(t *testing.T, caseName string, expected, actual dsl.Expr) {
	t.Helper()
	expStr := expected.String()
	actStr := actual.String()
	if expStr != actStr {
		// Replace common problematic characters for easier diffing in test output if needed
		replacer := strings.NewReplacer("(", "{", ")", "}", " ", "_")
		t.Errorf("[%s]  mismatch:\nExpected: %s\nActual:   %s\nExpected (No Space): %s\nActual (No Space):   %s",
			caseName,
			expStr, actStr,
			replacer.Replace(expStr), replacer.Replace(actStr)) // Added No Space version
	}
}

// Helpers to create  nodes easily in tests
func litStr(v string) dsl.Expr { return &dsl.LiteralExpr{Kind: "STRING", Value: v} }
func litInt(v int) dsl.Expr    { return &dsl.LiteralExpr{Kind: "INT", Value: strconv.Itoa(v)} }
func litFloat(v float64) dsl.Expr {
	return &dsl.LiteralExpr{Kind: "FLOAT", Value: strconv.FormatFloat(v, 'f', -1, 64)}
}
func litBool(v bool) dsl.Expr                    { return &dsl.LiteralExpr{Kind: "BOOL", Value: strconv.FormatBool(v)} }
func ident(n string) dsl.Expr                    { return &dsl.IdentifierExpr{Name: n} }
func member(r dsl.Expr, m string) dsl.Expr       { return &dsl.MemberAccessExpr{Receiver: r, Member: m} }
func call(f dsl.Expr, args ...dsl.Expr) dsl.Expr { return &dsl.CallExpr{Function: f, Args: args} }
func and(l, r dsl.Expr) dsl.Expr                 { return &dsl.AndExpr{Left: l, Right: r} }
func parallel(l, r dsl.Expr) dsl.Expr            { return &dsl.ParallelExpr{Left: l, Right: r} }
func internalCall(fName string, args ...dsl.Expr) dsl.Expr {
	return &dsl.InternalCallExpr{FuncName: fName, Args: args}
}

// --- Test Cases ---

func TestDeclDisk(t *testing.T) {
	diskSSD := NewDisk("ssd1", "SSD")
	diskHDD := NewDisk("hdd1", "HDD")
	procTime := ident("processingTime") // Assume this var exists in VM context

	// Test ReadAST
	expectedReadSSD := internalCall("GetDiskReadProfile", litStr("SSD"))
	assertASTMatches(t, "SSD Read", expectedReadSSD, diskSSD.Read())

	expectedReadHDD := internalCall("GetDiskReadProfile", litStr("HDD"))
	assertASTMatches(t, "HDD Read", expectedReadHDD, diskHDD.Read())

	// Test WriteAST
	expectedWriteSSD := internalCall("GetDiskWriteProfile", litStr("SSD"))
	assertASTMatches(t, "SSD Write", expectedWriteSSD, diskSSD.Write())

	// Test ReadProcessWriteAST
	expectedRPW := and(and(diskSSD.Read(), procTime), diskSSD.Write())
	assertASTMatches(t, "SSD RPW", expectedRPW, diskSSD.ReadProcessWrite(procTime))
}

func TestDeclNetworkLink(t *testing.T) {
	baseLat := litFloat(0.01) // 10ms
	jitter := litFloat(0.002) // 2ms
	loss := litFloat(0.01)    // 1%
	buckets := 5

	link := NewNetworkLink("net1", baseLat, jitter, loss, buckets)

	expected := internalCall("GenerateNetworkOutcome",
		baseLat,
		jitter,
		loss,
		litInt(buckets),
	)
	assertASTMatches(t, "Network Transfer", expected, link.Transfer())
}

func TestDeclBTreeIndex(t *testing.T) {
	// Setup base index config using identifiers (VM resolves these)
	base := IndexBase{
		Name:                 "TestBTree",
		DiskName:             "disk1", // VM resolves this dependency name
		RecordProcessingTime: ident("defaultRecProcTime"),
		NumRecords:           ident("numRecordsVar"),
		RecordSize:           litInt(128),
		PageSize:             litInt(8192),
		NodeFanout:           litInt(100),
	}

	btree := NewBTreeIndex(base, litFloat(0.67), litFloat(0.1), litFloat(0.1))

	// --- Test Find ---
	// Expected structure: Internal.RepeatSequence( AND( DiskRead, NodeSearchCPU ), HeightExpr )
	heightExpr := internalCall("CalculateBTreeHeight", base.NumRecords, base.RecordSize, base.PageSize, base.NodeFanout)
	nodeSearchCpu := internalCall("CalculateNodeSearchCPU", base.RecordProcessingTime, base.NodeFanout)
	diskRead := call(member(ident(base.DiskName), "Read")) // Call Disk.Read()
	levelCost := and(diskRead, nodeSearchCpu)
	expectedFind := internalCall("RepeatSequence", levelCost, heightExpr)

	assertASTMatches(t, "BTree Find", expectedFind, btree.Find())

	// --- Test Insert ---
	// Expected: AND( AND( AND(Find, ModifyCPU), WriteLeaf), AvgPropCost)
	find := btree.Find() // Reuse Find AST
	modifyCpu := internalCall("GetRecordProcessingTime", base.RecordProcessingTime)
	writeLeaf := call(member(ident(base.DiskName), "Write"))
	// Prop cost is Internal.ScaleLatency( AND(Read, Write), Factor )
	readWritePair := and(diskRead, writeLeaf)
	avgSplitCost := internalCall("ScaleLatency", readWritePair, btree.AvgSplitPropCost)
	expectedInsert := and(and(and(find, modifyCpu), writeLeaf), avgSplitCost)

	assertASTMatches(t, "BTree Insert", expectedInsert, btree.Insert())

	// --- Test Delete (similar structure to Insert, uses AvgMergePropCost) ---
	avgMergeCost := internalCall("ScaleLatency", readWritePair, btree.AvgMergePropCost)
	expectedDelete := and(and(and(find, modifyCpu), writeLeaf), avgMergeCost)

	assertASTMatches(t, "BTree Delete", expectedDelete, btree.Delete())
}

func TestDeclHashIndex(t *testing.T) {
	base := IndexBase{
		Name:                 "TestHash",
		DiskName:             "disk2",
		RecordProcessingTime: ident("recProc"),
		NumRecords:           litInt(1000000),
		RecordSize:           litInt(100),
		PageSize:             litInt(4096),
	}
	hashIdx := NewHashIndex(base, litFloat(0.2), litFloat(1.5))

	// --- Test Find ---
	// Expected: AND( AND(HashCPU, ReadPrimary), AvgOverflowReadCost )
	hashCpu := internalCall("GetRecordProcessingTime", base.RecordProcessingTime) // Simplification
	readPrimary := call(member(ident(base.DiskName), "Read"))
	avgOverflow := internalCall("ScaleLatency", readPrimary, hashIdx.AvgOverflowReads)
	expectedFind := and(and(hashCpu, readPrimary), avgOverflow)

	assertASTMatches(t, "Hash Find", expectedFind, hashIdx.Find())

	// --- Test Insert ---
	// Expected: AND( AND( AND(Find, ModifyCPU), Write), AvgResizeCost )
	find := hashIdx.Find()
	modifyCpu := internalCall("GetRecordProcessingTime", base.RecordProcessingTime)
	write := call(member(ident(base.DiskName), "Write"))
	numPagesExpr := internalCall("CalculateNumPages", base.NumRecords, base.RecordSize, base.PageSize)
	avgResize := internalCall("CalculateResizeCost", readPrimary, write, numPagesExpr, hashIdx.ResizeCostFactor)
	expectedInsert := and(and(and(find, modifyCpu), write), avgResize)

	assertASTMatches(t, "Hash Insert", expectedInsert, hashIdx.Insert())

	// --- Test Delete ---
	// Expected: AND( AND(Find, ModifyCPU), Write)
	expectedDelete := and(and(find, modifyCpu), write)
	assertASTMatches(t, "Hash Delete", expectedDelete, hashIdx.Delete())
}

func TestDeclBitmapIndex(t *testing.T) {
	base := IndexBase{
		Name:                 "TestBitmap",
		DiskName:             "disk3",
		RecordProcessingTime: ident("recProcBitmap"),
		NumRecords:           ident("numRows"),
		RecordSize:           litInt(4), // Typically small for indexed value itself
		PageSize:             litInt(8192),
	}
	bitmapIdx := NewBitmapIndex(base, litInt(50), litBool(true), litFloat(3.0), litFloat(0.01))

	// --- Test Find ---
	// Expected: AND( AND(LoadIndex, BitwiseOps), ResultProcessing)
	loadIndex := call(member(ident(base.DiskName), "Read"))
	bitwiseOp := internalCall("CalculateBitwiseOpCost", litInt(3), bitmapIdx.IsCompressed)
	resultProcessing := internalCall("CalculateResultProcessingCost", base.NumRecords, bitmapIdx.QuerySelectivity, base.RecordProcessingTime)
	expectedFind := and(and(loadIndex, bitwiseOp), resultProcessing)

	assertASTMatches(t, "Bitmap Find", expectedFind, bitmapIdx.Find())

	// --- Test Insert/Delete/Update (modifyBitmapCost) ---
	// Expected: Internal.ScaleLatency( AND(AND(Read, ModifyCPU), Write), UpdateFactor )
	read := call(member(ident(base.DiskName), "Read"))
	modifyCpu := internalCall("GetRecordProcessingTime", base.RecordProcessingTime)
	write := call(member(ident(base.DiskName), "Write"))
	rmw := and(and(read, modifyCpu), write)
	expectedModify := internalCall("ScaleLatency", rmw, bitmapIdx.UpdateCostFactor)

	assertASTMatches(t, "Bitmap Insert", expectedModify, bitmapIdx.Insert())
	assertASTMatches(t, "Bitmap Delete", expectedModify, bitmapIdx.Delete())
	assertASTMatches(t, "Bitmap Update", expectedModify, bitmapIdx.Update())
}

func TestDeclCache(t *testing.T) {
	// Define parameters as  expressions (using literals here)
	hr := litFloat(0.8)
	hLat := ident("cacheHitLatencyProfile") // Assume VM resolves this var
	mLat := ident("cacheMissLatencyProfile")
	wLat := ident("cacheWriteLatencyProfile")
	fp := litFloat(0.001)
	fLat := ident("cacheFailLatencyProfile")

	cache := NewCache("c1", hr, hLat, mLat, wLat, fp, fLat)

	// Test ReadAST
	expectedRead := internalCall("CalculateCacheRead", hr, hLat, mLat, fp, fLat)
	assertASTMatches(t, "Cache Read", expectedRead, cache.Read())

	// Test WriteAST
	expectedWrite := internalCall("CalculateCacheWrite", wLat, fp, fLat)
	assertASTMatches(t, "Cache Write", expectedWrite, cache.Write())
}

func TestDeclHeapFile(t *testing.T) {
	base := IndexBase{
		Name:                 "TestHeap",
		DiskName:             "disk_heap",
		RecordProcessingTime: ident("procTimeHeap"),
		NumRecords:           litInt(50000),
		RecordSize:           litInt(256),
		PageSize:             litInt(4096),
	}
	heap := NewHeapFile(base)

	// Common expressions
	numPagesExpr := internalCall("CalculateNumPages", base.NumRecords, base.RecordSize, base.PageSize)
	recsPerPageExpr := internalCall("CalculateRecsPerPage", base.RecordSize, base.PageSize)
	procPageCost := internalCall("ScaleLatency", base.RecordProcessingTime, recsPerPageExpr)
	read := call(member(ident(base.DiskName), "Read"))
	write := call(member(ident(base.DiskName), "Write"))
	readAndProcessPage := and(read, procPageCost)

	// Test ScanAST
	expectedScan := &dsl.RepeatExpr{Input: readAndProcessPage, Count: numPagesExpr, Mode: dsl.Sequential}
	assertASTMatches(t, "Heap Scan", expectedScan, heap.Scan())

	// Test InsertAST
	proc := internalCall("GetRecordProcessingTime", base.RecordProcessingTime)
	expectedInsert := and(and(read, proc), write)
	assertASTMatches(t, "Heap Insert", expectedInsert, heap.Insert())

	// Test FindAST
	halfPagesExpr := internalCall("DivideInt", numPagesExpr, litInt(2))
	expectedFind := &dsl.RepeatExpr{Input: readAndProcessPage, Count: halfPagesExpr, Mode: dsl.Sequential}
	assertASTMatches(t, "Heap Find", expectedFind, heap.Find())

	// Test DeleteAST
	find := heap.Find()
	expectedDelete := and(find, write)
	assertASTMatches(t, "Heap Delete", expectedDelete, heap.Delete())
}

func TestDeclLSMTree(t *testing.T) {
	base := IndexBase{Name: "TestLSM", DiskName: "disk_lsm", RecordProcessingTime: ident("procLSM"), NumRecords: ident("nRecsLSM")}
	mtHit := litFloat(0.1)
	l0Hit := litFloat(0.3)
	levels := litInt(4)
	rAmp := litFloat(3.0)
	wAmp := litFloat(1.5)
	cProb := litFloat(0.05)
	cSlow := ident("compSlowDist")

	lsm := NewLSMTree(base, mtHit, l0Hit, levels, rAmp, wAmp, cProb, cSlow)

	readDisk := call(member(ident(base.DiskName), "Read"))
	writeDisk := call(member(ident(base.DiskName), "Write"))

	// Test ReadAST
	expectedRead := internalCall("CalculateLSMRead", readDisk, base.RecordProcessingTime, mtHit, l0Hit, levels, rAmp, cProb, cSlow)
	assertASTMatches(t, "LSM Read", expectedRead, lsm.Read())

	// Test WriteAST
	expectedWrite := internalCall("CalculateLSMWrite", writeDisk, base.RecordProcessingTime, wAmp, cProb, cSlow)
	assertASTMatches(t, "LSM Write", expectedWrite, lsm.Write())
}

func TestDeclMM1Queue(t *testing.T) {
	lambda := ident("arrivalRateVar")
	ts := ident("serviceTimeVar")
	q := NewMM1Queue("q1", lambda, ts)

	// Test EnqueueAST
	expectedEnqueue := internalCall("MM1Enqueue")
	assertASTMatches(t, "MM1 Enqueue", expectedEnqueue, q.Enqueue())

	// Test DequeueAST
	expectedDequeue := internalCall("MM1CalculateWq", lambda, ts)
	assertASTMatches(t, "MM1 Dequeue", expectedDequeue, q.Dequeue())
}

func TestDeclQueue(t *testing.T) {
	lambda := ident("q_lambda")
	ts := ident("q_ts")
	servers := litInt(4)
	capacity := litInt(100)
	q := NewQueue("q_mmck", lambda, ts, servers, capacity)

	// Test EnqueueAST
	expectedEnqueue := internalCall("MMCKEnqueue", lambda, ts, servers, capacity)
	assertASTMatches(t, "MMCK Enqueue", expectedEnqueue, q.Enqueue())

	// Test DequeueAST
	expectedDequeue := internalCall("MMCKCalculateWq", lambda, ts, servers, capacity)
	assertASTMatches(t, "MMCK Dequeue", expectedDequeue, q.Dequeue())
}

func TestDeclResourcePool(t *testing.T) {
	size := litInt(10)
	lambda := ident("pool_lambda")
	ts := ident("pool_holdtime")
	pool := NewResourcePool("rp1", size, lambda, ts)

	// Test AcquireAST
	expectedAcquire := internalCall("PoolAcquireMMc", size, lambda, ts)
	assertASTMatches(t, "Pool Acquire", expectedAcquire, pool.Acquire())
}

func TestDeclSortedFile(t *testing.T) {
	base := IndexBase{Name: "TestSorted", DiskName: "disk_sorted", RecordProcessingTime: ident("procSorted"), NumRecords: ident("nRecsSorted"), RecordSize: litInt(512), PageSize: litInt(4096)}
	sf := NewSortedFile(base)

	read := call(member(ident(base.DiskName), "Read"))
	write := call(member(ident(base.DiskName), "Write"))
	numPagesExpr := internalCall("CalculateNumPages", base.NumRecords, base.RecordSize, base.PageSize)
	recsPerPageExpr := internalCall("CalculateRecsPerPage", base.RecordSize, base.PageSize)

	// Test ScanAST
	procPageCostScan := internalCall("ScaleLatency", base.RecordProcessingTime, recsPerPageExpr)
	readAndProcessPageScan := and(read, procPageCostScan)
	expectedScan := &dsl.RepeatExpr{Input: readAndProcessPageScan, Count: numPagesExpr, Mode: dsl.Sequential}
	assertASTMatches(t, "Sorted Scan", expectedScan, sf.Scan())

	// Test FindAST
	logPagesExpr := internalCall("Log2", numPagesExpr)
	logRecsExpr := internalCall("Log2", recsPerPageExpr)
	pageSearchCpu := internalCall("ScaleLatency", base.RecordProcessingTime, logRecsExpr)
	readAndSearchPage := and(read, pageSearchCpu)
	expectedFind := &dsl.RepeatExpr{Input: readAndSearchPage, Count: logPagesExpr, Mode: dsl.Sequential}
	assertASTMatches(t, "Sorted Find", expectedFind, sf.Find())

	// Test Insert / DeleteAST
	find := sf.Find() // Use Find as built above
	shiftPagesExpr := internalCall("DivideInt", numPagesExpr, litInt(4))
	readWritePair := and(read, write)
	shiftCost := &dsl.RepeatExpr{Input: readWritePair, Count: shiftPagesExpr, Mode: dsl.Sequential}
	expectedInsertDelete := and(find, shiftCost)
	assertASTMatches(t, "Sorted Insert", expectedInsertDelete, sf.Insert())
	assertASTMatches(t, "Sorted Delete", expectedInsertDelete, sf.Delete())
}

// Mock Batch Processor for Batcher Test
type MockDeclBatchProcessor struct {
	ProcName string
}

func (p *MockDeclBatchProcessor) ProcessorName() string { return p.ProcName }
func (p *MockDeclBatchProcessor) ProcessBatch(batchSize dsl.Expr) dsl.Expr {
	// Simulate returning an  representing the batch processing call
	return internalCall("ProcessMyBatch", batchSize)
}

func TestDeclBatcher(t *testing.T) {
	policy := litStr("SizeBased")
	batchSize := litInt(100)
	timeout := litStr("0ms") // Duration literal
	arrivalRate := ident("reqLambda")
	downstream := &MockDeclBatchProcessor{ProcName: "MyMockProc"}

	batcher := NewBatcher("b1", policy, batchSize, timeout, arrivalRate, downstream)

	// Test SubmitAST
	// Expected: AND( WaitAsAccessResult, DownstreamCall )
	waitTime := internalCall("CalculateBatcherWaitTime", policy, batchSize, timeout, arrivalRate)
	waitAsAccessResult := internalCall("DurationToAccessResult", waitTime)
	avgBatchSize := internalCall("CalculateBatcherAvgSize", policy, batchSize, timeout, arrivalRate)
	effBatchSize := internalCall("CeilInt", avgBatchSize)
	downstreamResult := downstream.ProcessBatch(effBatchSize) // Call interface method
	expectedSubmit := and(waitAsAccessResult, downstreamResult)

	assertASTMatches(t, "Batcher Submit", expectedSubmit, batcher.Submit())
}

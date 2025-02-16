package sdl

import "time"

type Disk struct {
	// Attributes for a Disk component
	AccessTime       time.Duration
	DataTransferRate int64
}

type DiskAccessResult struct {
	Success bool
	Latency Value[TimeUnit]
}

func (d *Disk) Read() (out *Outcomes[DiskAccessResult]) {
	// Option 1:
	return out.Add(900, DiskAccessResult{true, Val(1, MilliSeconds)}).
		Add(90, DiskAccessResult{true, Val(10, MilliSeconds)}).
		Add(8, DiskAccessResult{true, Val(1000, MilliSeconds)}).
		Add(1, DiskAccessResult{false, Val(10, MilliSeconds)}).
		Add(1, DiskAccessResult{false, Val(50, MilliSeconds)})
}

func (d *Disk) Write() (out *Outcomes[DiskAccessResult]) {
	return out.
		Add(900, DiskAccessResult{true, Val(2, MilliSeconds)}).
		Add(90, DiskAccessResult{true, Val(20, MilliSeconds)}).
		Add(8, DiskAccessResult{true, Val(2000, MilliSeconds)}).
		Add(1, DiskAccessResult{false, Val(20, MilliSeconds)}).
		Add(1, DiskAccessResult{false, Val(100, MilliSeconds)})
}

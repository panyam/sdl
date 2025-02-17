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
	return out.Add(.9, DiskAccessResult{true, Val(1, MilliSeconds)}).
		Add(0.09, DiskAccessResult{true, Val(10, MilliSeconds)}).
		Add(0.008, DiskAccessResult{true, Val(1000, MilliSeconds)}).
		Add(0.001, DiskAccessResult{false, Val(10, MilliSeconds)}).
		Add(0.001, DiskAccessResult{false, Val(50, MilliSeconds)})
}

func (d *Disk) Write() (out *Outcomes[DiskAccessResult]) {
	return out.Add(.9, DiskAccessResult{true, Val(1, MilliSeconds)}).
		Add(0.09, DiskAccessResult{true, Val(10, MilliSeconds)}).
		Add(0.008, DiskAccessResult{true, Val(1000, MilliSeconds)}).
		Add(0.001, DiskAccessResult{false, Val(10, MilliSeconds)}).
		Add(0.001, DiskAccessResult{false, Val(50, MilliSeconds)})
}

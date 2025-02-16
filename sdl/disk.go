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

func (d *Disk) Read() (out Outcomes[DiskAccessResult]) {
	// Option 1:
	out.Add(900, DiskAccessResult{true, Val(1, MilliSeconds)})
	out.Add(90, DiskAccessResult{true, Val(1, MilliSeconds)})
	out.Add(8, DiskAccessResult{true, Val(1, MilliSeconds)})
	out.Add(1, DiskAccessResult{false, Val(1, MilliSeconds)})
	out.Add(1, DiskAccessResult{false, Val(1, MilliSeconds)})
	return out
}

func (d *Disk) Write() (out Outcomes[DiskAccessResult]) {
	out.Add(900, DiskAccessResult{true, Val(1, MilliSeconds)})
	out.Add(90, DiskAccessResult{true, Val(10, MilliSeconds)})
	out.Add(8, DiskAccessResult{true, Val(1000, MilliSeconds)})
	out.Add(1, DiskAccessResult{false, Val(10, MilliSeconds)})
	out.Add(1, DiskAccessResult{false, Val(50, MilliSeconds)})
	return out
}

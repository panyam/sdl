package sdl

import "time"

type Disk struct {
	// Attributes for a Disk component
	AccessTime       time.Duration
	DataTransferRate int64
}

func (d *Disk) Read() (out *Outcomes[AccessResult]) {
	// Option 1:
	return out.Add(.9, AccessResult{true, time.Millisecond}).
		Add(0.09, AccessResult{true, 10 * time.Millisecond}).
		Add(0.008, AccessResult{true, 1000 * time.Millisecond}).
		Add(0.001, AccessResult{false, 10 * time.Millisecond}).
		Add(0.001, AccessResult{false, 50 * time.Millisecond})
}

func (d *Disk) Write() (out *Outcomes[AccessResult]) {
	return out.Add(.9, AccessResult{true, 1 * time.Millisecond}).
		Add(0.09, AccessResult{true, 10 * time.Millisecond}).
		Add(0.008, AccessResult{true, 1000 * time.Millisecond}).
		Add(0.001, AccessResult{false, 10 * time.Millisecond}).
		Add(0.001, AccessResult{false, 50 * time.Millisecond})
}

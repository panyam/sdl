package sdl

type Disk struct {
	// Attributes for a Disk component
	AccessTime       Duration
	DataTransferRate int64
}

func (d *Disk) Read() (out *Outcomes[AccessResult]) {
	// Option 1:
	return out.Add(.9, AccessResult{true, Millis(1)}).
		Add(0.09, AccessResult{true, Millis(10)}).
		Add(0.008, AccessResult{true, Millis(1000)}).
		Add(0.001, AccessResult{false, Millis(10)}).
		Add(0.001, AccessResult{false, Millis(50)})
}

func (d *Disk) Write() (out *Outcomes[AccessResult]) {
	return out.Add(.9, AccessResult{true, Millis(1)}).
		Add(0.09, AccessResult{true, Millis(10)}).
		Add(0.008, AccessResult{true, Millis(1000)}).
		Add(0.001, AccessResult{false, Millis(10)}).
		Add(0.001, AccessResult{false, Millis(50)})
}

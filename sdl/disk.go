package sdl

type Disk struct {
	// Attributes for a Disk component
	ReadOutcomes  Outcomes[AccessResult]
	WriteOutcomes Outcomes[AccessResult]
}

// Init with defaults
func (d *Disk) Init() *Disk {
	d.ReadOutcomes.Add(.9, AccessResult{true, Millis(1)}).
		Add(0.09, AccessResult{true, Millis(1)}).
		Add(0.008, AccessResult{true, Millis(100)}).
		Add(0.001, AccessResult{false, Millis(10)}).
		Add(0.001, AccessResult{false, Millis(50)})

	d.WriteOutcomes.Add(.9, AccessResult{true, Millis(1)}).
		Add(0.09, AccessResult{true, Millis(10)}).
		Add(0.008, AccessResult{true, Millis(1000)}).
		Add(0.001, AccessResult{false, Millis(10)}).
		Add(0.001, AccessResult{false, Millis(50)})

	return d
}

func (d *Disk) Read() (out *Outcomes[AccessResult]) {
	// Option 1:
	// This is equivalent to:
	// o = ReadOutcomes.Sample()
	// Delay(o.Latency)
	// return o.Success
	return &d.ReadOutcomes
}

func (d *Disk) Write() (out *Outcomes[AccessResult]) {
	return &d.WriteOutcomes
}

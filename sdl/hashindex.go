package sdl

// HashIndex represents a hash-based index structure
type HashIndex struct {
	Index

	// Occupancy (between 0 and 1) - usually leave room in leaf pages
	// so that inserts/deletes Do not need complete shifts
	Occupancy float64
}

func (h *HashIndex) Init() *HashIndex {
	h.Index.Init()
	h.Occupancy = 0.8
	return h
}

// Insert adds a new key to the hash index
func (h *HashIndex) Insert() (out *Outcomes[AccessResult]) {
	// Inserting into a hash index can be approximated to 4 disk reads
	// TODO - More detailed on

	out = h.Disk.Read()
	out = out.If(AccessResult.IsSuccess, h.Disk.Read(), nil, AndAccessResults)
	out = out.If(AccessResult.IsSuccess, h.Disk.Read(), nil, AndAccessResults)
	out = out.If(AccessResult.IsSuccess, h.Disk.Read(), nil, AndAccessResults)
	return
}

// Find searches for a key in the hash index
func (h *HashIndex) Find() (out *Outcomes[AccessResult]) {
	// typically 2 disk reads on a Find
	// TODO - Implement more options like overflow chains etc
	out = h.Disk.Read()
	out = out.If(AccessResult.IsSuccess, h.Disk.Read(), nil, AndAccessResults)
	return
}

// Delete removes a key from the hash index
func (h *HashIndex) Delete() (out *Outcomes[AccessResult]) {
	// typically a Find followed by an update and a write
	out = h.Disk.Read()
	out = out.If(AccessResult.IsSuccess, h.Disk.Read(), nil, AndAccessResults)
	out = out.If(AccessResult.IsSuccess, h.Disk.Write(), nil, AndAccessResults)
	out = out.If(AccessResult.IsSuccess, h.Disk.Read(), nil, AndAccessResults)
	return
}

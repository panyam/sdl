package sdl

import (
	"testing"
)

func TestDiskRead(t *testing.T) {
	d := Disk{}
	dr := d.Read()
	d1 := Then(dr, dr, func(a, b AccessResult) AccessResult {
		return AccessResult{a.Success && b.Success, a.Latency + b.Latency}
	})
	Then(d1, dr, func(a, b AccessResult) AccessResult {
		return AccessResult{a.Success && b.Success, a.Latency + b.Latency}
	})
}

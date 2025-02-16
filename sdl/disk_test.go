package sdl

import (
	"log"
	"testing"
)

func TestDiskRead(t *testing.T) {
	d := Disk{}
	dr := d.Read()
	log.Println("Disk Read Outcomes: ", d.Read().Values)
	log.Println("Disk Write Outcomes: ", d.Write().Values)
	d1 := Then(dr, dr, func(a, b DiskAccessResult) DiskAccessResult {
		return DiskAccessResult{a.Success && b.Success, a.Latency.Add(b.Latency)}
	})
	log.Println("2 Disk Reads: ", len(d1.Values), d1.Values)
	d2 := Then(d1, dr, func(a, b DiskAccessResult) DiskAccessResult {
		return DiskAccessResult{a.Success && b.Success, a.Latency.Add(b.Latency)}
	})
	log.Println("3 Disk Reads: ", len(d2.Values), d2.Values)
}

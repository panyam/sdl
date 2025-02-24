package sdl

import (
	"log"
	"testing"
)

func TestDiskRead(t *testing.T) {
	d := Disk{}
	dr := d.Read()
	log.Println("Disk Read Outcomes: ", d.Read().Buckets)
	log.Println("Disk Write Outcomes: ", d.Write().Buckets)
	d1 := Then(dr, dr, func(a, b AccessResult) AccessResult {
		return AccessResult{a.Success && b.Success, a.Latency + b.Latency}
	})
	log.Println("2 Disk Reads: ", len(d1.Buckets), d1.Buckets)
	d2 := Then(d1, dr, func(a, b AccessResult) AccessResult {
		return AccessResult{a.Success && b.Success, a.Latency + b.Latency}
	})
	log.Println("3 Disk Reads: ", len(d2.Buckets), d2.Buckets)
}

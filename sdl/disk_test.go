package sdl

import (
	"log"
	"testing"
)

func TestDiskRead(t *testing.T) {
	d := Disk{}
	log.Println("Disk Read Outcomes: ", d.Read().Values)
	log.Println("Disk Write Outcomes: ", d.Write().Values)
}

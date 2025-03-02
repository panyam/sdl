package sdl

import (
	"testing"
)

func TestDiskRead(t *testing.T) {
	d := (&Disk{}).Init()
	dr := d.Read()
	dr.Then(dr, dr)
}

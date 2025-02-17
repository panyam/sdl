package sdl

import (
	"log"
	"testing"
)

func TestHeapFile(t *testing.T) {
	hf := HeapFile{}
	log.Println("HF Insert: ", hf.Init().Insert().Values)
	log.Println("=======================")
	log.Println("HF Find: ", hf.Init().Find().Values)
}

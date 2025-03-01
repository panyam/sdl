package sdl

import (
	"log"
	"testing"
)

func TestHeapFile(t *testing.T) {
	hf := (&HeapFile{}).Init()
	log.Println("HF Insert: ", hf.Insert().Buckets)
	log.Println("=======================")
	log.Println("HF Find: ", hf.Find().Buckets)
}

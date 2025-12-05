package runtime

import "fmt"

type IDGen interface {
	NextID(class string) string
}

type SimpleIDGen struct {
	counter int
}

func (s *SimpleIDGen) NextID(class string) string {
	s.counter += 1
	return fmt.Sprintf("%s:%d", class, s.counter)
}

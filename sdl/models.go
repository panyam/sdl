package sdl

// SDL is the system design language
// It allows us to:

// Declare components/services
// Declare their APIs
// For each API models a single operation on the component.
// Allow evaluation of SLOs based on how the components are wired.
// Declare interaction between components
// Note that actual work is not performed.  Values are only simulated
/*
For example to describe a database we might have:

component Database {
	Operations: {
		Read(Params): {
			Responses: {
				Success: {
					80: 1ms
					10: 2ms
					5: 10ms
					1: 1s
				}
				Failure: {
					100: 1ms	// failure happens instantaneously
				}
				Retry: {
					50: 1ms,
					50: 10s
				}
			},
		}
		Write(Params): {
			Responses: {
				Success: {
					80: 1ms
					10: 2ms
					5: 10ms
					1: 1s
				}
				Failure: {
					100: 1ms	// failure happens instantaneously
				}
				Retry: {
					50: 1ms,
					50: 10s
				}
			},
		}
	}
}

component ApiServer {
	children: {
		d: Database
	}
	operations: {
		GetUser: { d.Read(50), d.Read(10), d.Write(20) }
	}
}

component WebServer {
	children: {
		d: ApiServer
	}
	operations: {
		GetUser: { d.Read(50), d.Read(10), d.Write(20) }
	}
}

This is similar to:

[WebServer] -> [ApiServer] -> [Database]
*/

// A histogram distribution of what values can be expected in each bucket
type Distribution[V any] struct {
	Weights []int
	Values  []V
}

type Component struct {
	Name string
	Apis []Api
}

type Type struct {
}

type Api struct {
	Operation string
	Inputs    []Type
	Outputs   []Type
	Latencies Distribution[float64]
}

package types

// RunResult represents the result of a single simulation run.
// This struct is used for serialization between commands and web API.
type RunResult struct {
	Timestamp   int64   `json:"ts"`       // UnixMilli of when the run was recorded
	Latency     float64 `json:"latency"`  // End-to-end latency in milliseconds
	ResultValue string  `json:"result"`   // The string representation of the returned decl.Value
	IsError     bool    `json:"is_error"` // Whether this run resulted in an error
	ErrorString string  `json:"error,omitempty"` // Error message if IsError is true
}
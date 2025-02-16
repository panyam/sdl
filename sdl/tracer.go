package bitly

type Tracer struct {
}

// Push the outcome distribution of another method onto here
func (t *Tracer) Call(exps ...*Outcome) {
}

// Perform a conditioanl call
func (t *Tracer) If(matcher func(ch *Choice) bool, body *Outcome, otherwise *Outcome) {
}

// Add a delay
func (t *Tracer) Busy(value *Outcome) {
}

// Repeat an outcome N times
func (t *Tracer) Repeat(count *Outcome, exp *Outcome) {
}

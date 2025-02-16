package bitly

// Temperature units
type TemperatureUnit int

const (
	Degrees TemperatureUnit = iota + 1
	Kelvin
	Farenheit
)

func (u TemperatureUnit) Type() string {
	return "Temperature"
}

func (l TemperatureUnit) Label() string {
	return [...]string{
		"Degrees",
		"Kelvin",
		"Farenheit",
	}[l-1]
}

func (t TemperatureUnit) Index() int {
	return int(t)
}

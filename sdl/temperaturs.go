package sdl

// Temperature units
type TemperatureUnit int

const (
	Degrees TemperatureUnit = iota + 1
	Kelvin
	Farenheit
	MaxTemperatureUnit
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

func (u TemperatureUnit) Convert(value Fraction, dest Unit) Fraction {
	switch u {
	case Degrees:
		switch dest {
		case Kelvin:
			return value.Plus(Frac(27315, 100)).Factorized()
		case Farenheit:
			return value.Times(Frac(9, 5)).PlusNum(32).Factorized()
		}
	case Kelvin:
		switch dest {
		case Degrees:
			return value.Minus(Frac(27315, 100)).Factorized()
		case Farenheit:
			return value.Minus(Frac(27315, 100)).Times(Frac(9, 5)).PlusNum(32).Factorized()
		}
	case Farenheit:
		switch dest {
		case Degrees:
			return value.Minus(FracN(32)).Times(Frac(5, 9))
		case Kelvin:
			return value.Minus(FracN(32)).Times(Frac(5, 9)).Plus(Frac(27315, 100)).Factorized()
		}
	}
	return value
}

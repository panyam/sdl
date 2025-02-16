package sdl

import "log"

// Length units
type LengthUnit int

type LengthValue Value[LengthUnit]

const (
	MilliMeters LengthUnit = iota + 1
	CentiMeters
	Meters
	KiloMeters
	Inches
	Feet
	Yards
	Miles
	MaxLengthUnit
)

var lengthUnitConversionTable = ConversionTable{int(MaxLengthUnit), nil}

func init() {
	lengthUnitConversionTable.Init()
	lengthUnitConversionTable.Set(CentiMeters.Index(), MilliMeters.Index(), Frac(10, 1))
	lengthUnitConversionTable.Set(Meters.Index(), CentiMeters.Index(), Frac(100, 1))
	lengthUnitConversionTable.Set(KiloMeters.Index(), Meters.Index(), Frac(1000, 1))
	lengthUnitConversionTable.Set(Feet.Index(), Inches.Index(), Frac(12, 1))
	// TODO - some others to complete the graph
}

func (l LengthUnit) Label() string {
	return [...]string{
		"MilliMeters",
		"CentiMeters",
		"Meters",
		"KiloMeters",
		"Inches",
		"Feet",
		"Yards",
		"Miles",
	}[l-1]
}

func (l LengthUnit) Type() string {
	return "Length"
}

func (l LengthUnit) Index() int {
	return int(l)
}

func (u LengthUnit) Convert(value Fraction, dest Unit) Fraction {
	factor := timeUnitConversionTable.Get(int(u), dest.Index())
	if factor.IsZero() {
		log.Fatalf("No conversion found between %s and %s", u.Label(), dest.Label())
	}
	return value.Times(factor)
}

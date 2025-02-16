package sdl

import "log"

// Time units
type TimeUnit int

type TimeValue = Value[TimeUnit]

const (
	NanoSeconds TimeUnit = iota + 1
	MicroSeconds
	MilliSeconds
	Seconds
	Minutes
	Hours
	Days
	Weeks
	MaxTimeUnit
)

var timeUnitConversionTable = ConversionTable{int(MaxTimeUnit), nil}

func init() {
	timeUnitConversionTable.Init()
	timeUnitConversionTable.Set(MicroSeconds.Index(), NanoSeconds.Index(), Frac(1000, 1))
	timeUnitConversionTable.Set(MilliSeconds.Index(), MicroSeconds.Index(), Frac(1000, 1))
	timeUnitConversionTable.Set(Seconds.Index(), MilliSeconds.Index(), Frac(1000, 1))
	timeUnitConversionTable.Set(Minutes.Index(), Seconds.Index(), Frac(60, 1))
	timeUnitConversionTable.Set(Hours.Index(), Minutes.Index(), Frac(60, 1))
	timeUnitConversionTable.Set(Days.Index(), Hours.Index(), Frac(24, 1))
	timeUnitConversionTable.Set(Weeks.Index(), Days.Index(), Frac(7, 1))
}

func (u TimeUnit) Convert(value Fraction, dest Unit) Fraction {
	factor := timeUnitConversionTable.Get(int(u), dest.Index())
	if factor.IsZero() {
		log.Fatalf("No conversion found between %s and %s", u.Label(), dest.Label())
	}
	return value.Times(factor)
}

func (u TimeUnit) ConversionFactor(src, dest TimeUnit) Fraction {
	return Fraction{}
}

func (u TimeUnit) Type() string {
	return "Time"
}

func (l TimeUnit) Label() string {
	return [...]string{
		"NanoSeconds",
		"MicroSeconds",
		"MilliSeconds",
		"Seconds",
		"Minutes",
		"Hours",
		"Days",
		"Weeks",
	}[l-1]
}

func (t TimeUnit) Index() int {
	return int(t)
}

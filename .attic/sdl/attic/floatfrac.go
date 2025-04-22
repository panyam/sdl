package sdl

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

type Fraction float64

var FracZero = Fraction(0)
var FracOne = Fraction(1)
var InvalidFractionString = errors.New("invalid fraction string")

func (f Fraction) Factorized() Fraction {
	return f
	// gcd := GCD(f.Num, f.Den)
	// return Fraction{f.Num / gcd, f.Den / gcd}
}

func FracN(n int64) Fraction {
	return Fraction(n)
}

func Frac(n, d int64) Fraction {
	return Fraction(float64(n) / float64(d))
}

// Converts a string to a fraction.
// The input can be in the format "<integer>" or "<integer>/<integer>"
func StrToFrac(val string) (out Fraction, err error) {
	parts := strings.Split(val, "/")
	var num, den int64
	if len(parts) == 1 {
		num, err = strconv.ParseInt(strings.TrimSpace(parts[0]), 10, 64)
	} else if len(parts) != 2 {
		err = InvalidFractionString
	} else {
		num, err = strconv.ParseInt(strings.TrimSpace(parts[0]), 10, 64)
		if err == nil {
			den, err = strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64)
		}
	}
	if err == nil {
		out = Frac(num, den)
	}
	return
}

func (f Fraction) IsZero() bool {
	return f == 0
}

func (f Fraction) IsOne() bool {
	return f == 1
}

func (f Fraction) Plus(another Fraction) Fraction {
	return f + another
}

func (f Fraction) PlusNum(another int64) Fraction {
	return Fraction(float64(f) + float64(another))
}

func (f Fraction) Minus(another Fraction) Fraction {
	return f - another
}

func (f Fraction) MinusNum(another int64) Fraction {
	return Fraction(float64(f) - float64(another))
}

func (f Fraction) Times(another Fraction) Fraction {
	return f * another
}

func (f Fraction) TimesNum(another int64) Fraction {
	return Fraction(float64(f) * float64(another))
}

func (f Fraction) DivBy(another Fraction) Fraction {
	return f / another
}

func (f Fraction) DivByNum(another int64) Fraction {
	return Fraction(float64(f) / float64(another))
}

/**
 * Returns another / f.
 */
func (f Fraction) NumDivBy(another int64) Fraction {
	return Fraction(float64(another) / float64(f))
}

func (f Fraction) Inverse() Fraction {
	return 1 / f
}

func (f Fraction) Equals(another Fraction) bool {
	return f == another
}

func (f Fraction) equalsNum(another int64) bool {
	return f == Fraction(another)
}

func (f Fraction) IsLT(another Fraction) bool {
	return f < another
}

func (f Fraction) IsLTE(another Fraction) bool {
	return f <= another
}

func (f Fraction) IsLTNum(another int64) bool {
	return f < Fraction(another)
}

func (f Fraction) IsLTENum(another int64) bool {
	return f <= Fraction(another)
}

func (f Fraction) IsGT(another Fraction) bool {
	return f > another
}

func (f Fraction) IsGTE(another Fraction) bool {
	return f >= another
}

func (f Fraction) IsGTNum(another int64) bool {
	return f > Fraction(another)
}

func (f Fraction) IsGTENum(another int64) bool {
	return f >= Fraction(another)
}

func (f Fraction) Abs() Fraction {
	return Fraction(math.Abs(float64(f)))
}

func (f Fraction) String() string {
	return fmt.Sprintf("%f", f)
}

func FracMax(f1 Fraction, f2 Fraction) Fraction {
	if f1 > f2 {
		return f1
	} else {
		return f2
	}
}

func FracMin(f1 Fraction, f2 Fraction) Fraction {
	if f1 < f2 {
		return f1
	} else {
		return f2
	}
}

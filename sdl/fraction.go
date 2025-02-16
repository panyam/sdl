package bitly

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Fraction struct {
	Num int64
	Den int64
}

var ZERO = Fraction{}
var ONE = Fraction{1, 1}
var InvalidFractionString = errors.New("invalid fraction string")

func GCD(x int64, y int64) int64 {
	if x < 0 {
		x = -1
	}
	if y < 0 {
		y = -1
	}
	for y > 0 {
		t := y
		y = x % y
		x = t
	}
	return x
}

func (f Fraction) Factorized() Fraction {
	gcd := GCD(f.Num, f.Den)
	return Fraction{f.Num / gcd, f.Den / gcd}
}

func Frac(n, d int64) Fraction {
	return Fraction{n, d}
}

// Converts a string to a fraction.
// The input can be in the format "<integer>" or "<integer>/<integer>"
func StrToFrac(val string) (out Fraction, err error) {
	parts := strings.Split(val, "/")
	if len(parts) == 1 {
		out.Num, err = strconv.ParseInt(strings.TrimSpace(parts[0]), 10, 64)
	} else if len(parts) != 2 {
		err = InvalidFractionString
	} else {
		out.Num, err = strconv.ParseInt(strings.TrimSpace(parts[0]), 10, 64)
		if err == nil {
			out.Den, err = strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64)
		}
	}
	return
}

func (f Fraction) IsWhole() bool {
	return f.Num%f.Den == 0
}

func (f Fraction) IsZero() bool {
	return f.Num == 0
}

func (f Fraction) IsOne() bool {
	return f.Num == f.Den
}

func (f Fraction) Ceil() int64 {
	if f.Num%f.Den == 0 {
		return f.Num / f.Den
	} else {
		return 1 + f.Num/f.Den
	}
}

func (f Fraction) Floor() int64 {
	if f.Num%f.Den == 0 {
		return f.Num / f.Den
	} else {
		return f.Num / f.Den
	}
}

func (f Fraction) Plus(another Fraction) Fraction {
	return Frac(f.Num*another.Den+f.Den*another.Num, f.Den*another.Den)
}

func (f Fraction) PlusNum(another int64) Fraction {
	return Frac(f.Num+f.Den*another, f.Den)
}

func (f Fraction) Minus(another Fraction) Fraction {
	return Frac(f.Num*another.Den-f.Den*another.Num, f.Den*another.Den)
}

func (f Fraction) MinusNum(another int64) Fraction {
	return Frac(f.Num-f.Den*another, f.Den)
}

func (f Fraction) Times(another Fraction) Fraction {
	return Frac(f.Num*another.Num, f.Den*another.Den)
}

func (f Fraction) TimesNum(another int64) Fraction {
	return Frac(f.Num*another, f.Den)
}

func (f Fraction) DivBy(another Fraction) Fraction {
	return Frac(f.Num*another.Den, f.Den*another.Num)
}

func (f Fraction) DivByNum(another int64) Fraction {
	return Frac(f.Num, f.Den*another)
}

/**
 * Returns another / f.
 */
func (f Fraction) NumDivBy(another int64) Fraction {
	return Frac(f.Den*another, f.Num)
}

/**
 * Returns this % another
 */
func (f Fraction) Mod(another Fraction) Fraction {
	// a (mod b) = a − b ⌊a / b⌋
	d := f.DivBy(another)
	floorOfD := int64(d.Num / d.Den)
	return f.Minus(another.TimesNum(floorOfD))
}

/*
 * Returns this % another
 */
func (f Fraction) ModNum(another int64) Fraction {
	// a (mod b) = a − b ⌊a / b⌋
	d := f.DivByNum(another)
	floorOfD := int64(d.Num / d.Den)
	return f.MinusNum(another * floorOfD)
}

func (f Fraction) Inverse() Fraction {
	return Frac(f.Den, f.Num)
}

func (f Fraction) Equals(another Fraction) bool {
	return f.Num*another.Den == f.Den*another.Num
}

func (f Fraction) equalsNum(another int64) bool {
	return f.Num == f.Den*another
}

func (f Fraction) Compare(another Fraction) int64 {
	return f.Num*another.Den - f.Den*another.Num
}

func (f Fraction) CompareNum(another int64) int64 {
	return f.Num - f.Den*another
}

func (f Fraction) IsLT(another Fraction) bool {
	return f.Compare(another) < 0
}

func (f Fraction) IsLTE(another Fraction) bool {
	return f.Compare(another) <= 0
}

func (f Fraction) IsLTNum(another int64) bool {
	return f.CompareNum(another) < 0
}

func (f Fraction) IsLTENum(another int64) bool {
	return f.CompareNum(another) <= 0
}

func (f Fraction) IsGT(another Fraction) bool {
	return f.Compare(another) > 0
}

func (f Fraction) IsGTE(another Fraction) bool {
	return f.Compare(another) >= 0
}

func (f Fraction) IsGTNum(another int64) bool {
	return f.CompareNum(another) > 0
}

func (f Fraction) IsGTENum(another int64) bool {
	return f.CompareNum(another) >= 0
}

func (f Fraction) String() string {
	return fmt.Sprintf("%d/%d", f.Num, f.Den)
}

func FracMax(f1 Fraction, f2 Fraction) Fraction {
	if f1.Compare(f2) > 0 {
		return f1
	} else {
		return f2
	}
}

func FracMin(f1 Fraction, f2 Fraction) Fraction {
	if f1.Compare(f2) < 0 {
		return f1
	} else {
		return f2
	}
}

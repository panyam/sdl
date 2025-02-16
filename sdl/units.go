package sdl

import (
	"fmt"
)

type Unit interface {
	Type() string
	Label() string
	Index() int
	Convert(value Fraction, dest Unit) Fraction
}

type UnitConverter[U Unit] interface {
	ConversionFactor(source Unit, dest Unit) Fraction
	Convert(source Value[Unit], dest Value[Unit]) Value[Unit]
}

type Value[U Unit] struct {
	Fraction
	Unit U
}

func (v *Value[U]) Key() string {
	return fmt.Sprintf("%s%s", v.Fraction, v.Unit.Label())
}

func Val[U Unit](v int, u U) Value[U] {
	return Value[U]{Fraction: Fraction{int64(v), 1}, Unit: u}
}

func Valf[U Unit](v Fraction, u U) Value[U] {
	return Value[U]{Fraction: v, Unit: u}
}

func (v Value[U]) Convert(toUnit U) (out Value[U]) {
	out.Unit = toUnit
	out.Fraction = v.Unit.Convert(v.Fraction, toUnit)
	return
}

func (v *Value[U]) Add(another Value[U]) (out Value[U]) {
	// convert another to our units
	out.Unit = v.Unit
	// Convert another value to our unit
	f2 := another.Unit.Convert(another.Fraction, v.Unit)
	out.Fraction = v.Fraction.Plus(f2).Factorized()
	return
}

func (v *Value[U]) TimesN(another int64) (out Value[U]) {
	// convert another to our units
	// Convert another value to our unit
	out.Unit = v.Unit
	out.Fraction = v.Fraction.TimesNum(another).Factorized()
	return
}

func (v *Value[U]) Times(another Fraction) (out Value[U]) {
	// convert another to our units
	// Convert another value to our unit
	out.Unit = v.Unit
	out.Fraction = v.Fraction.Times(another).Factorized()
	return
}

type ConversionTable struct {
	N     int
	Table [][]Fraction
}

func (c *ConversionTable) Init() {
	if c.Table == nil {
		for i := range c.N + 1 {
			c.Table = append(c.Table, []Fraction{})
			for range c.N + 1 {
				c.Table[i] = append(c.Table[i], Fraction{})
			}
			c.Table[i][i] = Fraction{1, 1}
		}
	}
}

func (c *ConversionTable) Set(i, j int, factor Fraction) {
	c.Table[i][j] = factor
	c.Table[j][i] = factor.Inverse()
}

func (c *ConversionTable) Get(i, j int) (factor Fraction) {
	factor = c.Table[i][j]
	if factor.IsZero() {
		// TODO: evaluate it
	}
	return
}

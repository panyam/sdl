package decl

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValueTypeString(t *testing.T) {
	assert.Equal(t, "Nil", NilType.String())
	assert.Equal(t, "Bool", BoolType.String())
	assert.Equal(t, "Int", IntType.String())
	assert.Equal(t, "Float", FloatType.String())
	assert.Equal(t, "String", StrType.String())
	assert.Equal(t, "List[Int]", ListType(IntType).String())
	assert.Equal(t, "Outcomes[String]", OutcomesType(StrType).String())
	assert.Equal(t, "List[List[Bool]]", ListType(ListType(BoolType)).String())
	assert.Equal(t, "Outcomes[List[Int]]", OutcomesType(ListType(IntType)).String())
}

func TestValueTypeEquals(t *testing.T) {
	// Basic types
	assert.True(t, NilType.Equals(NilType))
	assert.True(t, IntType.Equals(IntType))
	assert.False(t, IntType.Equals(StrType))
	assert.False(t, IntType.Equals(nil))
	assert.False(t, NilType.Equals(IntType))

	// List types
	listInt1 := ListType(IntType)
	listInt2 := ListType(IntType)
	listStr := ListType(StrType)
	listListInt := ListType(ListType(IntType))
	listListInt2 := ListType(ListType(IntType))
	listListStr := ListType(ListType(StrType))

	assert.True(t, listInt1.Equals(listInt2))
	assert.False(t, listInt1.Equals(listStr))
	assert.False(t, listInt1.Equals(IntType))
	assert.False(t, listInt1.Equals(listListInt))
	assert.True(t, listListInt.Equals(listListInt2))
	assert.False(t, listListInt.Equals(listListStr))

	// Outcomes types
	outInt1 := OutcomesType(IntType)
	outInt2 := OutcomesType(IntType)
	outStr := OutcomesType(StrType)
	assert.True(t, outInt1.Equals(outInt2))
	assert.False(t, outInt1.Equals(outStr))
	assert.False(t, outInt1.Equals(listInt1)) // List != Outcomes
}

func TestNewValue(t *testing.T) {
	// No initial value
	rv, err := NewValue(IntType)
	require.NoError(t, err)
	assert.True(t, rv.Type.Equals(IntType))
	assert.Nil(t, rv.Value)

	// Correct initial value
	rvBool, err := NewValue(BoolType, true)
	require.NoError(t, err)
	assert.True(t, rvBool.Type.Equals(BoolType))
	assert.Equal(t, true, rvBool.Value)

	// Incorrect initial value
	_, err = NewValue(IntType, "hello")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "type mismatch: expected int, got string")

	// Nil initial value
	rvNil, err := NewValue(NilType, nil)
	require.NoError(t, err)
	assert.True(t, rvNil.Type.Equals(NilType))
	assert.Nil(t, rvNil.Value)

	// Nil initial value for non-nil type
	_, err = NewValue(IntType, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "type mismatch: cannot set nil value")
}

func TestValueSet(t *testing.T) {
	// --- Basic Types ---
	rvInt, _ := NewValue(IntType)
	err := rvInt.Set(123) // Correct
	assert.NoError(t, err)
	assert.Equal(t, int64(123), rvInt.Value)
	err = rvInt.Set("abc") // Incorrect
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expected int, got string")
	assert.Equal(t, 123, rvInt.Value) // Value should not change on error
	err = rvInt.Set(nil)              // Incorrect
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot set nil value")

	rvFloat, _ := NewValue(FloatType)
	err = rvFloat.Set(123.45) // Correct
	assert.NoError(t, err)
	assert.Equal(t, 123.45, rvFloat.Value)
	err = rvFloat.Set(123) // Incorrect (int != float64)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expected float64, got int")

	rvNil, _ := NewValue(NilType)
	err = rvNil.Set(nil) // Correct
	assert.NoError(t, err)
	assert.Nil(t, rvNil.Value)
	err = rvNil.Set(123) // Incorrect
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expected nil, got int")

	// --- List Type ---
	intListType := ListType(IntType)
	rvList, _ := NewValue(intListType)

	// Correct list
	rvElem1, _ := NewValue(IntType, 10)
	rvElem2, _ := NewValue(IntType, 20)
	correctList := []*Value{rvElem1, rvElem2}
	err = rvList.Set(correctList)
	assert.NoError(t, err)
	assert.Equal(t, correctList, rvList.Value)

	// Empty list
	emptyList := []*Value{}
	err = rvList.Set(emptyList)
	assert.NoError(t, err)
	assert.Equal(t, emptyList, rvList.Value)

	// Nil slice (Set should accept this for list type?) - Current Set returns error for nil
	// err = rvList.Set(nil)
	// assert.NoError(t, err) // Or Error if Set rejects nil slice for list type
	// assert.Nil(t, rvList.Value)

	// Incorrect list element type
	rvStrElem, _ := NewValue(StrType, "hello")
	incorrectList := []*Value{rvElem1, rvStrElem}
	err = rvList.Set(incorrectList)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "type error in list/outcomes element 1: expected int, got string")
	assert.Equal(t, emptyList, rvList.Value) // Should retain previous value (emptyList)

	// Incorrect slice type
	wrongSliceType := []int{1, 2}
	err = rvList.Set(wrongSliceType)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expected List ([]*Value), got []int")
}

func TestValueString(t *testing.T) {
	rvInt, _ := NewValue(IntType, 10)
	assert.Equal(t, "RV(int: 10)", rvInt.String())

	rvNil, _ := NewValue(NilType, nil)
	assert.Equal(t, "RV(nil: <nil>)", rvNil.String())

	rvList, _ := NewValue(ListType(StrType))
	rvS1, _ := NewValue(StrType, "a")
	rvS2, _ := NewValue(StrType, "b")
	_ = rvList.Set([]*Value{rvS1, rvS2})
	assert.Equal(t, "RV(List[string]: [RV(string: a), RV(string: b)])", rvList.String())

	rvEmptyList, _ := NewValue(ListType(IntType))
	_ = rvEmptyList.Set([]*Value{})
	assert.Equal(t, "RV(List[int]: [])", rvEmptyList.String())

	rvUnitialized, _ := NewValue(BoolType)
	assert.Equal(t, "RV(bool: <nil>)", rvUnitialized.String()) // Shows internal Go nil
}

func TestValueGetters(t *testing.T) {
	// --- Setup some values ---
	rvInt, _ := NewValue(IntType, 123)
	rvBool, _ := NewValue(BoolType, true)
	rvFloat, _ := NewValue(FloatType, 98.6)
	rvStr, _ := NewValue(StrType, "hello")
	rvNil, _ := NewValue(NilType, nil)

	rvListInt, _ := NewValue(ListType(IntType))
	elem1, _ := NewValue(IntType, 1)
	listIntVal := []*Value{elem1}
	_ = rvListInt.Set(listIntVal)

	rvListStr, _ := NewValue(ListType(StrType))
	elemS, _ := NewValue(StrType, "a")
	listStrVal := []*Value{elemS}
	_ = rvListStr.Set(listStrVal)

	rvOutcomes, _ := NewValue(OutcomesType(BoolType))
	elemB, _ := NewValue(BoolType, false)
	outcomesVal := []*Value{elemB}
	_ = rvOutcomes.Set(outcomesVal)

	rvEmptyList, _ := NewValue(ListType(FloatType))
	_ = rvEmptyList.Set([]*Value{})

	rvUninitList, _ := NewValue(ListType(NilType)) // Uninitialized list

	// --- Test GetInt ---
	valI, errI := rvInt.GetInt()
	assert.NoError(t, errI)
	assert.Equal(t, 123, valI)
	_, errI = rvStr.GetInt() // Wrong type
	assert.Error(t, errI)
	assert.Contains(t, errI.Error(), "cannot get Int, value is type string")

	// --- Test GetBool ---
	valB, errB := rvBool.GetBool()
	assert.NoError(t, errB)
	assert.Equal(t, true, valB)
	_, errB = rvInt.GetBool() // Wrong type
	assert.Error(t, errB)
	assert.Contains(t, errB.Error(), "cannot get Bool, value is type int")

	// --- Test GetFloat ---
	valF, errF := rvFloat.GetFloat()
	assert.NoError(t, errF)
	assert.Equal(t, 98.6, valF)
	_, errF = rvBool.GetFloat() // Wrong type
	assert.Error(t, errF)
	assert.Contains(t, errF.Error(), "cannot get Float, value is type bool")

	// --- Test GetString ---
	valS, errS := rvStr.GetString()
	assert.NoError(t, errS)
	assert.Equal(t, "hello", valS)
	_, errS = rvNil.GetString() // Wrong type
	assert.Error(t, errS)
	assert.Contains(t, errS.Error(), "cannot get String, value is type nil")

	// --- Test GetList ---
	valLi, errLi := rvListInt.GetList()
	assert.NoError(t, errLi)
	assert.Equal(t, listIntVal, valLi)
	_, errLi = rvStr.GetList() // Wrong type
	assert.Error(t, errLi)
	assert.Contains(t, errLi.Error(), "cannot get List, value is type string")
	valEmpty, errEmpty := rvEmptyList.GetList() // Empty list
	assert.NoError(t, errEmpty)
	assert.Empty(t, valEmpty)
	valUninit, errUninit := rvUninitList.GetList() // Uninitialized list (Go nil value)
	assert.NoError(t, errUninit)
	assert.Nil(t, valUninit)

	// --- Test GetOutcomes ---
	valO, errO := rvOutcomes.GetOutcomes()
	assert.NoError(t, errO)
	assert.Equal(t, outcomesVal, valO)
	_, errO = rvListStr.GetOutcomes() // Wrong type (List != Outcomes)
	assert.Error(t, errO)
	assert.Contains(t, errO.Error(), "cannot get Outcomes, value is type List[string]")

	// --- Test GetNil ---
	errN := rvNil.GetNil() // Correct type
	assert.NoError(t, errN)
	errN = rvInt.GetNil() // Wrong type
	assert.Error(t, errN)
	assert.Contains(t, errN.Error(), "cannot get Nil, value is type int")
}

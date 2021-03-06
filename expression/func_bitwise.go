//  Copyright (c) 2014 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
//  except in compliance with the License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing, software distributed under the
//  License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
//  either express or implied. See the License for the specific language governing permissions
//  and limitations under the License.

package expression

import (
	"math"

	"github.com/couchbase/query/value"
)

const (
	_BITONE   = 0x01
	_BITSTART = 1
	_BITEND   = 64
)

///////////////////////////////////////////////////
//
// BITAND
//
///////////////////////////////////////////////////

/*
This represents the bit manipulation function BITAND(num1,num2...).
It returns result of the bitwise AND on all input arguments.
*/

type BitAnd struct {
	FunctionBase
}

func NewBitAnd(operands ...Expression) Function {
	rv := &BitAnd{
		*NewFunctionBase("bitand", operands...),
	}

	rv.expr = rv
	return rv
}

/*
Visitor pattern.
*/
func (this *BitAnd) Accept(visitor Visitor) (interface{}, error) {
	return visitor.VisitFunction(this)
}

func (this *BitAnd) Type() value.Type { return value.NUMBER }

func (this *BitAnd) Evaluate(item value.Value, context Context) (value.Value, error) {
	return this.Eval(this, item, context)
}

func (this *BitAnd) Apply(context Context, args ...value.Value) (value.Value, error) {

	for _, arg := range args {
		if arg.Type() == value.MISSING {
			return value.MISSING_VALUE, nil
		}

		// If not a numeric value return NULL.
		if arg.Type() != value.NUMBER {
			return value.NULL_VALUE, nil
		}
	}

	var result, val int64
	var ok bool

	if result, ok = isInt(args[0]); !ok {
		return value.NULL_VALUE, nil
	}

	for _, arg := range args[1:] {
		if val, ok = isInt(arg); !ok {
			return value.NULL_VALUE, nil
		}
		result = result & val
	}
	return value.NewValue(result), nil
}

/*
Minimum input arguments required is 2.
*/
func (this *BitAnd) MinArgs() int { return 2 }

/*
Maximum input arguments allowed.
*/
func (this *BitAnd) MaxArgs() int { return math.MaxInt16 }

/*
Factory method pattern.
*/
func (this *BitAnd) Constructor() FunctionConstructor {
	return NewBitAnd
}

///////////////////////////////////////////////////
//
// BITOR
//
///////////////////////////////////////////////////

/*
This represents the bit manipulation function BITOR(num1,num2...).
It returns result of the bitwise OR on all input arguments.
*/

type BitOr struct {
	FunctionBase
}

func NewBitOr(operands ...Expression) Function {
	rv := &BitOr{
		*NewFunctionBase("bitor", operands...),
	}

	rv.expr = rv
	return rv
}

/*
Visitor pattern.
*/
func (this *BitOr) Accept(visitor Visitor) (interface{}, error) {
	return visitor.VisitFunction(this)
}

func (this *BitOr) Type() value.Type { return value.NUMBER }

func (this *BitOr) Evaluate(item value.Value, context Context) (value.Value, error) {
	return this.Eval(this, item, context)
}

func (this *BitOr) Apply(context Context, args ...value.Value) (value.Value, error) {

	for _, arg := range args {
		if arg.Type() == value.MISSING {
			return value.MISSING_VALUE, nil
		}

		// If not a numeric value return NULL.
		if arg.Type() != value.NUMBER {
			return value.NULL_VALUE, nil
		}
	}

	var result, val int64
	var ok bool

	if result, ok = isInt(args[0]); !ok {
		return value.NULL_VALUE, nil
	}

	for _, arg := range args[1:] {
		if val, ok = isInt(arg); !ok {
			return value.NULL_VALUE, nil
		}
		result = result | val
	}
	return value.NewValue(result), nil
}

/*
Minimum input arguments required is 2.
*/
func (this *BitOr) MinArgs() int { return 2 }

/*
Maximum input arguments allowed.
*/
func (this *BitOr) MaxArgs() int { return math.MaxInt16 }

/*
Factory method pattern.
*/
func (this *BitOr) Constructor() FunctionConstructor {
	return NewBitOr
}

///////////////////////////////////////////////////
//
// BITXOR
//
///////////////////////////////////////////////////

/*
This represents the bit manipulation function BITXOR(num1,num2...).
It returns result of the bitwise XOR on all input arguments.
*/

type BitXor struct {
	FunctionBase
}

func NewBitXor(operands ...Expression) Function {
	rv := &BitXor{
		*NewFunctionBase("bitxor", operands...),
	}

	rv.expr = rv
	return rv
}

/*
Visitor pattern.
*/
func (this *BitXor) Accept(visitor Visitor) (interface{}, error) {
	return visitor.VisitFunction(this)
}

func (this *BitXor) Type() value.Type { return value.NUMBER }

func (this *BitXor) Evaluate(item value.Value, context Context) (value.Value, error) {
	return this.Eval(this, item, context)
}

func (this *BitXor) Apply(context Context, args ...value.Value) (value.Value, error) {

	for _, arg := range args {
		if arg.Type() == value.MISSING {
			return value.MISSING_VALUE, nil
		}

		// If not a numeric value return NULL.
		if arg.Type() != value.NUMBER {
			return value.NULL_VALUE, nil
		}
	}

	var result, val int64
	var ok bool

	if result, ok = isInt(args[0]); !ok {
		return value.NULL_VALUE, nil
	}

	for _, arg := range args[1:] {
		if val, ok = isInt(arg); !ok {
			return value.NULL_VALUE, nil
		}
		result = result ^ val
	}
	return value.NewValue(result), nil

}

/*
Minimum input arguments required is 2.
*/
func (this *BitXor) MinArgs() int { return 2 }

/*
Maximum input arguments allowed.
*/
func (this *BitXor) MaxArgs() int { return math.MaxInt16 }

/*
Factory method pattern.
*/
func (this *BitXor) Constructor() FunctionConstructor {
	return NewBitXor
}

///////////////////////////////////////////////////
//
// BITNOT
//
///////////////////////////////////////////////////

/*
This represents the bit manipulation function BITNOT(num1).
It returns result of the bitwise NOT on all input arguments.
*/

type BitNot struct {
	UnaryFunctionBase
}

func NewBitNot(operand Expression) Function {
	rv := &BitNot{
		*NewUnaryFunctionBase("bitnot", operand),
	}

	rv.expr = rv
	return rv
}

/*
Visitor pattern.
*/
func (this *BitNot) Accept(visitor Visitor) (interface{}, error) {
	return visitor.VisitFunction(this)
}

func (this *BitNot) Type() value.Type { return value.NUMBER }

func (this *BitNot) Evaluate(item value.Value, context Context) (value.Value, error) {
	return this.UnaryEval(this, item, context)
}

/*
This method reverses the input array value and returns it.
If the input value is of type missing return a missing
value, and for all non array values return null.
*/
func (this *BitNot) Apply(context Context, arg value.Value) (value.Value, error) {

	if arg.Type() == value.MISSING {
		return value.MISSING_VALUE, nil
	}

	// If not a numeric value return NULL.
	if arg.Type() != value.NUMBER {
		return value.NULL_VALUE, nil
	}

	var result int64
	var ok bool

	if result, ok = isInt(arg); !ok {
		return value.NULL_VALUE, nil
	}

	result = ^result
	return value.NewValue(result), nil
}

/*
Factory method pattern.
*/
func (this *BitNot) Constructor() FunctionConstructor {
	return func(operands ...Expression) Function {
		return NewBitNot(operands[0])
	}
}

///////////////////////////////////////////////////
//
// BITSHIFT
//
///////////////////////////////////////////////////

/*
This represents the bit manipulation function BITSHIFT(num1,shift amt,is_rotate).
It returns result of the bitwise left or right shift on input argument. If is_rotate
is true then it performs a circular shift. Otherwise it performs a logical shift.
*/

type BitShift struct {
	FunctionBase
}

func NewBitShift(operands ...Expression) Function {
	rv := &BitShift{
		*NewFunctionBase("bitshift", operands...),
	}

	rv.expr = rv
	return rv
}

/*
Visitor pattern.
*/
func (this *BitShift) Accept(visitor Visitor) (interface{}, error) {
	return visitor.VisitFunction(this)
}

func (this *BitShift) Type() value.Type { return value.NUMBER }

func (this *BitShift) Evaluate(item value.Value, context Context) (value.Value, error) {
	return this.Eval(this, item, context)
}

func (this *BitShift) Apply(context Context, args ...value.Value) (value.Value, error) {
	isRotate := false
	for k, arg := range args {
		if arg.Type() == value.MISSING {
			return value.MISSING_VALUE, nil
		}

		if arg.Type() == value.NULL {
			return value.NULL_VALUE, nil
		}

		if k == 2 {
			if arg.Type() != value.BOOLEAN {
				return value.NULL_VALUE, nil
			}
			isRotate = args[2].Actual().(bool)
		} else {
			// If not a numeric value return NULL.
			if arg.Type() != value.NUMBER {
				return value.NULL_VALUE, nil
			}
		}

	}

	var num1, shift int64
	var ok bool

	if num1, ok = isInt(args[0]); !ok {
		return value.NULL_VALUE, nil
	}

	if shift, ok = isInt(args[1]); !ok {
		return value.NULL_VALUE, nil
	}

	var result uint64
	// Check if it is rotate and shift
	if isRotate == true {
		result = rotateLeft(uint64(num1), int(shift))
	} else {
		result = shiftLeft(uint64(num1), int(shift))
	}

	return value.NewValue(int64(result)), nil

}

/*
Minimum input arguments required is 2.
*/
func (this *BitShift) MinArgs() int { return 2 }

/*
Maximum input arguments allowed.
*/
func (this *BitShift) MaxArgs() int { return 3 }

/*
Factory method pattern.
*/
func (this *BitShift) Constructor() FunctionConstructor {
	return NewBitShift
}

///////////////////////////////////////////////////
//
// BITSET
//
///////////////////////////////////////////////////

/*
This represents the bit manipulation function BITSET(num1,[list of positions]]).
It returns the value after setting the bits at the input positions.
*/

type BitSet struct {
	BinaryFunctionBase
}

func NewBitSet(first, second Expression) Function {
	rv := &BitSet{
		*NewBinaryFunctionBase("bitset", first, second),
	}

	rv.expr = rv
	return rv
}

/*
Visitor pattern.
*/
func (this *BitSet) Accept(visitor Visitor) (interface{}, error) {
	return visitor.VisitFunction(this)
}

func (this *BitSet) Type() value.Type { return value.NUMBER }

func (this *BitSet) Evaluate(item value.Value, context Context) (value.Value, error) {
	return this.BinaryEval(this, item, context)
}

func (this *BitSet) Apply(context Context, first, second value.Value) (value.Value, error) {
	return bitSetNClear(true, context, first, second)
}

func (this *BitSet) Constructor() FunctionConstructor {
	return func(operands ...Expression) Function {
		return NewBitSet(operands[0], operands[1])
	}
}

///////////////////////////////////////////////////
//
// BITCLEAR
//
///////////////////////////////////////////////////

/*
This represents the bit manipulation function BITCLEAR(num1,[list of positions]]).
It returns the value after clearing the bits at the input positions.
*/

type BitClear struct {
	BinaryFunctionBase
}

func NewBitClear(first, second Expression) Function {
	rv := &BitClear{
		*NewBinaryFunctionBase("bitclear", first, second),
	}

	rv.expr = rv
	return rv
}

/*
Visitor pattern.
*/
func (this *BitClear) Accept(visitor Visitor) (interface{}, error) {
	return visitor.VisitFunction(this)
}

func (this *BitClear) Type() value.Type { return value.NUMBER }

func (this *BitClear) Evaluate(item value.Value, context Context) (value.Value, error) {
	return this.BinaryEval(this, item, context)
}

func (this *BitClear) Apply(context Context, first, second value.Value) (value.Value, error) {
	return bitSetNClear(false, context, first, second)
}

func (this *BitClear) Constructor() FunctionConstructor {
	return func(operands ...Expression) Function {
		return NewBitClear(operands[0], operands[1])
	}
}

///////////////////////////////////////////////////
//
// BITTEST OR ISBITSET
//
///////////////////////////////////////////////////

/*
This represents the bit manipulation function BitTest(num1, <list of bit positions>,<all set>).
It returns true if any or all the bits in positions are set.
*/

type BitTest struct {
	FunctionBase
}

func NewBitTest(operands ...Expression) Function {
	rv := &BitTest{
		*NewFunctionBase("bittest", operands...),
	}

	rv.expr = rv
	return rv
}

/*
Visitor pattern.
*/
func (this *BitTest) Accept(visitor Visitor) (interface{}, error) {
	return visitor.VisitFunction(this)
}

func (this *BitTest) Type() value.Type { return value.BOOLEAN }

func (this *BitTest) Evaluate(item value.Value, context Context) (value.Value, error) {
	return this.Eval(this, item, context)
}

func (this *BitTest) Apply(context Context, args ...value.Value) (value.Value, error) {
	isAll := false
	for k, arg := range args {
		if arg.Type() == value.MISSING {
			return value.MISSING_VALUE, nil
		}

		if arg.Type() == value.NULL {
			return value.NULL_VALUE, nil
		}

		// For 2nd arg - num or array ok
		if k == 1 {
			if arg.Type() != value.NUMBER && arg.Type() != value.ARRAY {
				return value.NULL_VALUE, nil
			}
		} else if k == 2 {
			if arg.Type() != value.BOOLEAN {
				return value.NULL_VALUE, nil
			}
			isAll = args[2].Actual().(bool)
		} else {
			if arg.Type() != value.NUMBER {
				return value.NULL_VALUE, nil
			}
		}
	}

	var num1 int64
	var bitP uint64
	var ok bool

	if num1, ok = isInt(args[0]); !ok {
		return value.NULL_VALUE, nil
	}

	bitP, ok = bitPositions(args[1])

	if !ok {
		return value.NULL_VALUE, nil
	} else {
		if isAll {
			return value.NewValue((uint64(num1) & bitP) == bitP), nil
		}
		return value.NewValue((uint64(num1) & bitP) != 0), nil
	}
}

/*
Minimum input arguments required is 2.
*/
func (this *BitTest) MinArgs() int { return 2 }

/*
Maximum input arguments allowed.
*/
func (this *BitTest) MaxArgs() int { return 3 }

/*
Factory method pattern.
*/
func (this *BitTest) Constructor() FunctionConstructor {
	return NewBitTest
}

// Function to set a bit or clear a bit in the input number
func bitSetNClear(bitset bool, context Context, first, second value.Value) (value.Value, error) {
	if first.Type() == value.MISSING || second.Type() == value.MISSING {
		return value.MISSING_VALUE, nil
	} else if first.Type() != value.NUMBER {
		return value.NULL_VALUE, nil
	} else if second.Type() != value.NUMBER && second.Type() != value.ARRAY {
		return value.NULL_VALUE, nil
	}

	var num1 int64
	var bitP, result uint64
	var ok bool

	if num1, ok = isInt(first); !ok {
		return value.NULL_VALUE, nil
	}

	bitP, ok = bitPositions(second)

	if !ok {
		return value.NULL_VALUE, nil
	} else {
		if bitset {
			result = uint64(num1) | bitP
		} else {
			result = uint64(num1) & ^bitP
		}
	}
	return value.NewValue(result), nil
}

func bitPositions(arg value.Value) (uint64, bool) {

	var pp, ppos int64
	var ok bool

	var pos []interface{}
	num1 := uint64(0)

	if arg.Type() == value.NUMBER {
		if pp, ok = isInt(arg); !ok {
			return num1, false
		}
		pos = []interface{}{pp}
	} else {
		pos = arg.Actual().([]interface{})
	}

	// now that the array or positions has been populated.
	// range through

	for _, p := range pos {
		if ppos, ok = isInt(value.NewValue(p)); !ok || ppos < _BITSTART || ppos > _BITEND {
			return num1, false
		}
		num1 = num1 | _BITONE<<uint64(ppos-_BITSTART)
	}
	return num1, true
}

// RotateLeft returns the value of x rotated left by (k mod 64) bits.
// To rotate x right by k bits, call RotateLeft64(x, -k).

// shift count type int64, must be unsigned integer

func rotateLeft(x uint64, k int) uint64 {
	const n = 64
	s := uint(k) & (n - 1)
	return x<<s | x>>(n-s)
}

// ShiftLeft returns the value of x shift left by input bits.
// To shift x right by k bits, call ShiftLeft(x, -k).
func shiftLeft(x uint64, k int) uint64 {
	if k < 0 {
		return x >> uint64(-k)
	}
	return x << uint64(k)
}

func isInt(val value.Value) (int64, bool) {
	actual := val.ActualForIndex()
	switch actual := actual.(type) {
	case float64:
		if value.IsInt(actual) {
			return int64(actual), true
		}
	case int64:
		return int64(actual), true
	}
	return 0, false
}

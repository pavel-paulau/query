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
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/couchbaselabs/query/value"
)

type ClockNowMillis struct {
	ExpressionBase
}

func NewClockNowMillis() Function {
	return &ClockNowMillis{}
}

func (this *ClockNowMillis) Evaluate(item value.Value, context Context) (value.Value, error) {
	nanos := time.Now().UnixNano()
	return value.NewValue(float64(nanos) / (1000000.0)), nil
}

func (this *ClockNowMillis) EquivalentTo(other Expression) bool {
	return this.equivalentTo(this, other)
}

func (this *ClockNowMillis) Fold() (Expression, error) {
	return this, nil
}

func (this *ClockNowMillis) Formalize(forbidden, allowed value.Value, keyspace string) (Expression, error) {
	return this, nil
}

func (this *ClockNowMillis) SubsetOf(other Expression) bool {
	return this.subsetOf(this, other)
}

func (this *ClockNowMillis) VisitChildren(visitor Visitor) (Expression, error) {
	return this, nil
}

func (this *ClockNowMillis) Constructor() FunctionConstructor {
	return func(Expressions) Function { return this }
}

type ClockNowStr struct {
	nAryBase
}

func NewClockNowStr(args Expressions) Function {
	return &ClockNowStr{
		nAryBase{
			operands: args,
		},
	}
}

func (this *ClockNowStr) Evaluate(item value.Value, context Context) (value.Value, error) {
	return this.evaluate(this, item, context)
}

func (this *ClockNowStr) EquivalentTo(other Expression) bool {
	return this.equivalentTo(this, other)
}

func (this *ClockNowStr) Fold() (Expression, error) {
	return this.fold(this)
}

func (this *ClockNowStr) Formalize(forbidden, allowed value.Value, keyspace string) (Expression, error) {
	return this.formalize(this, forbidden, allowed, keyspace)
}

func (this *ClockNowStr) SubsetOf(other Expression) bool {
	return this.subsetOf(this, other)
}

func (this *ClockNowStr) VisitChildren(visitor Visitor) (Expression, error) {
	return this.visitChildren(this, visitor)
}

func (this *ClockNowStr) eval(args value.Values) (value.Value, error) {
	fmt := _DEFAULT_FORMAT
	if len(args) > 0 {
		fv := args[0]
		if fv.Type() == value.MISSING {
			return value.MISSING_VALUE, nil
		} else if fv.Type() != value.STRING {
			return value.NULL_VALUE, nil
		}

		fmt = fv.Actual().(string)
	}

	return value.NewValue(timeToStr(time.Now(), fmt)), nil
}

func (this *ClockNowStr) MinArgs() int { return 0 }

func (this *ClockNowStr) MaxArgs() int { return 1 }

func (this *ClockNowStr) Constructor() FunctionConstructor { return NewClockNowStr }

type DateAddMillis struct {
	nAryBase
}

func NewDateAddMillis(args Expressions) Function {
	return &DateAddMillis{
		nAryBase{
			operands: args,
		},
	}
}

func (this *DateAddMillis) Evaluate(item value.Value, context Context) (value.Value, error) {
	return this.evaluate(this, item, context)
}

func (this *DateAddMillis) EquivalentTo(other Expression) bool {
	return this.equivalentTo(this, other)
}

func (this *DateAddMillis) Fold() (Expression, error) {
	return this.fold(this)
}

func (this *DateAddMillis) Formalize(forbidden, allowed value.Value, keyspace string) (Expression, error) {
	return this.formalize(this, forbidden, allowed, keyspace)
}

func (this *DateAddMillis) SubsetOf(other Expression) bool {
	return this.subsetOf(this, other)
}

func (this *DateAddMillis) VisitChildren(visitor Visitor) (Expression, error) {
	return this.visitChildren(this, visitor)
}

func (this *DateAddMillis) eval(args value.Values) (value.Value, error) {
	ev := args[0]
	nv := args[1]
	pv := args[2]

	if ev.Type() == value.MISSING || nv.Type() == value.MISSING ||
		pv.Type() == value.MISSING {
		return value.MISSING_VALUE, nil
	} else if ev.Type() != value.NUMBER || nv.Type() != value.NUMBER ||
		pv.Type() != value.STRING {
		return value.NULL_VALUE, nil
	}

	ea := ev.Actual().(float64)
	na := nv.Actual().(float64)
	if na != math.Trunc(na) {
		return value.NULL_VALUE, nil
	}

	pa := pv.Actual().(string)
	t, e := dateAdd(millisToTime(ea), int(na), pa)
	if e != nil {
		return value.NULL_VALUE, nil
	}

	return value.NewValue(timeToMillis(t)), nil
}

func (this *DateAddMillis) MinArgs() int { return 3 }

func (this *DateAddMillis) MaxArgs() int { return 3 }

func (this *DateAddMillis) Constructor() FunctionConstructor { return NewDateAddMillis }

type DateAddStr struct {
	nAryBase
}

func NewDateAddStr(args Expressions) Function {
	return &DateAddStr{
		nAryBase{
			operands: args,
		},
	}
}

func (this *DateAddStr) Evaluate(item value.Value, context Context) (value.Value, error) {
	return this.evaluate(this, item, context)
}

func (this *DateAddStr) EquivalentTo(other Expression) bool {
	return this.equivalentTo(this, other)
}

func (this *DateAddStr) Fold() (Expression, error) {
	return this.fold(this)
}

func (this *DateAddStr) Formalize(forbidden, allowed value.Value, keyspace string) (Expression, error) {
	return this.formalize(this, forbidden, allowed, keyspace)
}

func (this *DateAddStr) SubsetOf(other Expression) bool {
	return this.subsetOf(this, other)
}

func (this *DateAddStr) VisitChildren(visitor Visitor) (Expression, error) {
	return this.visitChildren(this, visitor)
}

func (this *DateAddStr) eval(args value.Values) (value.Value, error) {
	ev := args[0]
	nv := args[1]
	pv := args[2]

	if ev.Type() == value.MISSING || nv.Type() == value.MISSING ||
		pv.Type() == value.MISSING {
		return value.MISSING_VALUE, nil
	} else if ev.Type() != value.STRING || nv.Type() != value.NUMBER ||
		pv.Type() != value.STRING {
		return value.NULL_VALUE, nil
	}

	ea := ev.Actual().(string)
	t, e := strToTime(ea)
	if e != nil {
		return value.NULL_VALUE, nil
	}

	na := nv.Actual().(float64)
	if na != math.Trunc(na) {
		return value.NULL_VALUE, nil
	}

	pa := pv.Actual().(string)
	t, e = dateAdd(t, int(na), pa)
	if e != nil {
		return value.NULL_VALUE, nil
	}

	return value.NewValue(timeToStr(t, ea)), nil
}

func (this *DateAddStr) MinArgs() int { return 3 }

func (this *DateAddStr) MaxArgs() int { return 3 }

func (this *DateAddStr) Constructor() FunctionConstructor { return NewDateAddStr }

type DateDiffMillis struct {
	nAryBase
}

func NewDateDiffMillis(args Expressions) Function {
	return &DateDiffMillis{
		nAryBase{
			operands: args,
		},
	}
}

func (this *DateDiffMillis) Evaluate(item value.Value, context Context) (value.Value, error) {
	return this.evaluate(this, item, context)
}

func (this *DateDiffMillis) EquivalentTo(other Expression) bool {
	return this.equivalentTo(this, other)
}

func (this *DateDiffMillis) Fold() (Expression, error) {
	return this.fold(this)
}

func (this *DateDiffMillis) Formalize(forbidden, allowed value.Value, keyspace string) (Expression, error) {
	return this.formalize(this, forbidden, allowed, keyspace)
}

func (this *DateDiffMillis) SubsetOf(other Expression) bool {
	return this.subsetOf(this, other)
}

func (this *DateDiffMillis) VisitChildren(visitor Visitor) (Expression, error) {
	return this.visitChildren(this, visitor)
}

func (this *DateDiffMillis) eval(args value.Values) (value.Value, error) {
	dv1 := args[0]
	dv2 := args[1]
	pv := args[2]

	if dv2.Type() == value.MISSING || dv2.Type() == value.MISSING ||
		pv.Type() == value.MISSING {
		return value.MISSING_VALUE, nil
	} else if dv1.Type() != value.NUMBER || dv2.Type() != value.NUMBER ||
		pv.Type() != value.STRING {
		return value.NULL_VALUE, nil
	}

	da1 := dv1.Actual().(float64)
	da2 := dv2.Actual().(float64)
	pa := pv.Actual().(string)
	diff, e := dateDiff(millisToTime(da1), millisToTime(da2), pa)
	if e != nil {
		return value.NULL_VALUE, nil
	}

	return value.NewValue(float64(diff)), nil
}

func (this *DateDiffMillis) MinArgs() int { return 3 }

func (this *DateDiffMillis) MaxArgs() int { return 3 }

func (this *DateDiffMillis) Constructor() FunctionConstructor { return NewDateDiffMillis }

type DateDiffStr struct {
	nAryBase
}

func NewDateDiffStr(args Expressions) Function {
	return &DateDiffStr{
		nAryBase{
			operands: args,
		},
	}
}

func (this *DateDiffStr) Evaluate(item value.Value, context Context) (value.Value, error) {
	return this.evaluate(this, item, context)
}

func (this *DateDiffStr) EquivalentTo(other Expression) bool {
	return this.equivalentTo(this, other)
}

func (this *DateDiffStr) Fold() (Expression, error) {
	return this.fold(this)
}

func (this *DateDiffStr) Formalize(forbidden, allowed value.Value, keyspace string) (Expression, error) {
	return this.formalize(this, forbidden, allowed, keyspace)
}

func (this *DateDiffStr) SubsetOf(other Expression) bool {
	return this.subsetOf(this, other)
}

func (this *DateDiffStr) VisitChildren(visitor Visitor) (Expression, error) {
	return this.visitChildren(this, visitor)
}

func (this *DateDiffStr) eval(args value.Values) (value.Value, error) {
	dv1 := args[0]
	dv2 := args[1]
	pv := args[2]

	if dv2.Type() == value.MISSING || dv2.Type() == value.MISSING ||
		pv.Type() == value.MISSING {
		return value.MISSING_VALUE, nil
	} else if dv1.Type() != value.STRING || dv2.Type() != value.STRING ||
		pv.Type() != value.STRING {
		return value.NULL_VALUE, nil
	}

	da1 := dv1.Actual().(string)
	t1, e := strToTime(da1)
	if e != nil {
		return value.NULL_VALUE, nil
	}

	da2 := dv2.Actual().(string)
	t2, e := strToTime(da2)
	if e != nil {
		return value.NULL_VALUE, nil
	}

	pa := pv.Actual().(string)
	diff, e := dateDiff(t1, t2, pa)
	if e != nil {
		return value.NULL_VALUE, nil
	}

	return value.NewValue(float64(diff)), nil
}

func (this *DateDiffStr) MinArgs() int { return 3 }

func (this *DateDiffStr) MaxArgs() int { return 3 }

func (this *DateDiffStr) Constructor() FunctionConstructor { return NewDateDiffStr }

type DatePartMillis struct {
	binaryBase
}

func NewDatePartMillis(first, second Expression) Function {
	return &DatePartMillis{
		binaryBase{
			first:  first,
			second: second,
		},
	}
}

func (this *DatePartMillis) Evaluate(item value.Value, context Context) (value.Value, error) {
	return this.evaluate(this, item, context)
}

func (this *DatePartMillis) EquivalentTo(other Expression) bool {
	return this.equivalentTo(this, other)
}

func (this *DatePartMillis) Fold() (Expression, error) {
	return this.fold(this)
}

func (this *DatePartMillis) Formalize(forbidden, allowed value.Value, keyspace string) (Expression, error) {
	return this.formalize(this, forbidden, allowed, keyspace)
}

func (this *DatePartMillis) SubsetOf(other Expression) bool {
	return this.subsetOf(this, other)
}

func (this *DatePartMillis) VisitChildren(visitor Visitor) (Expression, error) {
	return this.visitChildren(this, visitor)
}

func (this *DatePartMillis) eval(first, second value.Value) (value.Value, error) {
	if first.Type() == value.MISSING || second.Type() == value.MISSING {
		return value.MISSING_VALUE, nil
	} else if first.Type() != value.NUMBER || second.Type() != value.STRING {
		return value.NULL_VALUE, nil
	}

	millis := first.Actual().(float64)
	part := second.Actual().(string)
	rv, e := datePart(millisToTime(millis), part)
	if e != nil {
		return value.NULL_VALUE, nil
	}

	return value.NewValue(float64(rv)), nil
}

func (this *DatePartMillis) Constructor() FunctionConstructor {
	return func(args Expressions) Function {
		return NewDatePartMillis(args[0], args[1])
	}
}

type DatePartStr struct {
	binaryBase
}

func NewDatePartStr(first, second Expression) Function {
	return &DatePartStr{
		binaryBase{
			first:  first,
			second: second,
		},
	}
}

func (this *DatePartStr) Evaluate(item value.Value, context Context) (value.Value, error) {
	return this.evaluate(this, item, context)
}

func (this *DatePartStr) EquivalentTo(other Expression) bool {
	return this.equivalentTo(this, other)
}

func (this *DatePartStr) Fold() (Expression, error) {
	return this.fold(this)
}

func (this *DatePartStr) Formalize(forbidden, allowed value.Value, keyspace string) (Expression, error) {
	return this.formalize(this, forbidden, allowed, keyspace)
}

func (this *DatePartStr) SubsetOf(other Expression) bool {
	return this.subsetOf(this, other)
}

func (this *DatePartStr) VisitChildren(visitor Visitor) (Expression, error) {
	return this.visitChildren(this, visitor)
}

func (this *DatePartStr) eval(first, second value.Value) (value.Value, error) {
	if first.Type() == value.MISSING || second.Type() == value.MISSING {
		return value.MISSING_VALUE, nil
	} else if first.Type() != value.STRING || second.Type() != value.STRING {
		return value.NULL_VALUE, nil
	}

	str := first.Actual().(string)
	part := second.Actual().(string)
	t, e := strToTime(str)
	if e != nil {
		return value.NULL_VALUE, nil
	}

	rv, e := datePart(t, part)
	if e != nil {
		return value.NULL_VALUE, nil
	}

	return value.NewValue(float64(rv)), nil
}

func (this *DatePartStr) Constructor() FunctionConstructor {
	return func(args Expressions) Function {
		return NewDatePartStr(args[0], args[1])
	}
}

type DateTruncMillis struct {
	binaryBase
}

func NewDateTruncMillis(first, second Expression) Function {
	return &DateTruncMillis{
		binaryBase{
			first:  first,
			second: second,
		},
	}
}

func (this *DateTruncMillis) Evaluate(item value.Value, context Context) (value.Value, error) {
	return this.evaluate(this, item, context)
}

func (this *DateTruncMillis) EquivalentTo(other Expression) bool {
	return this.equivalentTo(this, other)
}

func (this *DateTruncMillis) Fold() (Expression, error) {
	return this.fold(this)
}

func (this *DateTruncMillis) Formalize(forbidden, allowed value.Value, keyspace string) (Expression, error) {
	return this.formalize(this, forbidden, allowed, keyspace)
}

func (this *DateTruncMillis) SubsetOf(other Expression) bool {
	return this.subsetOf(this, other)
}

func (this *DateTruncMillis) VisitChildren(visitor Visitor) (Expression, error) {
	return this.visitChildren(this, visitor)
}

func (this *DateTruncMillis) eval(first, second value.Value) (value.Value, error) {
	if first.Type() == value.MISSING || second.Type() == value.MISSING {
		return value.MISSING_VALUE, nil
	} else if first.Type() != value.NUMBER || second.Type() != value.STRING {
		return value.NULL_VALUE, nil
	}

	millis := first.Actual().(float64)
	part := second.Actual().(string)
	t := millisToTime(millis)

	var e error
	t, e = dateTrunc(t, part)
	if e != nil {
		return value.NULL_VALUE, nil
	}

	return value.NewValue(timeToMillis(t)), nil
}

func (this *DateTruncMillis) Constructor() FunctionConstructor {
	return func(args Expressions) Function {
		return NewDateTruncMillis(args[0], args[1])
	}
}

type DateTruncStr struct {
	binaryBase
}

func NewDateTruncStr(first, second Expression) Function {
	return &DateTruncStr{
		binaryBase{
			first:  first,
			second: second,
		},
	}
}

func (this *DateTruncStr) Evaluate(item value.Value, context Context) (value.Value, error) {
	return this.evaluate(this, item, context)
}

func (this *DateTruncStr) EquivalentTo(other Expression) bool {
	return this.equivalentTo(this, other)
}

func (this *DateTruncStr) Fold() (Expression, error) {
	return this.fold(this)
}

func (this *DateTruncStr) Formalize(forbidden, allowed value.Value, keyspace string) (Expression, error) {
	return this.formalize(this, forbidden, allowed, keyspace)
}

func (this *DateTruncStr) SubsetOf(other Expression) bool {
	return this.subsetOf(this, other)
}

func (this *DateTruncStr) VisitChildren(visitor Visitor) (Expression, error) {
	return this.visitChildren(this, visitor)
}

func (this *DateTruncStr) eval(first, second value.Value) (value.Value, error) {
	if first.Type() == value.MISSING || second.Type() == value.MISSING {
		return value.MISSING_VALUE, nil
	} else if first.Type() != value.STRING || second.Type() != value.STRING {
		return value.NULL_VALUE, nil
	}

	str := first.Actual().(string)
	part := second.Actual().(string)
	t, e := strToTime(str)
	if e != nil {
		return value.NULL_VALUE, nil
	}

	t, e = dateTrunc(t, part)
	if e != nil {
		return value.NULL_VALUE, nil
	}

	return value.NewValue(timeToStr(t, str)), nil
}

func (this *DateTruncStr) Constructor() FunctionConstructor {
	return func(args Expressions) Function {
		return NewDateTruncStr(args[0], args[1])
	}
}

type MillisToStr struct {
	nAryBase
}

func NewMillisToStr(args Expressions) Function {
	return &MillisToStr{
		nAryBase{
			operands: args,
		},
	}
}

func (this *MillisToStr) Evaluate(item value.Value, context Context) (value.Value, error) {
	return this.evaluate(this, item, context)
}

func (this *MillisToStr) EquivalentTo(other Expression) bool {
	return this.equivalentTo(this, other)
}

func (this *MillisToStr) Fold() (Expression, error) {
	return this.fold(this)
}

func (this *MillisToStr) Formalize(forbidden, allowed value.Value, keyspace string) (Expression, error) {
	return this.formalize(this, forbidden, allowed, keyspace)
}

func (this *MillisToStr) SubsetOf(other Expression) bool {
	return this.subsetOf(this, other)
}

func (this *MillisToStr) VisitChildren(visitor Visitor) (Expression, error) {
	return this.visitChildren(this, visitor)
}

func (this *MillisToStr) eval(args value.Values) (value.Value, error) {
	ev := args[0]
	fv := _DEFAULT_FMT_VALUE
	if len(args) > 1 {
		fv = args[1]
	}

	if ev.Type() == value.MISSING || fv.Type() == value.MISSING {
		return value.MISSING_VALUE, nil
	} else if ev.Type() != value.NUMBER || fv.Type() != value.STRING {
		return value.NULL_VALUE, nil
	}

	millis := ev.Actual().(float64)
	fmt := fv.Actual().(string)
	t := millisToTime(millis)
	return value.NewValue(timeToStr(t, fmt)), nil
}

func (this *MillisToStr) MinArgs() int { return 1 }

func (this *MillisToStr) MaxArgs() int { return 2 }

func (this *MillisToStr) Constructor() FunctionConstructor { return NewMillisToStr }

type MillisToUTC struct {
	nAryBase
}

func NewMillisToUTC(args Expressions) Function {
	return &MillisToUTC{
		nAryBase{
			operands: args,
		},
	}
}

func (this *MillisToUTC) Evaluate(item value.Value, context Context) (value.Value, error) {
	return this.evaluate(this, item, context)
}

func (this *MillisToUTC) EquivalentTo(other Expression) bool {
	return this.equivalentTo(this, other)
}

func (this *MillisToUTC) Fold() (Expression, error) {
	return this.fold(this)
}

func (this *MillisToUTC) Formalize(forbidden, allowed value.Value, keyspace string) (Expression, error) {
	return this.formalize(this, forbidden, allowed, keyspace)
}

func (this *MillisToUTC) SubsetOf(other Expression) bool {
	return this.subsetOf(this, other)
}

func (this *MillisToUTC) VisitChildren(visitor Visitor) (Expression, error) {
	return this.visitChildren(this, visitor)
}

func (this *MillisToUTC) eval(args value.Values) (value.Value, error) {
	ev := args[0]
	fv := _DEFAULT_FMT_VALUE
	if len(args) > 1 {
		fv = args[1]
	}

	if ev.Type() == value.MISSING || fv.Type() == value.MISSING {
		return value.MISSING_VALUE, nil
	} else if ev.Type() != value.NUMBER || fv.Type() != value.STRING {
		return value.NULL_VALUE, nil
	}

	millis := ev.Actual().(float64)
	fmt := fv.Actual().(string)
	t := millisToTime(millis).UTC()
	return value.NewValue(timeToStr(t, fmt)), nil
}

func (this *MillisToUTC) MinArgs() int { return 1 }

func (this *MillisToUTC) MaxArgs() int { return 2 }

func (this *MillisToUTC) Constructor() FunctionConstructor { return NewMillisToUTC }

type MillisToZoneName struct {
	nAryBase
}

func NewMillisToZoneName(args Expressions) Function {
	return &MillisToZoneName{
		nAryBase{
			operands: args,
		},
	}
}

func (this *MillisToZoneName) Evaluate(item value.Value, context Context) (value.Value, error) {
	return this.evaluate(this, item, context)
}

func (this *MillisToZoneName) EquivalentTo(other Expression) bool {
	return this.equivalentTo(this, other)
}

func (this *MillisToZoneName) Fold() (Expression, error) {
	return this.fold(this)
}

func (this *MillisToZoneName) Formalize(forbidden, allowed value.Value, keyspace string) (Expression, error) {
	return this.formalize(this, forbidden, allowed, keyspace)
}

func (this *MillisToZoneName) SubsetOf(other Expression) bool {
	return this.subsetOf(this, other)
}

func (this *MillisToZoneName) VisitChildren(visitor Visitor) (Expression, error) {
	return this.visitChildren(this, visitor)
}

func (this *MillisToZoneName) eval(args value.Values) (value.Value, error) {
	ev := args[0]
	zv := args[1]
	fv := _DEFAULT_FMT_VALUE
	if len(args) > 2 {
		fv = args[2]
	}

	if ev.Type() == value.MISSING || zv.Type() == value.MISSING || fv.Type() == value.MISSING {
		return value.MISSING_VALUE, nil
	} else if ev.Type() != value.NUMBER || zv.Type() != value.STRING || fv.Type() != value.STRING {
		return value.NULL_VALUE, nil
	}

	millis := ev.Actual().(float64)
	tz := zv.Actual().(string)
	loc, e := time.LoadLocation(tz)
	if e != nil {
		return value.NULL_VALUE, nil
	}

	fmt := fv.Actual().(string)
	t := millisToTime(millis).In(loc)
	return value.NewValue(timeToStr(t, fmt)), nil
}

func (this *MillisToZoneName) MinArgs() int { return 2 }

func (this *MillisToZoneName) MaxArgs() int { return 3 }

func (this *MillisToZoneName) Constructor() FunctionConstructor { return NewMillisToZoneName }

type NowMillis struct {
	ExpressionBase
}

func NewNowMillis() Function {
	return &NowMillis{}
}

func (this *NowMillis) Evaluate(item value.Value, context Context) (value.Value, error) {
	nanos := context.(Context).Now().UnixNano()
	return value.NewValue(float64(nanos) / (1000000.0)), nil
}

func (this *NowMillis) EquivalentTo(other Expression) bool {
	return this.equivalentTo(this, other)
}

func (this *NowMillis) Fold() (Expression, error) {
	return this, nil
}

func (this *NowMillis) Formalize(forbidden, allowed value.Value, keyspace string) (Expression, error) {
	return this, nil
}

func (this *NowMillis) SubsetOf(other Expression) bool {
	return this.subsetOf(this, other)
}

func (this *NowMillis) VisitChildren(visitor Visitor) (Expression, error) {
	return this, nil
}

func (this *NowMillis) Constructor() FunctionConstructor {
	return func(Expressions) Function { return this }
}

type NowStr struct {
	nAryBase
}

func NewNowStr(args Expressions) Function {
	return &NowStr{
		nAryBase{
			operands: args,
		},
	}
}

func (this *NowStr) Evaluate(item value.Value, context Context) (value.Value, error) {
	var e error
	args := make([]value.Value, len(this.operands))
	for i, o := range this.operands {
		args[i], e = o.Evaluate(item, context)
		if e != nil {
			return nil, e
		}
	}

	fmt := _DEFAULT_FORMAT
	if len(args) > 0 {
		fv := args[0]
		if fv.Type() == value.MISSING {
			return value.MISSING_VALUE, nil
		} else if fv.Type() != value.STRING {
			return value.NULL_VALUE, nil
		}

		fmt = fv.Actual().(string)
	}

	now := context.(Context).Now()
	return value.NewValue(timeToStr(now, fmt)), nil
}

func (this *NowStr) EquivalentTo(other Expression) bool {
	return this.equivalentTo(this, other)
}

func (this *NowStr) Fold() (Expression, error) {
	return this, nil
}

func (this *NowStr) Formalize(forbidden, allowed value.Value, keyspace string) (Expression, error) {
	return this.formalize(this, forbidden, allowed, keyspace)
}

func (this *NowStr) SubsetOf(other Expression) bool {
	return this.subsetOf(this, other)
}

func (this *NowStr) VisitChildren(visitor Visitor) (Expression, error) {
	return this.visitChildren(this, visitor)
}

func (this *NowStr) eval(args value.Values) (value.Value, error) {
	panic("Cannot eval without context.")
}

func (this *NowStr) MinArgs() int { return 0 }

func (this *NowStr) MaxArgs() int { return 1 }

func (this *NowStr) Constructor() FunctionConstructor { return NewNowStr }

type StrToMillis struct {
	unaryBase
}

func NewStrToMillis(arg Expression) Function {
	return &StrToMillis{
		unaryBase{
			operand: arg,
		},
	}
}

func (this *StrToMillis) Evaluate(item value.Value, context Context) (value.Value, error) {
	return this.evaluate(this, item, context)
}

func (this *StrToMillis) EquivalentTo(other Expression) bool {
	return this.equivalentTo(this, other)
}

func (this *StrToMillis) Fold() (Expression, error) {
	return this.fold(this)
}

func (this *StrToMillis) Formalize(forbidden, allowed value.Value, keyspace string) (Expression, error) {
	return this.formalize(this, forbidden, allowed, keyspace)
}

func (this *StrToMillis) SubsetOf(other Expression) bool {
	return this.subsetOf(this, other)
}

func (this *StrToMillis) VisitChildren(visitor Visitor) (Expression, error) {
	return this.visitChildren(this, visitor)
}

func (this *StrToMillis) eval(arg value.Value) (value.Value, error) {
	if arg.Type() == value.MISSING {
		return value.MISSING_VALUE, nil
	} else if arg.Type() != value.STRING {
		return value.NULL_VALUE, nil
	}

	str := arg.Actual().(string)
	t, e := strToTime(str)
	if e != nil {
		return value.NULL_VALUE, nil
	}

	return value.NewValue(timeToMillis(t)), nil
}

func (this *StrToMillis) Constructor() FunctionConstructor {
	return func(args Expressions) Function {
		return NewStrToMillis(args[0])
	}
}

type StrToUTC struct {
	unaryBase
}

func NewStrToUTC(arg Expression) Function {
	return &StrToUTC{
		unaryBase{
			operand: arg,
		},
	}
}

func (this *StrToUTC) Evaluate(item value.Value, context Context) (value.Value, error) {
	return this.evaluate(this, item, context)
}

func (this *StrToUTC) EquivalentTo(other Expression) bool {
	return this.equivalentTo(this, other)
}

func (this *StrToUTC) Fold() (Expression, error) {
	return this.fold(this)
}

func (this *StrToUTC) Formalize(forbidden, allowed value.Value, keyspace string) (Expression, error) {
	return this.formalize(this, forbidden, allowed, keyspace)
}

func (this *StrToUTC) SubsetOf(other Expression) bool {
	return this.subsetOf(this, other)
}

func (this *StrToUTC) VisitChildren(visitor Visitor) (Expression, error) {
	return this.visitChildren(this, visitor)
}

func (this *StrToUTC) eval(arg value.Value) (value.Value, error) {
	if arg.Type() == value.MISSING {
		return value.MISSING_VALUE, nil
	} else if arg.Type() != value.STRING {
		return value.NULL_VALUE, nil
	}

	str := arg.Actual().(string)
	t, e := strToTime(str)
	if e != nil {
		return value.NULL_VALUE, nil
	}

	t = t.UTC()
	return value.NewValue(timeToStr(t, str)), nil
}

func (this *StrToUTC) Constructor() FunctionConstructor {
	return func(args Expressions) Function {
		return NewStrToUTC(args[0])
	}
}

type StrToZoneName struct {
	binaryBase
}

func NewStrToZoneName(first, second Expression) Function {
	return &StrToZoneName{
		binaryBase{
			first:  first,
			second: second,
		},
	}
}

func (this *StrToZoneName) Evaluate(item value.Value, context Context) (value.Value, error) {
	return this.evaluate(this, item, context)
}

func (this *StrToZoneName) EquivalentTo(other Expression) bool {
	return this.equivalentTo(this, other)
}

func (this *StrToZoneName) Fold() (Expression, error) {
	return this.fold(this)
}

func (this *StrToZoneName) Formalize(forbidden, allowed value.Value, keyspace string) (Expression, error) {
	return this.formalize(this, forbidden, allowed, keyspace)
}

func (this *StrToZoneName) SubsetOf(other Expression) bool {
	return this.subsetOf(this, other)
}

func (this *StrToZoneName) VisitChildren(visitor Visitor) (Expression, error) {
	return this.visitChildren(this, visitor)
}

func (this *StrToZoneName) eval(first, second value.Value) (value.Value, error) {
	if first.Type() == value.MISSING || second.Type() == value.MISSING {
		return value.MISSING_VALUE, nil
	} else if first.Type() != value.STRING || second.Type() != value.STRING {
		return value.NULL_VALUE, nil
	}

	str := first.Actual().(string)
	t, e := strToTime(str)
	if e != nil {
		return value.NULL_VALUE, nil
	}

	tz := second.Actual().(string)
	loc, e := time.LoadLocation(tz)
	if e != nil {
		return value.NULL_VALUE, nil
	}

	return value.NewValue(timeToStr(t.In(loc), str)), nil
}

func (this *StrToZoneName) Constructor() FunctionConstructor {
	return func(args Expressions) Function {
		return NewStrToZoneName(args[0], args[1])
	}
}

func strToTime(s string) (time.Time, error) {
	var t time.Time
	var err error
	for _, f := range _DATE_FORMATS {
		t, err = time.Parse(f, s)
		if err == nil {
			return t, nil
		}
	}

	return t, err
}

func timeToStr(t time.Time, format string) string {
	return t.Format(format)
}

func millisToTime(millis float64) time.Time {
	return time.Unix(0, int64(millis*1000000.0))
}

func timeToMillis(t time.Time) float64 {
	return float64(t.UnixNano() / 1000000)
}

var _DATE_FORMATS = []string{
	"2006-01-02T15:04:05.999Z07:00", // time.RFC3339Milli
	"2006-01-02T15:04:05Z07:00",     // time.RFC3339
	"2006-01-02T15:04:05.999",
	"2006-01-02T15:04:05",
	"2006-01-02 15:04:05.999Z07:00",
	"2006-01-02 15:04:05Z07:00",
	"2006-01-02 15:04:05.999",
	"2006-01-02 15:04:05",
	"2006-01-02",
	"15:04:05.999Z07:00",
	"15:04:05Z07:00",
	"15:04:05.999",
	"15:04:05",
}

const _DEFAULT_FORMAT = "2006-01-02T15:04:05.999Z07:00"

var _DEFAULT_FMT_VALUE = value.NewValue(_DEFAULT_FORMAT)

func datePart(t time.Time, part string) (int, error) {
	p := strings.ToLower(part)

	switch p {
	case "millennium":
		return (t.Year() / 1000) + 1, nil
	case "century":
		return (t.Year() / 100) + 1, nil
	case "decade":
		return t.Year() / 10, nil
	case "year":
		return t.Year(), nil
	case "quarter":
		return (int(t.Month()) + 2) / 3, nil
	case "month":
		return int(t.Month()), nil
	case "day":
		return t.Day(), nil
	case "hour":
		return t.Hour(), nil
	case "minute":
		return t.Minute(), nil
	case "second":
		return t.Second(), nil
	case "millisecond":
		return t.Nanosecond() / int(time.Millisecond), nil
	case "week":
		return int(math.Ceil(float64(t.YearDay()) / 7.0)), nil
	case "day_of_year", "doy":
		return t.YearDay(), nil
	case "day_of_week", "dow":
		return int(t.Weekday()), nil
	case "iso_week":
		_, w := t.ISOWeek()
		return w, nil
	case "iso_year":
		y, _ := t.ISOWeek()
		return y, nil
	case "iso_dow":
		d := int(t.Weekday())
		if d == 0 {
			d = 7
		}
		return d, nil
	case "timezone":
		_, z := t.Zone()
		return z, nil
	case "timezone_hour":
		_, z := t.Zone()
		return z / (60 * 60), nil
	case "timezone_minute":
		_, z := t.Zone()
		zh := z / (60 * 60)
		z = z - (zh * (60 * 60))
		return z / 60, nil
	default:
		return 0, fmt.Errorf("Unsupported date part %s.", part)
	}
}

func dateAdd(t time.Time, n int, part string) (time.Time, error) {
	p := strings.ToLower(part)

	switch p {
	case "millennium":
		return t.AddDate(n*1000, 0, 0), nil
	case "century":
		return t.AddDate(n*100, 0, 0), nil
	case "decade":
		return t.AddDate(n*10, 0, 0), nil
	case "year":
		return t.AddDate(n, 0, 0), nil
	case "quarter":
		return t.AddDate(0, n*3, 0), nil
	case "month":
		return t.AddDate(0, n, 0), nil
	case "week":
		return t.AddDate(0, 0, n*7), nil
	case "day":
		return t.AddDate(0, 0, n), nil
	case "hour":
		return t.Add(time.Duration(n) * time.Hour), nil
	case "minute":
		return t.Add(time.Duration(n) * time.Minute), nil
	case "second":
		return t.Add(time.Duration(n) * time.Second), nil
	case "millisecond":
		return t.Add(time.Duration(n) * time.Millisecond), nil
	default:
		return t, fmt.Errorf("Unsupported date add part %s.", part)
	}
}

func dateTrunc(t time.Time, part string) (time.Time, error) {
	p := strings.ToLower(part)

	switch p {
	case "millennium":
		t = yearTrunc(t)
		return t.AddDate(-(t.Year() % 1000), 0, 0), nil
	case "century":
		t = yearTrunc(t)
		return t.AddDate(-(t.Year() % 100), 0, 0), nil
	case "decade":
		t = yearTrunc(t)
		return t.AddDate(-(t.Year() % 10), 0, 0), nil
	case "year":
		return yearTrunc(t), nil
	case "quarter":
		t = monthTrunc(t)
		return t.AddDate(0, -((int(t.Month()) - 1) % 3), 0), nil
	case "month":
		return monthTrunc(t), nil
	default:
		return timeTrunc(t, p)
	}
}

func yearTrunc(t time.Time) time.Time {
	t, _ = timeTrunc(t, "day")
	return t.AddDate(0, 0, 1-t.YearDay())
}

func monthTrunc(t time.Time) time.Time {
	t, _ = timeTrunc(t, "day")
	return t.AddDate(0, 0, 1-t.Day())
}

func timeTrunc(t time.Time, part string) (time.Time, error) {
	switch part {
	case "day":
		return t.Truncate(time.Duration(24) * time.Hour), nil
	case "hour":
		return t.Truncate(time.Hour), nil
	case "minute":
		return t.Truncate(time.Minute), nil
	case "second":
		return t.Truncate(time.Second), nil
	case "millisecond":
		return t.Truncate(time.Millisecond), nil
	default:
		return t, fmt.Errorf("Unsupported date trunc part %s.", part)
	}
}

func dateDiff(t1, t2 time.Time, part string) (int64, error) {
	diff := diffDates(t1, t2)
	return diffPart(t1, t2, diff, part)
}

func diffPart(t1, t2 time.Time, diff *date, part string) (int64, error) {
	p := strings.ToLower(part)

	switch p {
	case "millisecond":
		sec, e := diffPart(t1, t2, diff, "second")
		if e != nil {
			return 0, e
		}
		return (sec * 1000) + int64(diff.millisecond), nil
	case "second":
		min, e := diffPart(t1, t2, diff, "min")
		if e != nil {
			return 0, e
		}
		return (min * 60) + int64(diff.second), nil
	case "minute":
		hour, e := diffPart(t1, t2, diff, "hour")
		if e != nil {
			return 0, e
		}
		return (hour * 60) + int64(diff.minute), nil
	case "hour":
		day, e := diffPart(t1, t2, diff, "day")
		if e != nil {
			return 0, e
		}
		return (day * 24) + int64(diff.hour), nil
	case "day":
		days := (diff.year * 365) + diff.doy
		if diff.year != 0 {
			days += leapYearsBetween(t1.Year(), t2.Year())
		}
		return int64(days), nil
	case "week":
		day, e := diffPart(t1, t2, diff, "day")
		if e != nil {
			return 0, e
		}
		return day / 7, nil
	case "year":
		return int64(diff.year), nil
	case "decade":
		return int64(diff.year) / 10, nil
	case "century":
		return int64(diff.year) / 100, nil
	case "millenium":
		return int64(diff.year) / 1000, nil
	default:
		return 0, fmt.Errorf("Unsupported date diff part %s.", part)
	}
}

func diffDates(t1, t2 time.Time) *date {
	var d1, d2, diff date
	setDate(&d1, t1)
	setDate(&d2, t2)

	if d1.millisecond < d2.millisecond {
		d1.millisecond += 1000
		d1.second--
	}
	diff.millisecond = d1.millisecond - d2.millisecond

	if d1.second < d2.second {
		d1.second += 60
		d1.minute--
	}
	diff.second = d1.second - d2.second

	if d1.minute < d2.minute {
		d1.minute += 60
		d1.hour--
	}
	diff.minute = d1.minute - d2.minute

	if d1.hour < d2.hour {
		d1.hour += 24
		d1.doy--
	}
	diff.hour = d1.hour - d2.hour

	if d1.doy < d2.doy {
		if isLeapYear(d2.year) {
			d2.doy -= 366
		} else {
			d2.doy -= 365
		}
		d2.year++
	}
	diff.doy = d1.doy - d2.doy

	diff.year = d1.year - d2.year
	return &diff
}

type date struct {
	year        int
	doy         int
	hour        int
	minute      int
	second      int
	millisecond int
}

func setDate(d *date, t time.Time) {
	d.year = t.Year()
	d.doy = t.YearDay()
	d.hour, d.minute, d.second = t.Clock()
	d.millisecond = t.Nanosecond() / 1000000
}

func leapYearsBetween(end, start int) int {
	return leapYearsWithin(end) - leapYearsWithin(start)
}

func leapYearsWithin(year int) int {
	if year > 0 {
		year--
	} else {
		year++
	}

	return (year / 4) - (year / 100) + (year / 400)
}

func isLeapYear(year int) bool {
	return year%400 == 0 || (year%4 == 0 && year%100 != 0)
}

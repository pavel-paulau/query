//  Copyright (c) 2017 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
//  except in compliance with the License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing, software distributed under the
//  License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
//  either express or implied. See the License for the specific language governing permissions
//  and limitations under the License.

package execution

import (
	"encoding/json"
	"fmt"

	"github.com/couchbase/query/errors"
	"github.com/couchbase/query/plan"
	"github.com/couchbase/query/value"
)

type AnsiNest struct {
	base
	plan      *plan.AnsiNest
	child     Operator
	ansiFlags uint32
}

func NewAnsiNest(plan *plan.AnsiNest, context *Context, child Operator) *AnsiNest {
	rv := &AnsiNest{
		plan:  plan,
		child: child,
	}

	newBase(&rv.base, context)
	rv.trackChildren(1)
	rv.execPhase = ANSI_NEST
	rv.output = rv
	return rv
}

func (this *AnsiNest) Accept(visitor Visitor) (interface{}, error) {
	return visitor.VisitAnsiNest(this)
}

func (this *AnsiNest) Copy() Operator {
	rv := &AnsiNest{
		plan:  this.plan,
		child: this.child.Copy(),
	}
	this.base.copy(&rv.base)
	return rv
}

func (this *AnsiNest) RunOnce(context *Context, parent value.Value) {
	this.runConsumer(this, context, parent)
}

func (this *AnsiNest) beforeItems(context *Context, parent value.Value) bool {
	if !context.assert(this.child != nil, "ANSI NEST has no child") {
		return false
	}
	if !context.assert(this.plan.Onclause() != nil, "ANSI NEST does not have onclause") {
		return false
	}

	// check for constant TRUE or FALSE onclause
	cpred := this.plan.Onclause().Value()
	if cpred != nil {
		if cpred.Truth() {
			this.ansiFlags |= ANSI_ONCLAUSE_TRUE
		} else {
			this.ansiFlags |= ANSI_ONCLAUSE_FALSE
		}
	}

	return true
}

func (this *AnsiNest) processItem(item value.AnnotatedValue, context *Context) bool {
	defer this.switchPhase(_EXECTIME)

	if (this.ansiFlags & ANSI_REOPEN_CHILD) != 0 {
		if this.child != nil {
			this.child.SendStop()
			this.child.reopen(context)
		}
	} else {
		this.ansiFlags |= ANSI_REOPEN_CHILD
	}

	this.child.SetOutput(this.child)
	this.child.SetInput(nil)
	this.child.SetParent(this)
	this.child.SetStop(nil)

	go this.child.RunOnce(context, item)

	var right_items value.AnnotatedValues
	ok := true
	stopped := false
	n := 1

loop:
	for ok {
		right_item, child, cont := this.getItemChildrenOp(this.child)
		if cont {
			if right_item != nil {
				var match bool
				match, ok, _ = processAnsiExec(item, right_item, this.plan.Onclause(),
					this.plan.Alias(), this.ansiFlags, context, "nest")
				if ok && match {
					right_items = append(right_items, right_item)
				}
			} else if child >= 0 {
				n--
			} else {
				break loop
			}
		} else {
			stopped = true
			break loop
		}
	}

	if n > 0 {
		notifyChildren(this.child)
		this.childrenWaitNoStop(n)
	}

	if stopped || !ok {
		return false
	}

	return this.processAnsiNest(item, right_items, context)
}

func (this *AnsiNest) processAnsiNest(item value.AnnotatedValue, right_items value.AnnotatedValues, context *Context) bool {

	joined := item
	alias := this.plan.Alias()

	if len(right_items) == 0 {
		if this.plan.Outer() {
			joined.SetField(alias, value.EMPTY_ARRAY_VALUE)
			return this.sendItem(joined)
		} else {
			return true
		}
	}

	vals := make([]interface{}, 0, len(right_items))

	for _, right_item := range right_items {
		// only interested in the value corresponding to "alias"
		val, ok := right_item.Field(alias)
		if !ok {
			context.Error(errors.NewExecutionInternalError(fmt.Sprintf("processAnsiNest: annotated value not found for %s", alias)))
			return false
		}

		vals = append(vals, val)
	}

	joined.SetField(alias, vals)

	return this.sendItem(joined)
}

func (this *AnsiNest) MarshalJSON() ([]byte, error) {
	r := this.plan.MarshalBase(func(r map[string]interface{}) {
		this.marshalTimes(r)
		r["~child"] = this.child
	})
	return json.Marshal(r)
}

func (this *AnsiNest) SendStop() {
	this.baseSendStop()
	if this.child != nil {
		this.child.SendStop()
	}
}

func (this *AnsiNest) reopen(context *Context) {
	this.baseReopen(context)
	this.ansiFlags &^= ANSI_REOPEN_CHILD
	if this.child != nil {
		this.child.reopen(context)
	}
}

func (this *AnsiNest) Done() {
	this.baseDone()
	if this.child != nil {
		this.child.Done()
	}
	this.child = nil
}

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
	"github.com/couchbaselabs/query/value"
)

type Not struct {
	unaryBase
}

func NewNot(operand Expression) Expression {
	return &Not{
		unaryBase{
			operand: operand,
		},
	}
}

func (this *Not) Evaluate(item value.Value, context Context) (value.Value, error) {
	return this.evaluate(this, item, context)
}

func (this *Not) EquivalentTo(other Expression) bool {
	return this.equivalentTo(this, other)
}

func (this *Not) Fold() (Expression, error) {
	t, e := this.VisitChildren(&Folder{})
	if e != nil {
		return t, e
	}

	switch o := this.operand.(type) {
	case *Constant:
		v, e := this.eval(o.Value())
		if e != nil {
			return nil, e
		}
		return NewConstant(v), nil
	case *Not:
		return o.operand, nil
	}

	return this, nil
}

func (this *Not) Formalize(forbidden, allowed value.Value, keyspace string) (Expression, error) {
	return this.formalize(this, forbidden, allowed, keyspace)
}

func (this *Not) SubsetOf(other Expression) bool {
	return this.subsetOf(this, other)
}

func (this *Not) VisitChildren(visitor Visitor) (Expression, error) {
	return this.visitChildren(this, visitor)
}

func (this *Not) eval(operand value.Value) (value.Value, error) {
	if operand.Type() > value.NULL {
		return value.NewValue(!operand.Truth()), nil
	} else {
		return operand, nil
	}
}

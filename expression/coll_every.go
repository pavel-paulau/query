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

type Every struct {
	collPred
}

func NewEvery(bindings Bindings, satisfies Expression) Expression {
	return &Every{
		collPred: collPred{
			bindings:  bindings,
			satisfies: satisfies,
		},
	}
}

func (this *Every) Evaluate(item value.Value, context Context) (value.Value, error) {
	missing := false
	null := false

	barr := make([][]interface{}, len(this.bindings))
	for i, b := range this.bindings {
		bv, err := b.Expression().Evaluate(item, context)
		if err != nil {
			return nil, err
		}

		if b.Descend() {
			buffer := make([]interface{}, 0, 256)
			bv = value.NewValue(bv.Descendants(buffer))
		}

		switch bv.Type() {
		case value.ARRAY:
			barr[i] = bv.Actual().([]interface{})
		case value.MISSING:
			missing = true
		default:
			null = true
		}
	}

	if missing {
		return value.MISSING_VALUE, nil
	}

	if null {
		return value.NULL_VALUE, nil
	}

	n := -1
	for _, b := range barr {
		if n < 0 || len(b) < n {
			n = len(b)
		}
	}

	for i := 0; i < n; i++ {
		cv := value.NewScopeValue(make(map[string]interface{}, len(this.bindings)), item)
		for j, b := range this.bindings {
			cv.SetField(b.Variable(), barr[j][i])
		}

		sv, e := this.satisfies.Evaluate(cv, context)
		if e != nil {
			return nil, e
		}

		if !sv.Truth() {
			return value.NewValue(false), nil
		}
	}

	return value.NewValue(true), nil
}

func (this *Every) EquivalentTo(other Expression) bool {
	return this.equivalentTo(this, other)
}

func (this *Every) Fold() (Expression, error) {
	return this.fold(this)
}

func (this *Every) Formalize(forbidden, allowed value.Value, keyspace string) (Expression, error) {
	return this.formalize(this, forbidden, allowed, keyspace)
}

func (this *Every) SubsetOf(other Expression) bool {
	return this.subsetOf(this, other)
}

func (this *Every) VisitChildren(visitor Visitor) (Expression, error) {
	return this.visitChildren(this, visitor)
}

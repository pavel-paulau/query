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
	"encoding/base64"

	"github.com/couchbaselabs/query/value"
)

type Base64 struct {
	unaryBase
}

func NewBase64(operand Expression) Function {
	return &Base64{
		unaryBase{
			operand: operand,
		},
	}
}

func (this *Base64) Evaluate(item value.Value, context Context) (value.Value, error) {
	return this.evaluate(this, item, context)
}

func (this *Base64) EquivalentTo(other Expression) bool {
	return this.equivalentTo(this, other)
}

func (this *Base64) Fold() (Expression, error) {
	return this.fold(this)
}

func (this *Base64) Formalize(forbidden, allowed value.Value, keyspace string) (Expression, error) {
	return this.formalize(this, forbidden, allowed, keyspace)
}

func (this *Base64) SubsetOf(other Expression) bool {
	return this.subsetOf(this, other)
}

func (this *Base64) VisitChildren(visitor Visitor) (Expression, error) {
	return this.visitChildren(this, visitor)
}

func (this *Base64) eval(operand value.Value) (value.Value, error) {
	if operand.Type() == value.MISSING {
		return operand, nil
	}

	str := base64.StdEncoding.EncodeToString(operand.Bytes())
	return value.NewValue(str), nil
}

func (this *Base64) Constructor() FunctionConstructor {
	return func(args Expressions) Function {
		return NewBase64(args[0])
	}
}

type Meta struct {
	unaryBase
}

func NewMeta(operand Expression) Function {
	return &Meta{
		unaryBase{
			operand: operand,
		},
	}
}

func (this *Meta) Evaluate(item value.Value, context Context) (value.Value, error) {
	return this.evaluate(this, item, context)
}

func (this *Meta) EquivalentTo(other Expression) bool {
	return this.equivalentTo(this, other)
}

func (this *Meta) Fold() (Expression, error) {
	return this.fold(this)
}

func (this *Meta) Formalize(forbidden, allowed value.Value, keyspace string) (Expression, error) {
	return this.formalize(this, forbidden, allowed, keyspace)
}

func (this *Meta) SubsetOf(other Expression) bool {
	return this.subsetOf(this, other)
}

func (this *Meta) VisitChildren(visitor Visitor) (Expression, error) {
	return this.visitChildren(this, visitor)
}

func (this *Meta) eval(operand value.Value) (value.Value, error) {
	if operand.Type() == value.MISSING {
		return operand, nil
	}

	switch operand := operand.(type) {
	case value.AnnotatedValue:
		return value.NewValue(operand.GetAttachment("meta")), nil
	default:
		return value.NULL_VALUE, nil
	}
}

func (this *Meta) Constructor() FunctionConstructor {
	return func(args Expressions) Function {
		return NewMeta(args[0])
	}
}

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
	"regexp"

	"github.com/couchbaselabs/query/value"
)

type Like struct {
	reBinaryBase
}

func NewLike(first, second Expression) Expression {
	return &Like{
		reBinaryBase{
			binaryBase: binaryBase{
				first:  first,
				second: second,
			},
		},
	}
}

func (this *Like) Evaluate(item value.Value, context Context) (value.Value, error) {
	return this.evaluate(this, item, context)
}

func (this *Like) EquivalentTo(other Expression) bool {
	return this.equivalentTo(this, other)
}

func (this *Like) Fold() (Expression, error) {
	return this.fold(this)
}

func (this *Like) Formalize(forbidden, allowed value.Value, keyspace string) (Expression, error) {
	return this.formalize(this, forbidden, allowed, keyspace)
}

func (this *Like) SubsetOf(other Expression) bool {
	return this.subsetOf(this, other)
}

func (this *Like) VisitChildren(visitor Visitor) (Expression, error) {
	return this.visitChildren(this, visitor)
}

func (this *Like) eval(first, second value.Value) (value.Value, error) {
	if first.Type() == value.MISSING || second.Type() == value.MISSING {
		return value.MISSING_VALUE, nil
	} else if first.Type() != value.STRING || second.Type() != value.STRING {
		return value.NULL_VALUE, nil
	}

	f := first.Actual().(string)
	s := second.Actual().(string)

	re := this.re
	if re == nil {
		var e error
		re, e = this.compile(s)
		if e != nil {
			return nil, e
		}
	}

	return value.NewValue(re.MatchString(f)), nil
}

func (this *Like) compile(s string) (*regexp.Regexp, error) {
	repl := regexp.MustCompile("\\\\|\\_|\\%|_|%")
	s = repl.ReplaceAllStringFunc(s, replacer)

	re, err := regexp.Compile(s)
	if err != nil {
		return nil, err
	}

	return re, nil
}

func replacer(s string) string {
	switch s {
	case "\\\\":
		return "\\"
	case "\\_":
		return "_"
	case "\\%":
		return "%"
	case "_":
		return "(.)"
	case "%":
		return "(.*)"
	default:
		panic("Unknown regexp replacer " + s)
	}
}

func NewNotLike(first, second Expression) Expression {
	return NewNot(NewLike(first, second))
}

//  Copyright (c) 2014 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
//  except in compliance with the License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing, software distributed under the
//  License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
//  either express or implied. See the License for the specific language governing permissions
//  and limitations under the License.

package planner

import (
	"github.com/couchbase/query/algebra"
	"github.com/couchbase/query/errors"
	"github.com/couchbase/query/expression"
	"github.com/couchbase/query/plan"
	"github.com/couchbase/query/value"
)

func (this *builder) visitFrom(node *algebra.Subselect, group *algebra.Group) error {
	count, err := this.fastCount(node)
	if err != nil {
		return err
	}

	if count {
		this.maxParallelism = 1
		this.resetPushDowns()
	} else if node.From() != nil {
		prevFrom := this.from
		this.from = node.From()
		defer func() { this.from = prevFrom }()

		// gather keyspace references
		this.baseKeyspaces = make(map[string]*baseKeyspace, _MAP_KEYSPACE_CAP)
		keyspaceFinder := newKeyspaceFinder(this.baseKeyspaces)
		_, err := node.From().Accept(keyspaceFinder)
		if err != nil {
			return err
		}
		this.pushableOnclause = keyspaceFinder.pushableOnclause

		// Process where clause and pushable on clause
		if this.where != nil {
			err = this.processWhere(this.where)
			if err != nil {
				return err
			}
		}

		if this.pushableOnclause != nil {
			err = this.processPredicate(this.pushableOnclause, true)
			if err != nil {
				return err
			}
		}

		// Use FROM clause in index selection
		_, err = node.From().Accept(this)
		if err != nil {
			return err
		}
	} else {
		// No FROM clause
		this.resetPushDowns()
		scan := plan.NewDummyScan()
		this.children = append(this.children, scan)
		this.maxParallelism = 1
	}

	return nil
}

func (this *builder) VisitKeyspaceTerm(node *algebra.KeyspaceTerm) (interface{}, error) {
	node.SetDefaultNamespace(this.namespace)
	keyspace, err := this.getTermKeyspace(node)
	if err != nil {
		return nil, err
	}

	if this.subquery && this.correlated && node.Keys() == nil {
		return nil, errors.NewSubqueryMissingKeysError(node.Keyspace())
	}

	scan, err := this.selectScan(keyspace, node)
	if err != nil {
		return nil, err
	}

	if scan == nil {
		if node.IsPrimaryJoin() {
			return nil, nil
		} else {
			return nil, errors.NewPlanInternalError("VisitKeyspaceTerm: no plan generated")
		}
	}
	this.children = append(this.children, scan)

	if len(this.coveringScans) == 0 && this.countScan == nil {
		fetch := plan.NewFetch(keyspace, node)
		this.children = append(this.children, fetch)
	}

	err = this.processKeyspaceDone(node.Alias())
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (this *builder) VisitSubqueryTerm(node *algebra.SubqueryTerm) (interface{}, error) {
	sel, err := node.Subquery().Accept(this)
	if err != nil {
		return nil, err
	}

	this.resetPushDowns()

	this.children = make([]plan.Operator, 0, 16)    // top-level children, executed sequentially
	this.subChildren = make([]plan.Operator, 0, 16) // sub-children, executed across data-parallel streams
	this.children = append(this.children, sel.(plan.Operator), plan.NewAlias(node.Alias()))

	err = this.processKeyspaceDone(node.Alias())
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (this *builder) VisitExpressionTerm(node *algebra.ExpressionTerm) (interface{}, error) {
	if node.IsKeyspace() {
		return node.KeyspaceTerm().Accept(this)
	}

	this.resetPushDowns()

	this.children = make([]plan.Operator, 0, 16)    // top-level children, executed sequentially
	this.subChildren = make([]plan.Operator, 0, 16) // sub-children, executed across data-parallel streams

	scan := plan.NewExpressionScan(node.ExpressionTerm(), node.Alias())
	this.children = append(this.children, scan)

	err := this.processKeyspaceDone(node.Alias())
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (this *builder) VisitJoin(node *algebra.Join) (interface{}, error) {
	this.resetProjection()
	this.resetIndexGroupAggs()
	if term, ok := node.PrimaryTerm().(*algebra.ExpressionTerm); ok && term.IsKeyspace() {
		this.resetOffsetLimit()
	} else {
		this.resetOrderOffsetLimit()
	}

	_, err := node.Left().Accept(this)
	if err != nil {
		return nil, err
	}

	right := node.Right()
	right.SetDefaultNamespace(this.namespace)
	namespace, err := this.datastore.NamespaceByName(right.Namespace())
	if err != nil {
		return nil, err
	}

	keyspace, err := namespace.KeyspaceByName(right.Keyspace())
	if err != nil {
		return nil, err
	}

	join := plan.NewJoin(keyspace, node)
	if len(this.subChildren) > 0 {
		parallel := plan.NewParallel(plan.NewSequence(this.subChildren...), this.maxParallelism)
		this.children = append(this.children, parallel)
		this.subChildren = make([]plan.Operator, 0, 16)
	}
	this.children = append(this.children, join)

	err = this.processKeyspaceDone(node.Alias())
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (this *builder) VisitIndexJoin(node *algebra.IndexJoin) (interface{}, error) {
	this.requirePrimaryKey = true
	this.resetIndexGroupAggs()
	this.resetProjection()
	if term, ok := node.PrimaryTerm().(*algebra.ExpressionTerm); ok && term.IsKeyspace() {
		this.resetOffsetLimit()
	} else {
		this.resetOrderOffsetLimit()
	}

	_, err := node.Left().Accept(this)
	if err != nil {
		return nil, err
	}

	right := node.Right()
	right.SetDefaultNamespace(this.namespace)
	namespace, err := this.datastore.NamespaceByName(right.Namespace())
	if err != nil {
		return nil, err
	}

	keyspace, err := namespace.KeyspaceByName(right.Keyspace())
	if err != nil {
		return nil, err
	}

	join, err := this.buildIndexJoin(keyspace, node)
	if err != nil {
		return nil, err
	}

	this.subChildren = append(this.subChildren, join)

	err = this.processKeyspaceDone(node.Alias())
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (this *builder) VisitAnsiJoin(node *algebra.AnsiJoin) (interface{}, error) {
	this.requirePrimaryKey = true
	this.resetIndexGroupAggs()
	this.resetProjection()
	if term, ok := node.PrimaryTerm().(*algebra.ExpressionTerm); ok && term.IsKeyspace() {
		this.resetOffsetLimit()
	} else {
		this.resetOrderOffsetLimit()
	}

	_, err := node.Left().Accept(this)
	if err != nil {
		return nil, err
	}

	join, err := this.buildAnsiJoin(node)
	if err != nil {
		return nil, err
	}

	if njoin, ok := join.(*plan.Join); ok {
		if len(this.subChildren) > 0 {
			parallel := plan.NewParallel(plan.NewSequence(this.subChildren...), this.maxParallelism)
			this.children = append(this.children, parallel)
			this.subChildren = make([]plan.Operator, 0, 16)
		}
		this.children = append(this.children, njoin)
	} else {
		this.subChildren = append(this.subChildren, join)
	}

	err = this.processKeyspaceDone(node.Alias())
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (this *builder) VisitNest(node *algebra.Nest) (interface{}, error) {
	this.resetIndexGroupAggs()
	this.resetProjection()

	if this.hasOffsetOrLimit() && !node.Outer() {
		this.resetOffsetLimit()
	}

	_, err := node.Left().Accept(this)
	if err != nil {
		return nil, err
	}

	right := node.Right()
	right.SetDefaultNamespace(this.namespace)
	namespace, err := this.datastore.NamespaceByName(right.Namespace())
	if err != nil {
		return nil, err
	}

	keyspace, err := namespace.KeyspaceByName(right.Keyspace())
	if err != nil {
		return nil, err
	}

	if len(this.subChildren) > 0 {
		parallel := plan.NewParallel(plan.NewSequence(this.subChildren...), this.maxParallelism)
		this.children = append(this.children, parallel)
		this.subChildren = make([]plan.Operator, 0, 16)
	}

	nest := plan.NewNest(keyspace, node)
	this.children = append(this.children, nest)

	err = this.processKeyspaceDone(node.Alias())
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (this *builder) VisitIndexNest(node *algebra.IndexNest) (interface{}, error) {
	this.requirePrimaryKey = true
	this.resetIndexGroupAggs()
	this.resetProjection()

	if this.hasOffsetOrLimit() && !node.Outer() {
		this.resetOffsetLimit()
	}

	_, err := node.Left().Accept(this)
	if err != nil {
		return nil, err
	}

	right := node.Right()
	right.SetDefaultNamespace(this.namespace)
	namespace, err := this.datastore.NamespaceByName(right.Namespace())
	if err != nil {
		return nil, err
	}

	keyspace, err := namespace.KeyspaceByName(right.Keyspace())
	if err != nil {
		return nil, err
	}

	nest, err := this.buildIndexNest(keyspace, node)
	if err != nil {
		return nil, err
	}

	this.subChildren = append(this.subChildren, nest)

	err = this.processKeyspaceDone(node.Alias())
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (this *builder) VisitAnsiNest(node *algebra.AnsiNest) (interface{}, error) {
	this.requirePrimaryKey = true
	this.resetIndexGroupAggs()
	this.resetProjection()

	if this.hasOffsetOrLimit() && !node.Outer() {
		this.resetOffsetLimit()
	}

	_, err := node.Left().Accept(this)
	if err != nil {
		return nil, err
	}

	nest, err := this.buildAnsiNest(node)
	if err != nil {
		return nil, err
	}

	if nnest, ok := nest.(*plan.Nest); ok {
		if len(this.subChildren) > 0 {
			parallel := plan.NewParallel(plan.NewSequence(this.subChildren...), this.maxParallelism)
			this.children = append(this.children, parallel)
			this.subChildren = make([]plan.Operator, 0, 16)
		}
		this.children = append(this.children, nnest)
	} else {
		this.subChildren = append(this.subChildren, nest)
	}

	err = this.processKeyspaceDone(node.Alias())
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (this *builder) VisitUnnest(node *algebra.Unnest) (interface{}, error) {
	if term, ok := node.PrimaryTerm().(*algebra.ExpressionTerm); !ok || !term.IsKeyspace() {
		this.resetPushDowns()
	}

	_, err := node.Left().Accept(this)
	if err != nil {
		return nil, err
	}

	_, found := this.coveredUnnests[node]
	if found {
		return nil, nil
	}

	unnest := plan.NewUnnest(node)
	this.subChildren = append(this.subChildren, unnest)
	parallel := plan.NewParallel(plan.NewSequence(this.subChildren...), this.maxParallelism)
	this.children = append(this.children, parallel)
	this.subChildren = make([]plan.Operator, 0, 16)

	err = this.processKeyspaceDone(node.Alias())
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (this *builder) fastCount(node *algebra.Subselect) (bool, error) {
	if node.From() == nil ||
		(node.Where() != nil && (node.Where().Value() == nil || !node.Where().Value().Truth())) ||
		node.Group() != nil {
		return false, nil
	}

	var from *algebra.KeyspaceTerm
	switch other := node.From().(type) {
	case *algebra.KeyspaceTerm:
		from = other
	case *algebra.ExpressionTerm:
		if other.IsKeyspace() {
			from = other.KeyspaceTerm()
		} else {
			return false, nil
		}
	default:
		return false, nil
	}

	if from == nil || from.Keys() != nil {
		return false, nil
	}

	from.SetDefaultNamespace(this.namespace)
	keyspace, err := this.getTermKeyspace(from)
	if err != nil {
		return false, err
	}

	for _, term := range node.Projection().Terms() {
		count, ok := term.Expression().(*algebra.Count)
		if !ok {
			return false, nil
		}

		operand := count.Operand()
		if operand != nil {
			val := operand.Value()
			if val == nil || val.Type() <= value.NULL {
				return false, nil
			}
		}
	}

	scan := plan.NewCountScan(keyspace, from)
	this.children = append(this.children, scan)
	return true, nil
}

func (this *builder) resetOrderOffsetLimit() {
	this.resetOrder()
	this.resetLimit()
	this.resetOffset()
}

func (this *builder) resetOffsetLimit() {
	this.resetLimit()
	this.resetOffset()
}

func (this *builder) resetLimit() {
	this.limit = nil
}

func (this *builder) resetOffset() {
	this.offset = nil
}

func (this *builder) resetOrder() {
	this.order = nil
}

func (this *builder) hasOrderOrOffsetOrLimit() bool {
	return this.order != nil || this.offset != nil || this.limit != nil
}

func (this *builder) hasOffsetOrLimit() bool {
	return this.offset != nil || this.limit != nil
}

func (this *builder) resetProjection() {
	this.projection = nil
}

func (this *builder) resetIndexGroupAggs() {
	this.oldAggregates = false
	this.group = nil
	this.aggs = nil
	this.aggConstraint = nil
}

func (this *builder) resetPushDowns() {
	this.resetOrderOffsetLimit()
	this.resetProjection()
	this.resetIndexGroupAggs()
}

func offsetPlusLimit(offset, limit expression.Expression) expression.Expression {
	if offset != nil && limit != nil {
		if offset.Value() == nil {
			offset = expression.NewGreatest(expression.ZERO_EXPR, offset)
		}

		if limit.Value() == nil {
			limit = expression.NewGreatest(expression.ZERO_EXPR, limit)
		}

		return expression.NewAdd(limit, offset)
	} else {
		return limit
	}
}

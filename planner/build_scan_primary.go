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
	"fmt"

	"github.com/couchbase/query/algebra"
	"github.com/couchbase/query/datastore"
	"github.com/couchbase/query/expression"
	"github.com/couchbase/query/plan"
)

func (this *builder) buildPrimaryScan(keyspace datastore.Keyspace, node *algebra.KeyspaceTerm,
	limit expression.Expression, hintIndexes, otherIndexes []datastore.Index) (scan *plan.PrimaryScan, err error) {
	primary, err := buildPrimaryIndex(keyspace, hintIndexes, otherIndexes)
	if err != nil {
		return nil, err
	}

	return plan.NewPrimaryScan(primary, keyspace, node, limit), nil
}

func buildPrimaryIndex(keyspace datastore.Keyspace, hintIndexes, otherIndexes []datastore.Index) (
	primary datastore.PrimaryIndex, err error) {
	ok := false

	// Prefer hints
	for _, index := range hintIndexes {
		if !index.IsPrimary() {
			continue
		}

		primary, ok = index.(datastore.PrimaryIndex)
		if ok {
			return
		} else {
			return nil, fmt.Errorf("Unable to cast index %s to primary index", index.Name())
		}
	}

	// Consider other primary indexes
	if otherIndexes != nil {
		for _, index := range otherIndexes {
			if !index.IsPrimary() {
				continue
			}

			primary, ok = index.(datastore.PrimaryIndex)
			if ok {
				return
			} else {
				return nil, fmt.Errorf("Unable to cast index %s to primary index", index.Name())
			}
		}
	}

	// Return first online primary index
	indexers, er := keyspace.Indexers()
	if er != nil {
		return nil, err
	}

	for _, indexer := range indexers {
		primaries, er := indexer.PrimaryIndexes()
		if er != nil {
			return nil, er
		}

		for _, primary = range primaries {
			state, _, er := primary.State()
			if er != nil {
				return nil, er
			}

			if state == datastore.ONLINE {
				return
			}
		}
	}

	if primary == nil {
		return nil, fmt.Errorf(
			"No primary index on keyspace %s. Use CREATE PRIMARY INDEX to create one.",
			keyspace.Name())
	}

	return nil, fmt.Errorf("Primary index %s not online.", primary.Name())
}
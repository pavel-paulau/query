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
	"math"

	"github.com/couchbase/query/datastore"
	"github.com/couchbase/query/errors"
	"github.com/couchbase/query/logging"
	"github.com/couchbase/query/plan"
	"github.com/couchbase/query/value"
)

type PrimaryScan3 struct {
	base
	plan *plan.PrimaryScan3
}

func NewPrimaryScan3(plan *plan.PrimaryScan3, context *Context) *PrimaryScan3 {
	rv := &PrimaryScan3{
		plan: plan,
	}

	newBase(&rv.base, context)
	rv.newStopChannel()
	rv.output = rv
	return rv
}

func (this *PrimaryScan3) Accept(visitor Visitor) (interface{}, error) {
	return visitor.VisitPrimaryScan3(this)
}

func (this *PrimaryScan3) Copy() Operator {
	rv := &PrimaryScan3{plan: this.plan}
	this.base.copy(&rv.base)
	return rv
}

func (this *PrimaryScan3) RunOnce(context *Context, parent value.Value) {
	this.once.Do(func() {
		defer context.Recover() // Recover from any panic
		this.active()
		defer this.close(context)
		this.setExecPhase(PRIMARY_SCAN, context)
		defer this.notify() // Notify that I have stopped

		this.scanPrimary(context, parent)
	})
}

func (this *PrimaryScan3) scanPrimary(context *Context, parent value.Value) {
	this.switchPhase(_EXECTIME)
	defer this.switchPhase(_NOTIME)
	conn := this.newIndexConnection(context)
	defer notifyConn(conn.StopChannel()) // Notify index that I have stopped

	go this.scanEntries(context, conn)

	nitems := 0

	var docs uint64 = 0
	defer func() {
		if docs > 0 {
			context.AddPhaseCount(PRIMARY_SCAN, docs)
		}
	}()

	var lastEntry *datastore.IndexEntry
	for {
		entry, ok := this.getItemEntry(conn.EntryChannel())
		if ok {
			if entry != nil {
				// current policy is to only count 'in' documents
				// from operators, not kv
				// add this.addInDocs(1) if this changes
				cv := value.NewScopeValue(make(map[string]interface{}), parent)
				av := value.NewAnnotatedValue(cv)
				av.SetAttachment("meta", map[string]interface{}{"id": entry.PrimaryKey})
				ok = this.sendItem(av)
				lastEntry = entry
				nitems++
				docs++
				if docs > _PHASE_UPDATE_COUNT {
					context.AddPhaseCount(PRIMARY_SCAN, docs)
					docs = 0
				}
			} else {
				break
			}
		} else {
			return
		}
	}

	if conn.Timeout() {
		// Offset, Aggregates, Order needs to be exact.
		// On timeout return error because we cann't stitch the output
		if this.plan.Offset() != nil || len(this.plan.OrderTerms()) > 0 || this.plan.GroupAggs() != nil {
			context.Error(errors.NewCbIndexScanTimeoutError(nil))
			return
		}

		logging.Errorp("Primary index scan timeout - resorting to chunked scan",
			logging.Pair{"chunkSize", nitems},
			logging.Pair{"startingEntry", lastEntry})
		if lastEntry == nil {
			// no key for chunked scans (primary scan returned 0 items)
			context.Error(errors.NewCbIndexScanTimeoutError(nil))
		}
		// do chunked scans; nitems gives the chunk size, and lastEntry the starting point
		for lastEntry != nil {
			lastEntry = this.scanPrimaryChunk(context, parent, nitems, lastEntry)
		}
	}
}

func (this *PrimaryScan3) scanPrimaryChunk(context *Context, parent value.Value, chunkSize int, indexEntry *datastore.IndexEntry) *datastore.IndexEntry {
	this.switchPhase(_EXECTIME)
	defer this.switchPhase(_NOTIME)
	conn, _ := datastore.NewSizedIndexConnection(int64(chunkSize), context)
	conn.SetPrimary()
	defer notifyConn(conn.StopChannel()) // Notify index that I have stopped

	go this.scanChunk(context, conn, chunkSize, indexEntry)

	nitems := 0
	var docs uint64 = 0
	defer func() {
		if nitems > 0 {
			context.AddPhaseCount(PRIMARY_SCAN, docs)
		}
	}()

	var lastEntry *datastore.IndexEntry
	for {
		entry, ok := this.getItemEntry(conn.EntryChannel())
		if ok {
			if entry != nil {
				cv := value.NewScopeValue(make(map[string]interface{}), parent)
				av := value.NewAnnotatedValue(cv)
				av.SetAttachment("meta", map[string]interface{}{"id": entry.PrimaryKey})
				ok = this.sendItem(av)
				lastEntry = entry
				nitems++
				docs++
				if docs > _PHASE_UPDATE_COUNT {
					context.AddPhaseCount(PRIMARY_SCAN, docs)
					docs = 0
				}
			} else {
				break
			}
		} else {
			return nil
		}
	}
	logging.Debugp("Primary index chunked scan", logging.Pair{"chunkSize", nitems}, logging.Pair{"lastKey", lastEntry})
	return lastEntry
}

func (this *PrimaryScan3) scanEntries(context *Context, conn *datastore.IndexConnection) {
	defer context.Recover() // Recover from any panic

	index := this.plan.Index()
	keyspace := this.plan.Keyspace()
	scanVector := context.ScanVectorSource().ScanVector(keyspace.NamespaceId(), keyspace.Name())
	offset := evalLimitOffset(this.plan.Offset(), nil, int64(0), false, context)
	limit := evalLimitOffset(this.plan.Limit(), nil, math.MaxInt64, false, context)
	indexProjection, indexOrder, indexGroupAggs := planToScanMapping(index, this.plan.Projection(),
		this.plan.OrderTerms(), this.plan.GroupAggs(), nil)

	index.ScanEntries3(context.RequestId(), indexProjection, offset, limit, indexGroupAggs, indexOrder,
		context.ScanConsistency(), scanVector, conn)
}

func (this *PrimaryScan3) scanChunk(context *Context, conn *datastore.IndexConnection, chunkSize int, indexEntry *datastore.IndexEntry) {
	defer context.Recover() // Recover from any panic
	ds := &datastore.Span{}
	// do the scan starting from, but not including, the given index entry:
	ds.Range = datastore.Range{
		Inclusion: datastore.NEITHER,
		Low:       []value.Value{value.NewValue(indexEntry.PrimaryKey)},
	}
	keyspace := this.plan.Keyspace()
	scanVector := context.ScanVectorSource().ScanVector(keyspace.NamespaceId(), keyspace.Name())
	this.plan.Index().Scan(context.RequestId(), ds, true, int64(chunkSize),
		context.ScanConsistency(), scanVector, conn)
}

func (this *PrimaryScan3) newIndexConnection(context *Context) *datastore.IndexConnection {
	var conn *datastore.IndexConnection

	// Use keyspace count to create a sized index connection
	keyspace := this.plan.Keyspace()
	size, err := keyspace.Count(context)
	if err == nil {
		if size <= 0 {
			size = 1
		}

		conn, err = datastore.NewSizedIndexConnection(size, context)
		conn.SetPrimary()
	}

	// Use non-sized API and log error
	if err != nil {
		conn = datastore.NewIndexConnection(context)
		conn.SetPrimary()
		logging.Errorp("PrimaryScan3.newIndexConnection ", logging.Pair{"error", err})
	}

	return conn
}

func (this *PrimaryScan3) MarshalJSON() ([]byte, error) {
	r := this.plan.MarshalBase(func(r map[string]interface{}) {
		this.marshalTimes(r)
	})
	return json.Marshal(r)
}

// send a stop
func (this *PrimaryScan3) SendStop() {
	this.chanSendStop()
}

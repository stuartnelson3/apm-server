// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package sampling

import (
	"context"

	"github.com/elastic/apm-server/model"
	"github.com/elastic/beats/v7/libbeat/monitoring"
)

var (
	monitoringRegistry         = monitoring.Default.NewRegistry("apm-server.sampling")
	transactionsDroppedCounter = monitoring.NewInt(monitoringRegistry, "transactions_dropped")
)

// NewDiscardUnsampledBatchProcessor returns a model.BatchProcessor which
// discards unsampled transactions.
//
// The returned model.BatchProcessor does not guarantee order preservation
// of events retained in the batch.
func NewDiscardUnsampledBatchProcessor() model.BatchProcessor {
	return model.ProcessBatchFunc(func(ctx context.Context, batch *model.Batch) error {
		var dropped int64
		transactions := batch.Transactions
		for i := 0; i < len(transactions); {
			tx := transactions[i]
			if tx.Sampled == nil || *tx.Sampled {
				i++
				continue
			}
			n := len(transactions)
			transactions[i], transactions[n-1] = transactions[n-1], transactions[i]
			transactions = transactions[:n-1]
			dropped++
		}
		if dropped > 0 {
			transactionsDroppedCounter.Add(dropped)
		}
		batch.Transactions = transactions
		return nil
	})
}

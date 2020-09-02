// VulcanizeDB
// Copyright © 2019 Vulcanize

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.

// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package btc

import (
	"strconv"

	"github.com/vulcanize/ipld-btc-indexer/pkg/ipfs/ipld"
	"github.com/vulcanize/ipld-btc-indexer/pkg/postgres"
	"github.com/vulcanize/ipld-btc-indexer/pkg/shared"
)

// Publisher interface for substituting mocks in tests
type Publisher interface {
	Publish(payload ConvertedPayload) error
}

// IPLDPublisher satisfies the IPLDPublisher interface for bitcoin
// It interfaces directly with the public.blocks table of PG-IPFS rather than going through an ipfs intermediary
// It publishes and indexes IPLDs together in a single sqlx.Tx
type IPLDPublisher struct {
	indexer *CIDIndexer
}

// NewIPLDPublisher creates a pointer to a new eth IPLDPublisher which satisfies the IPLDPublisher interface
func NewIPLDPublisher(db *postgres.DB) *IPLDPublisher {
	return &IPLDPublisher{
		indexer: NewCIDIndexer(db),
	}
}

// Publish publishes an IPLDPayload to IPFS and returns the corresponding CIDPayload
func (pub *IPLDPublisher) Publish(payload ConvertedPayload) error {
	// Generate the iplds
	headerNode, txNodes, txTrieNodes, err := ipld.FromHeaderAndTxs(payload.Header, payload.Txs)
	if err != nil {
		return err
	}

	// Begin new db tx
	tx, err := pub.indexer.db.Beginx()
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			shared.Rollback(tx)
			panic(p)
		} else if err != nil {
			shared.Rollback(tx)
		} else {
			err = tx.Commit()
		}
	}()

	// Publish trie nodes
	for _, node := range txTrieNodes {
		if err := shared.PublishIPLD(tx, node); err != nil {
			return err
		}
	}

	// Publish and index header
	if err := shared.PublishIPLD(tx, headerNode); err != nil {
		return err
	}
	header := HeaderModel{
		CID:         headerNode.Cid().String(),
		MhKey:       shared.MultihashKeyFromCID(headerNode.Cid()),
		ParentHash:  payload.Header.PrevBlock.String(),
		BlockNumber: strconv.Itoa(int(payload.BlockPayload.BlockHeight)),
		BlockHash:   payload.Header.BlockHash().String(),
		Timestamp:   payload.Header.Timestamp.UnixNano(),
		Bits:        payload.Header.Bits,
	}
	headerID, err := pub.indexer.indexHeaderCID(tx, header)
	if err != nil {
		return err
	}

	// Publish and index txs
	for i, txNode := range txNodes {
		if err := shared.PublishIPLD(tx, txNode); err != nil {
			return err
		}
		txModel := payload.TxMetaData[i]
		txModel.CID = txNode.Cid().String()
		txModel.MhKey = shared.MultihashKeyFromCID(txNode.Cid())
		txID, err := pub.indexer.indexTransactionCID(tx, txModel, headerID)
		if err != nil {
			return err
		}
		for _, input := range txModel.TxInputs {
			if err := pub.indexer.indexTxInput(tx, input, txID); err != nil {
				return err
			}
		}
		for _, output := range txModel.TxOutputs {
			if err := pub.indexer.indexTxOutput(tx, output, txID); err != nil {
				return err
			}
		}
	}

	return err
}

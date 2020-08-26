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
	"database/sql"
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/vulcanize/ipld-btc-indexer/pkg/postgres"
)

// Retriever interface fr substituting mocks in tests
type Retriever interface {
	RetrieveFirstBlockNumber() (int64, error)
	RetrieveLastBlockNumber() (int64, error)
	RetrieveGapsInData(validationLevel int) ([]DBGap, error)
}

// GapRetriever type for Bitcoin
type GapRetriever struct {
	db *postgres.DB
}

// NewGapRetriever returns a pointer to a new GapRetriever
func NewGapRetriever(db *postgres.DB) *GapRetriever {
	return &GapRetriever{
		db: db,
	}
}

// RetrieveFirstBlockNumber is used to retrieve the first block number in the db
func (bcr *GapRetriever) RetrieveFirstBlockNumber() (int64, error) {
	var blockNumber int64
	err := bcr.db.Get(&blockNumber, "SELECT block_number FROM btc.header_cids ORDER BY block_number ASC LIMIT 1")
	return blockNumber, err
}

// RetrieveLastBlockNumber is used to retrieve the latest block number in the db
func (bcr *GapRetriever) RetrieveLastBlockNumber() (int64, error) {
	var blockNumber int64
	err := bcr.db.Get(&blockNumber, "SELECT block_number FROM btc.header_cids ORDER BY block_number DESC LIMIT 1 ")
	return blockNumber, err
}

// DBGap type for querying for gaps in db
type DBGap struct {
	Start uint64 `db:"start"`
	Stop  uint64 `db:"stop"`
}

// RetrieveGapsInData is used to find the the block numbers at which we are missing data in the db
func (bcr *GapRetriever) RetrieveGapsInData(validationLevel int) ([]DBGap, error) {
	log.Info("searching for gaps in the btc ipfs watcher database")
	startingBlock, err := bcr.RetrieveFirstBlockNumber()
	if err != nil {
		return nil, fmt.Errorf("btc GapRetriever RetrieveFirstBlockNumber error: %v", err)
	}
	var initialGap []DBGap
	if startingBlock != 0 {
		stop := uint64(startingBlock - 1)
		log.Infof("found gap at the beginning of the btc sync from 0 to %d", stop)
		initialGap = []DBGap{{
			Start: 0,
			Stop:  stop,
		}}
	}

	pgStr := `SELECT header_cids.block_number + 1 AS start, min(fr.block_number) - 1 AS stop FROM btc.header_cids
				LEFT JOIN btc.header_cids r on btc.header_cids.block_number = r.block_number - 1
				LEFT JOIN btc.header_cids fr on btc.header_cids.block_number < fr.block_number
				WHERE r.block_number is NULL and fr.block_number IS NOT NULL
				GROUP BY header_cids.block_number, r.block_number`
	results := make([]struct {
		Start uint64 `db:"start"`
		Stop  uint64 `db:"stop"`
	}, 0)
	if err := bcr.db.Select(&results, pgStr); err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	emptyGaps := make([]DBGap, len(results))
	for i, res := range results {
		emptyGaps[i] = DBGap{
			Start: res.Start,
			Stop:  res.Stop,
		}
	}

	// Find sections of blocks where we are below the validation level
	// There will be no overlap between these "gaps" and the ones above
	pgStr = `SELECT block_number FROM btc.header_cids
			WHERE times_validated < $1
			ORDER BY block_number`
	var heights []uint64
	if err := bcr.db.Select(&heights, pgStr, validationLevel); err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	return append(append(initialGap, emptyGaps...), MissingHeightsToGaps(heights)...), nil
}

// MissingHeightsToGaps returns a slice of gaps from a slice of missing block heights
func MissingHeightsToGaps(heights []uint64) []DBGap {
	if len(heights) == 0 {
		return nil
	}
	validationGaps := make([]DBGap, 0)
	start := heights[0]
	lastHeight := start
	for i, height := range heights[1:] {
		if height != lastHeight+1 {
			validationGaps = append(validationGaps, DBGap{
				Start: start,
				Stop:  lastHeight,
			})
			start = height
		}
		if i+2 == len(heights) {
			validationGaps = append(validationGaps, DBGap{
				Start: start,
				Stop:  height,
			})
		}
		lastHeight = height
	}
	return validationGaps
}

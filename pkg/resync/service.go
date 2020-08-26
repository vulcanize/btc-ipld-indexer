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

package resync

import (
	"fmt"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/vulcanize/ipld-btc-indexer/pkg/btc"

	"github.com/sirupsen/logrus"

	"github.com/vulcanize/ipld-btc-indexer/pkg/shared"
	"github.com/vulcanize/ipld-btc-indexer/utils"
)

type Resync interface {
	Sync() error
}

type Service struct {
	// Interface for converting payloads into IPLD object payloads
	Converter btc.Converter
	// Interface for publishing the IPLD payloads to IPFS
	Publisher btc.Publisher
	// Interface for searching and retrieving CIDs from Postgres index
	Retriever btc.Retriever
	// Interface for fetching payloads over at historical blocks; over http
	Fetcher btc.Fetcher
	// Interface for cleaning out data before resyncing (if clearOldCache is on)
	Cleaner btc.Cleaner
	// Size of batch fetches
	BatchSize uint64
	// Number of worker goroutines
	Workers int64
	// Channel for receiving quit signal
	quitChan chan bool
	// Chain config for btc
	ChainConfig *chaincfg.Params
	// Resync data type
	data shared.DataType
	// Resync ranges
	ranges [][2]uint64
	// Flag to turn on or off old cache destruction
	clearOldCache bool
	// Flag to turn on or off validation level reset
	resetValidation bool
}

// NewResyncService creates and returns a resync service from the provided settings
func NewResyncService(settings *Config) (Resync, error) {
	rs := new(Service)
	var err error
	rs.ChainConfig = &chaincfg.MainNetParams /// TODO make this configurable
	rs.Converter = btc.NewPayloadConverter(rs.ChainConfig)
	rs.Publisher = btc.NewIPLDPublisher(settings.DB)
	rs.Retriever = btc.NewGapRetriever(settings.DB)
	rs.Fetcher, err = btc.NewPayloadFetcher(settings.HTTPConfig)
	if err != nil {
		return nil, err
	}
	rs.Cleaner = btc.NewDBCleaner(settings.DB)
	rs.BatchSize = settings.BatchSize
	if rs.BatchSize == 0 {
		rs.BatchSize = shared.DefaultMaxBatchSize
	}
	rs.Workers = int64(settings.Workers)
	if rs.Workers == 0 {
		rs.Workers = shared.DefaultMaxBatchNumber
	}
	rs.resetValidation = settings.ResetValidation
	rs.clearOldCache = settings.ClearOldCache
	rs.data = settings.ResyncType
	rs.ranges = settings.Ranges
	rs.quitChan = make(chan bool)
	return rs, nil
}

func (rs *Service) Sync() error {
	if rs.resetValidation {
		logrus.Infof("resetting validation level")
		if err := rs.Cleaner.ResetValidation(rs.ranges); err != nil {
			return fmt.Errorf("validation reset failed: %v", err)
		}
	}
	if rs.clearOldCache {
		logrus.Infof("cleaning out old data from Postgres")
		if err := rs.Cleaner.Clean(rs.ranges, rs.data); err != nil {
			return fmt.Errorf("bitcoin %s data resync cleaning error: %v", rs.data.String(), err)
		}
	}
	// spin up worker goroutines
	heightsChan := make(chan []uint64)
	for i := 1; i <= int(rs.Workers); i++ {
		go rs.resync(i, heightsChan)
	}
	for _, rng := range rs.ranges {
		if rng[1] < rng[0] {
			logrus.Error("bitcoin resync range ending block number needs to be greater than the starting block number")
			continue
		}
		logrus.Infof("resyncing bitcoin data from %d to %d", rng[0], rng[1])
		// break the range up into bins of smaller ranges
		blockRangeBins, err := utils.GetBlockHeightBins(rng[0], rng[1], rs.BatchSize)
		if err != nil {
			return err
		}
		for _, heights := range blockRangeBins {
			heightsChan <- heights
		}
	}
	// send a quit signal to each worker
	// this blocks until each worker has finished its current task and can receive from the quit channel
	for i := 1; i <= int(rs.Workers); i++ {
		rs.quitChan <- true
	}
	return nil
}

func (rs *Service) resync(id int, heightChan chan []uint64) {
	for {
		select {
		case heights := <-heightChan:
			logrus.Debugf("bitcoin resync worker %d processing section from %d to %d", id, heights[0], heights[len(heights)-1])
			payloads, err := rs.Fetcher.FetchAt(heights)
			if err != nil {
				logrus.Errorf("bitcoin resync worker %d fetcher error: %s", id, err.Error())
			}
			for _, payload := range payloads {
				ipldPayload, err := rs.Converter.Convert(payload)
				if err != nil {
					logrus.Errorf("bitcoin resync worker %d converter error: %s", id, err.Error())
				}
				if err := rs.Publisher.Publish(*ipldPayload); err != nil {
					logrus.Errorf("bitcoin resync worker %d publisher error: %s", id, err.Error())
				}
			}
			logrus.Infof("bitcoin resync worker %d finished section from %d to %d", id, heights[0], heights[len(heights)-1])
		case <-rs.quitChan:
			logrus.Infof("bitcoin resync worker %d goroutine shutting down", id)
			return
		}
	}
}

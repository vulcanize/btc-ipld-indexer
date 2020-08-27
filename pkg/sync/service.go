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

package sync

import (
	"sync"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/vulcanize/ipld-btc-indexer/pkg/btc"

	ethnode "github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/rpc"
	log "github.com/sirupsen/logrus"

	"github.com/vulcanize/ipld-btc-indexer/pkg/node"
	"github.com/vulcanize/ipld-btc-indexer/pkg/postgres"
	"github.com/vulcanize/ipld-btc-indexer/pkg/shared"
)

const (
	PayloadChanBufferSize = 2000
)

// Indexer is the top level interface for streaming, converting to IPLDs, publishing, and indexing all chain data at head
// This service is compatible with the Ethereum service interface (node.Service)
type Indexer interface {
	// APIs(), Protocols(), Start() and Stop()
	ethnode.Service
	// Data processing event loop
	Sync(wg *sync.WaitGroup) error
	// Method to access the node info for the service
	Node() *node.Node
	// Method to access chain type
	Chain() shared.ChainType
}

// Service is the underlying struct for the watcher
type Service struct {
	// Used to sync access to the Subscriptions
	sync.Mutex
	// Interface for streaming payloads over an rpc subscription
	Streamer btc.Streamer
	// Interface for converting raw payloads into IPLD object payloads
	Converter btc.Converter
	// Interface for publishing and indexing the PG-IPLD payloads
	Publisher btc.Publisher
	// Interface for searching and retrieving CIDs from Postgres index
	Retriever btc.Retriever
	// Chan the processor uses to subscribe to payloads from the Streamer
	PayloadChan chan btc.BlockPayload
	// Used to signal shutdown of the service
	QuitChan chan bool
	// Info for the Geth node that this watcher is working with
	NodeInfo *node.Node
	// Number of worker goroutines
	Workers int64
	// chain type for this service
	ChainConfig *chaincfg.Params
	// Underlying db
	db *postgres.DB
	// wg for syncing serve processes
	serveWg *sync.WaitGroup
}

// NewIndexerService creates a new Indexer using an underlying Service struct
func NewIndexerService(settings *Config) (Indexer, error) {
	sn := new(Service)
	sn.Streamer = btc.NewHTTPPayloadStreamer(settings.ClientConfig)
	sn.ChainConfig = &chaincfg.Params{} /// TODO make this configurable
	sn.Converter = btc.NewPayloadConverter(sn.ChainConfig)
	sn.Publisher = btc.NewIPLDPublisher(settings.DB)
	sn.Retriever = btc.NewGapRetriever(settings.DB)

	sn.QuitChan = make(chan bool)
	sn.Workers = settings.Workers
	sn.NodeInfo = &settings.NodeInfo
	return sn, nil
}

// Protocols exports the services p2p protocols, this service has none
func (sap *Service) Protocols() []p2p.Protocol {
	return []p2p.Protocol{}
}

// APIs returns the RPC descriptors the watcher service offers
func (sap *Service) APIs() []rpc.API {
	return []rpc.API{}
}

// Sync streams incoming raw chain data and converts it for further processing
// It forwards the converted data to the publish process(es) it spins up
// This continues on no matter if or how many subscribers there are
func (sap *Service) Sync(wg *sync.WaitGroup) error {
	sub, err := sap.Streamer.Stream(sap.PayloadChan)
	if err != nil {
		return err
	}
	// spin up publish worker goroutines
	publishPayload := make(chan btc.ConvertedPayload, PayloadChanBufferSize)
	for i := 1; i <= int(sap.Workers); i++ {
		go sap.publish(wg, i, publishPayload)
		log.Debugf("bitcoin sync worker %d successfully spun up", i)
	}
	go func() {
		wg.Add(1)
		defer wg.Done()
		for {
			select {
			case payload := <-sap.PayloadChan:
				ipldPayload, err := sap.Converter.Convert(payload)
				if err != nil {
					log.Errorf("bitcoin data conversion error: %v", err)
					continue
				}
				log.Infof("bitcoin data streamed at head height %d", ipldPayload.Height())
				// Forward the payload to the publish workers
				// this channel acts as a ring buffer
				select {
				case publishPayload <- *ipldPayload:
				default:
					<-publishPayload
					publishPayload <- *ipldPayload
				}
			case err := <-sub.Err():
				log.Errorf("bitcoin subscription error for chain: %v", err)
			case <-sap.QuitChan:
				log.Info("quiting bitcoin sync process")
				return
			}
		}
	}()
	log.Info("bitcoin sync process successfully spun up")
	return nil
}

// publish is spun up by SyncAndConvert and receives converted chain data from that process
// it publishes this data to IPFS and indexes their CIDs with useful metadata in Postgres
func (sap *Service) publish(wg *sync.WaitGroup, id int, publishPayload <-chan btc.ConvertedPayload) {
	wg.Add(1)
	defer wg.Done()
	for {
		select {
		case payload := <-publishPayload:
			log.Debugf("bitcoin sync worker %d publishing and indexing data streamed at head height %d", id, payload.Height())
			if err := sap.Publisher.Publish(payload); err != nil {
				log.Errorf("bitcoin sync worker %d publishing error: %v", id, err)
				continue
			}
		case <-sap.QuitChan:
			log.Infof("bitcoin sync worker %d shutting down", id)
			return
		}
	}
}

// Start is used to begin the service
// This is mostly just to satisfy the node.Service interface
func (sap *Service) Start(*p2p.Server) error {
	log.Info("starting bitcoin sync service")
	wg := new(sync.WaitGroup)
	return sap.Sync(wg)
}

// Stop is used to close down the service
// This is mostly just to satisfy the node.Service interface
func (sap *Service) Stop() error {
	log.Info("stopping bitcoin sync service")
	sap.Lock()
	close(sap.QuitChan)
	sap.Unlock()
	return nil
}

// Node returns the node info for this service
func (sap *Service) Node() *node.Node {
	return sap.NodeInfo
}

// Chain returns the chain type for this service
func (sap *Service) Chain() shared.ChainType {
	return shared.Bitcoin
}

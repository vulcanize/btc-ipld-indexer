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
	"bytes"
	"time"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/sirupsen/logrus"
)

// Streamer interface for substituting mocks in tests
type Streamer interface {
	Stream(payloadChan chan BlockPayload) (*HTTPClientSubscription, error)
}

// HTTPPayloadStreamer satisfies the PayloadStreamer interface for bitcoin over http endpoints
// (bitcoin core doesn't support websockets, btcd doesn't support zmq- need to write adapter)
type HTTPPayloadStreamer struct {
	Config   *rpcclient.ConnConfig
	lastHash []byte
}

// NewHTTPPayloadStreamer creates a pointer to a new PayloadStreamer which satisfies the PayloadStreamer interface for bitcoin
func NewHTTPPayloadStreamer(clientConfig *rpcclient.ConnConfig) *HTTPPayloadStreamer {
	return &HTTPPayloadStreamer{
		Config: clientConfig,
	}
}

// Stream is the main loop for subscribing to data from the btc block notifications
// using only the standard http endpoints shared between bitcoind and btcd nodes
func (ps *HTTPPayloadStreamer) Stream(payloadChan chan BlockPayload) (*HTTPClientSubscription, error) {
	logrus.Debug("streaming block payloads from btc")
	client, err := rpcclient.New(ps.Config, nil)
	if err != nil {
		return nil, err
	}
	ticker := time.NewTicker(time.Second * 5)
	errChan := make(chan error)
	go func() {
		for {
			// start at
			select {
			case <-ticker.C:
				height, err := client.GetBlockCount()
				if err != nil {
					errChan <- err
					continue
				}
				blockHash, err := client.GetBlockHash(height)
				if err != nil {
					errChan <- err
					continue
				}
				blockHashBytes := blockHash.CloneBytes()
				if bytes.Equal(blockHashBytes, ps.lastHash) {
					continue
				}
				block, err := client.GetBlock(blockHash)
				if err != nil {
					errChan <- err
					continue
				}
				ps.lastHash = blockHashBytes
				payloadChan <- BlockPayload{
					Header:      &block.Header,
					BlockHeight: height,
					Txs:         msgTxsToUtilTxs(block.Transactions),
				}
			default:
			}
		}
	}()
	return &HTTPClientSubscription{client: client, errChan: errChan}, nil
}

// HTTPClientSubscription is a wrapper around the underlying bitcoind rpc client
// TODO: use ZMQ from bitcoind or use websockets from btcd
type HTTPClientSubscription struct {
	client  *rpcclient.Client
	errChan chan error
}

// Unsubscribe satisfies the rpc.Subscription interface
func (bcs *HTTPClientSubscription) Unsubscribe() {
	bcs.client.Shutdown()
}

// Err() satisfies the rpc.Subscription interface
func (bcs *HTTPClientSubscription) Err() <-chan error {
	return bcs.errChan
}

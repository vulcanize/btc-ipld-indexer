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

package shared

import (
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/spf13/viper"
	"github.com/vulcanize/ipld-btc-indexer/pkg/node"
)

// Env variables
const (
	HTTP_TIMEOUT = "HTTP_TIMEOUT"

	BTC_WS_PATH       = "BTC_WS_PATH"
	BTC_HTTP_PATH     = "BTC_HTTP_PATH"
	BTC_NODE_PASSWORD = "BTC_NODE_PASSWORD"
	BTC_NODE_USER     = "BTC_NODE_USER"
	BTC_NODE_ID       = "BTC_NODE_ID"
	BTC_CLIENT_NAME   = "BTC_CLIENT_NAME"
	BTC_GENESIS_BLOCK = "BTC_GENESIS_BLOCK"
	BTC_NETWORK_ID    = "BTC_NETWORK_ID"
	BTC_CHAIN_ID      = "BTC_CHAIN_ID"
)

// GetBtcNodeAndClient returns btc node info from path url
func GetBtcNodeAndClient(path string) (node.Node, *rpcclient.ConnConfig) {
	viper.BindEnv("bitcoin.nodeID", BTC_NODE_ID)
	viper.BindEnv("bitcoin.clientName", BTC_CLIENT_NAME)
	viper.BindEnv("bitcoin.genesisBlock", BTC_GENESIS_BLOCK)
	viper.BindEnv("bitcoin.networkID", BTC_NETWORK_ID)
	viper.BindEnv("bitcoin.pass", BTC_NODE_PASSWORD)
	viper.BindEnv("bitcoin.user", BTC_NODE_USER)
	viper.BindEnv("bitcoin.chainID", BTC_CHAIN_ID)

	// For bitcoin we load in node info from the config because there is no RPC endpoint to retrieve this from the node
	return node.Node{
			ID:           viper.GetString("bitcoin.nodeID"),
			ClientName:   viper.GetString("bitcoin.clientName"),
			GenesisBlock: viper.GetString("bitcoin.genesisBlock"),
			NetworkID:    viper.GetString("bitcoin.networkID"),
			ChainID:      viper.GetUint64("bitcoin.chainID"),
		}, &rpcclient.ConnConfig{
			Host:         path,
			HTTPPostMode: true, // Bitcoin core only supports HTTP POST mode
			DisableTLS:   true, // Bitcoin core does not provide TLS by default
			Pass:         viper.GetString("bitcoin.pass"),
			User:         viper.GetString("bitcoin.user"),
		}
}

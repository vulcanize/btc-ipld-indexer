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
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/spf13/viper"

	"github.com/vulcanize/ipld-btc-indexer/pkg/node"
	"github.com/vulcanize/ipld-btc-indexer/pkg/postgres"
	"github.com/vulcanize/ipld-btc-indexer/pkg/shared"
	"github.com/vulcanize/ipld-btc-indexer/utils"
)

// Env variables
const (
	SUPERNODE_WORKERS = "SYNC_WORKERS"

	SYNC_MAX_IDLE_CONNECTIONS = "SYNC_MAX_IDLE_CONNECTIONS"
	SYNC_MAX_OPEN_CONNECTIONS = "SYNC_MAX_OPEN_CONNECTIONS"
	SYNC_MAX_CONN_LIFETIME    = "SYNC_MAX_CONN_LIFETIME"
)

// Config struct
type Config struct {
	DB           *postgres.DB
	DBConfig     postgres.Config
	Workers      int64
	ClientConfig *rpcclient.ConnConfig
	NodeInfo     node.Node
}

// NewConfig is used to initialize a sync config from a .toml file
func NewConfig() *Config {
	c := new(Config)

	viper.BindEnv("sync.workers", SUPERNODE_WORKERS)
	viper.BindEnv("bitcoin.wsPath", shared.BTC_WS_PATH)

	workers := viper.GetInt64("sync.workers")
	if workers < 1 {
		workers = 1
	}
	c.Workers = workers

	btcWS := viper.GetString("bitcoin.wsPath")
	c.NodeInfo, c.ClientConfig = shared.GetBtcNodeAndClient(btcWS)

	c.DBConfig.Init()
	overrideDBConnConfig(&c.DBConfig)
	syncDB := utils.LoadPostgres(c.DBConfig, c.NodeInfo)
	c.DB = &syncDB
	return c
}

func overrideDBConnConfig(con *postgres.Config) {
	viper.BindEnv("database.sync.maxIdle", SYNC_MAX_IDLE_CONNECTIONS)
	viper.BindEnv("database.sync.maxOpen", SYNC_MAX_OPEN_CONNECTIONS)
	viper.BindEnv("database.sync.maxLifetime", SYNC_MAX_CONN_LIFETIME)
	con.MaxIdle = viper.GetInt("database.sync.maxIdle")
	con.MaxOpen = viper.GetInt("database.sync.maxOpen")
	con.MaxLifetime = viper.GetInt("database.sync.maxLifetime")
}

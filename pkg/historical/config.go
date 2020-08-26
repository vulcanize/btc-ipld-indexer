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

package historical

import (
	"time"

	"github.com/btcsuite/btcd/rpcclient"

	"github.com/spf13/viper"

	"github.com/vulcanize/ipld-btc-indexer/pkg/node"
	"github.com/vulcanize/ipld-btc-indexer/pkg/postgres"
	"github.com/vulcanize/ipld-btc-indexer/pkg/shared"
	"github.com/vulcanize/ipld-btc-indexer/utils"
)

// Env variables
const (
	BACKFILL_FREQUENCY        = "BACKFILL_FREQUENCY"
	BACKFILL_BATCH_SIZE       = "BACKFILL_BATCH_SIZE"
	BACKFILL_WORKERS          = "BACKFILL_WORKERS"
	BACKFILL_VALIDATION_LEVEL = "BACKFILL_VALIDATION_LEVEL"

	BACKFILL_MAX_IDLE_CONNECTIONS = "BACKFILL_MAX_IDLE_CONNECTIONS"
	BACKFILL_MAX_OPEN_CONNECTIONS = "BACKFILL_MAX_OPEN_CONNECTIONS"
	BACKFILL_MAX_CONN_LIFETIME    = "BACKFILL_MAX_CONN_LIFETIME"
)

// Config struct
type Config struct {
	DB              *postgres.DB
	DBConfig        postgres.Config
	HTTPConfig      *rpcclient.ConnConfig
	Frequency       time.Duration
	BatchSize       uint64
	Workers         uint64
	ValidationLevel int
	Timeout         time.Duration // HTTP connection timeout in seconds
	NodeInfo        node.Node
}

// NewConfig is used to initialize a historical config from a .toml file
func NewConfig() *Config {
	c := new(Config)

	viper.BindEnv("bitcoin.httpPath", shared.BTC_HTTP_PATH)
	viper.BindEnv("backfill.frequency", BACKFILL_FREQUENCY)
	viper.BindEnv("backfill.batchSize", BACKFILL_BATCH_SIZE)
	viper.BindEnv("backfill.workers", BACKFILL_WORKERS)
	viper.BindEnv("backfill.validationLevel", BACKFILL_VALIDATION_LEVEL)
	viper.BindEnv("backfill.timeout", shared.HTTP_TIMEOUT)

	timeout := viper.GetInt("backfill.timeout")
	if timeout < 15 {
		timeout = 15
	}
	c.Timeout = time.Second * time.Duration(timeout)

	freq := viper.GetInt("backfill.frequency")
	var frequency time.Duration
	if freq <= 0 {
		frequency = time.Second * 30
	} else {
		frequency = time.Second * time.Duration(freq)
	}
	c.Frequency = frequency
	c.BatchSize = uint64(viper.GetInt64("backfill.batchSize"))
	c.Workers = uint64(viper.GetInt64("backfill.workers"))
	c.ValidationLevel = viper.GetInt("backfill.validationLevel")

	btcHTTP := viper.GetString("bitcoin.httpPath")
	c.NodeInfo, c.HTTPConfig = shared.GetBtcNodeAndClient(btcHTTP)

	c.DBConfig.Init()
	overrideDBConnConfig(&c.DBConfig)
	db := utils.LoadPostgres(c.DBConfig, c.NodeInfo)
	c.DB = &db
	return c
}

func overrideDBConnConfig(con *postgres.Config) {
	viper.BindEnv("database.backfill.maxIdle", BACKFILL_MAX_IDLE_CONNECTIONS)
	viper.BindEnv("database.backfill.maxOpen", BACKFILL_MAX_OPEN_CONNECTIONS)
	viper.BindEnv("database.backfill.maxLifetime", BACKFILL_MAX_CONN_LIFETIME)
	con.MaxIdle = viper.GetInt("database.backfill.maxIdle")
	con.MaxOpen = viper.GetInt("database.backfill.maxOpen")
	con.MaxLifetime = viper.GetInt("database.backfill.maxLifetime")
}

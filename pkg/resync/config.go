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
	RESYNC_START            = "RESYNC_START"
	RESYNC_STOP             = "RESYNC_STOP"
	RESYNC_BATCH_SIZE       = "RESYNC_BATCH_SIZE"
	RESYNC_WORKERS          = "RESYNC_WORKERS"
	RESYNC_CLEAR_OLD_CACHE  = "RESYNC_CLEAR_OLD_CACHE"
	RESYNC_TYPE             = "RESYNC_TYPE"
	RESYNC_RESET_VALIDATION = "RESYNC_RESET_VALIDATION"

	RESYNC_MAX_IDLE_CONNECTIONS = "RESYNC_MAX_IDLE_CONNECTIONS"
	RESYNC_MAX_OPEN_CONNECTIONS = "RESYNC_MAX_OPEN_CONNECTIONS"
	RESYNC_MAX_CONN_LIFETIME    = "RESYNC_MAX_CONN_LIFETIME"
)

// Config holds the parameters needed to perform a resync
type Config struct {
	ResyncType      shared.DataType // The type of data to resync
	ClearOldCache   bool            // Resync will first clear all the data within the range
	ResetValidation bool            // If true, resync will reset the validation level to 0 for the given range

	// DB info
	DB       *postgres.DB
	DBConfig postgres.Config

	HTTPConfig *rpcclient.ConnConfig // Bitcoin rpc client config
	NodeInfo   node.Node             // Info for the associated node
	Ranges     [][2]uint64           // The block height ranges to resync
	BatchSize  uint64                // BatchSize for the resync http calls (client has to support batch sizing)
	Timeout    time.Duration         // HTTP connection timeout in seconds
	Workers    uint64
}

// NewConfig fills and returns a resync config from toml parameters
func NewConfig() (*Config, error) {
	c := new(Config)
	var err error

	viper.BindEnv("bitcoin.httpPath", shared.BTC_HTTP_PATH)
	viper.BindEnv("resync.start", RESYNC_START)
	viper.BindEnv("resync.stop", RESYNC_STOP)
	viper.BindEnv("resync.clearOldCache", RESYNC_CLEAR_OLD_CACHE)
	viper.BindEnv("resync.type", RESYNC_TYPE)
	viper.BindEnv("resync.batchSize", RESYNC_BATCH_SIZE)
	viper.BindEnv("resync.workers", RESYNC_WORKERS)
	viper.BindEnv("resync.resetValidation", RESYNC_RESET_VALIDATION)
	viper.BindEnv("resync.timeout", shared.HTTP_TIMEOUT)

	timeout := viper.GetInt("resync.timeout")
	if timeout < 5 {
		timeout = 5
	}
	c.Timeout = time.Second * time.Duration(timeout)

	start := uint64(viper.GetInt64("resync.start"))
	stop := uint64(viper.GetInt64("resync.stop"))
	c.Ranges = [][2]uint64{{start, stop}}
	c.ClearOldCache = viper.GetBool("resync.clearOldCache")
	c.ResetValidation = viper.GetBool("resync.resetValidation")
	c.BatchSize = uint64(viper.GetInt64("resync.batchSize"))
	c.Workers = uint64(viper.GetInt64("resync.workers"))

	resyncType := viper.GetString("resync.type")
	c.ResyncType, err = shared.GenerateDataTypeFromString(resyncType)
	if err != nil {
		return nil, err
	}
	if ok, err := shared.SupportedDataType(c.ResyncType, shared.Bitcoin); !ok {
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("bitcoin does not support data type %s", c.ResyncType.String())
	}

	btcHTTP := viper.GetString("bitcoin.httpPath")
	c.NodeInfo, c.HTTPConfig = shared.GetBtcNodeAndClient(btcHTTP)

	c.DBConfig.Init()
	overrideDBConnConfig(&c.DBConfig)
	db := utils.LoadPostgres(c.DBConfig, c.NodeInfo)
	c.DB = &db
	return c, nil
}

func overrideDBConnConfig(con *postgres.Config) {
	viper.BindEnv("database.resync.maxIdle", RESYNC_MAX_IDLE_CONNECTIONS)
	viper.BindEnv("database.resync.maxOpen", RESYNC_MAX_OPEN_CONNECTIONS)
	viper.BindEnv("database.resync.maxLifetime", RESYNC_MAX_CONN_LIFETIME)
	con.MaxIdle = viper.GetInt("database.resync.maxIdle")
	con.MaxOpen = viper.GetInt("database.resync.maxOpen")
	con.MaxLifetime = viper.GetInt("database.resync.maxLifetime")
}

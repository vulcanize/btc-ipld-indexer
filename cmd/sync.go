// Copyright © 2020 Vulcanize, Inc
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package cmd

import (
	"os"
	"os/signal"
	s "sync"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	w "github.com/vulcanize/ipld-btc-indexer/pkg/sync"
	v "github.com/vulcanize/ipld-btc-indexer/version"
)

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "sync bitcoin chain data into PG-IPFS",
	Long: `This command syncs all Bitcoin data from the head of the chain, processing this data into IPLD objects and
publishing and indexing them in PG-IPFS.
This command tracks the head of the chain, for filling in historical data see the backfill and resync commands

NOTE: Requires a btc full node'
`,
	Run: func(cmd *cobra.Command, args []string) {
		subCommand = cmd.CalledAs()
		logWithCommand = *log.WithField("SubCommand", subCommand)
		sync()
	},
}

func sync() {
	logWithCommand.Infof("running ipld-btc-indexer version: %s", v.VersionWithMeta)

	wg := new(s.WaitGroup)
	logWithCommand.Debug("loading sync configuration variables")
	syncerConfig := w.NewConfig()
	logWithCommand.Infof("config: %+v", syncerConfig)
	logWithCommand.Debug("initializing new sync service")
	syncer, err := w.NewIndexerService(syncerConfig)
	if err != nil {
		logWithCommand.Fatal(err)
	}

	logWithCommand.Info("starting up sync process")
	if err := syncer.Sync(wg); err != nil {
		logWithCommand.Fatal(err)
	}

	shutdown := make(chan os.Signal)
	signal.Notify(shutdown, os.Interrupt)
	<-shutdown
	syncer.Stop()
	wg.Wait()
}

func init() {
	rootCmd.AddCommand(syncCmd)

	// flags
	syncCmd.PersistentFlags().Int("sync-workers", 0, "how many worker goroutines to publish and index data")
	syncCmd.PersistentFlags().String("btc-ws-path", "", "ws url for bitcoin node")

	// and their .toml config bindings
	viper.BindPFlag("sync.workers", syncCmd.PersistentFlags().Lookup("sync-workers"))
	viper.BindPFlag("bitcoin.wsPath", syncCmd.PersistentFlags().Lookup("btc-ws-path"))
}

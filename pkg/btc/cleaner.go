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
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"

	"github.com/vulcanize/ipld-btc-indexer/pkg/postgres"
	"github.com/vulcanize/ipld-btc-indexer/pkg/shared"
)

// Cleaner interface for substituting mocks in tests
type Cleaner interface {
	ResetValidation(rngs [][2]uint64) error
	Clean(rngs [][2]uint64, t shared.DataType) error
}

// DBCleaner satisfies the Cleaner interface fo bitcoin
type DBCleaner struct {
	db *postgres.DB
}

// NewDBCleaner returns a new DBCleaner struct
func NewDBCleaner(db *postgres.DB) *DBCleaner {
	return &DBCleaner{
		db: db,
	}
}

// ResetValidation resets the validation level to 0 to enable revalidation
func (c *DBCleaner) ResetValidation(rngs [][2]uint64) error {
	tx, err := c.db.Beginx()
	if err != nil {
		return err
	}
	for _, rng := range rngs {
		logrus.Infof("btc db cleaner resetting validation level to 0 for block range %d to %d", rng[0], rng[1])
		pgStr := `UPDATE btc.header_cids
				SET times_validated = 0
				WHERE block_number BETWEEN $1 AND $2`
		if _, err := tx.Exec(pgStr, rng[0], rng[1]); err != nil {
			shared.Rollback(tx)
			return err
		}
	}
	return tx.Commit()
}

// Clean removes the specified data from the db within the provided block range
func (c *DBCleaner) Clean(rngs [][2]uint64, t shared.DataType) error {
	tx, err := c.db.Beginx()
	if err != nil {
		return err
	}
	for _, rng := range rngs {
		logrus.Infof("btc db cleaner cleaning up block range %d to %d", rng[0], rng[1])
		if err := c.clean(tx, rng, t); err != nil {
			shared.Rollback(tx)
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	logrus.Infof("btc db cleaner vacuum analyzing cleaned tables to free up space from deleted rows")
	return c.vacuumAnalyze(t)
}

func (c *DBCleaner) clean(tx *sqlx.Tx, rng [2]uint64, t shared.DataType) error {
	switch t {
	case shared.Full, shared.Headers:
		return c.cleanFull(tx, rng)
	case shared.Transactions:
		if err := c.cleanTransactionIPLDs(tx, rng); err != nil {
			return err
		}
		return c.cleanTransactionMetaData(tx, rng)
	default:
		return fmt.Errorf("btc cleaner unrecognized type: %s", t.String())
	}
}

func (c *DBCleaner) vacuumAnalyze(t shared.DataType) error {
	switch t {
	case shared.Full, shared.Headers:
		if err := c.vacuumHeaders(); err != nil {
			return err
		}
		if err := c.vacuumTxs(); err != nil {
			return err
		}
		if err := c.vacuumTxInputs(); err != nil {
			return err
		}
		if err := c.vacuumTxOutputs(); err != nil {
			return err
		}
	case shared.Transactions:
		if err := c.vacuumTxs(); err != nil {
			return err
		}
		if err := c.vacuumTxInputs(); err != nil {
			return err
		}
		if err := c.vacuumTxOutputs(); err != nil {
			return err
		}
	default:
		return fmt.Errorf("btc cleaner unrecognized type: %s", t.String())
	}
	return c.vacuumIPLDs()
}

func (c *DBCleaner) vacuumHeaders() error {
	_, err := c.db.Exec(`VACUUM ANALYZE btc.header_cids`)
	return err
}

func (c *DBCleaner) vacuumTxs() error {
	_, err := c.db.Exec(`VACUUM ANALYZE btc.transaction_cids`)
	return err
}

func (c *DBCleaner) vacuumTxInputs() error {
	_, err := c.db.Exec(`VACUUM ANALYZE btc.tx_inputs`)
	return err
}

func (c *DBCleaner) vacuumTxOutputs() error {
	_, err := c.db.Exec(`VACUUM ANALYZE btc.tx_outputs`)
	return err
}

func (c *DBCleaner) vacuumIPLDs() error {
	_, err := c.db.Exec(`VACUUM ANALYZE public.blocks`)
	return err
}

func (c *DBCleaner) cleanFull(tx *sqlx.Tx, rng [2]uint64) error {
	if err := c.cleanTransactionIPLDs(tx, rng); err != nil {
		return err
	}
	if err := c.cleanHeaderIPLDs(tx, rng); err != nil {
		return err
	}
	return c.cleanHeaderMetaData(tx, rng)
}

func (c *DBCleaner) cleanTransactionIPLDs(tx *sqlx.Tx, rng [2]uint64) error {
	pgStr := `DELETE FROM public.blocks A
			USING btc.transaction_cids B, btc.header_cids C
			WHERE A.key = B.mh_key
			AND B.header_id = C.id
			AND C.block_number BETWEEN $1 AND $2`
	_, err := tx.Exec(pgStr, rng[0], rng[1])
	return err
}

func (c *DBCleaner) cleanTransactionMetaData(tx *sqlx.Tx, rng [2]uint64) error {
	pgStr := `DELETE FROM btc.transaction_cids A
			USING btc.header_cids B
			WHERE A.header_id = B.id
			AND B.block_number BETWEEN $1 AND $2`
	_, err := tx.Exec(pgStr, rng[0], rng[1])
	return err
}

func (c *DBCleaner) cleanHeaderIPLDs(tx *sqlx.Tx, rng [2]uint64) error {
	pgStr := `DELETE FROM public.blocks A
			USING btc.header_cids B
			WHERE A.key = B.mh_key
			AND B.block_number BETWEEN $1 AND $2`
	_, err := tx.Exec(pgStr, rng[0], rng[1])
	return err
}

func (c *DBCleaner) cleanHeaderMetaData(tx *sqlx.Tx, rng [2]uint64) error {
	pgStr := `DELETE FROM btc.header_cids
			WHERE block_number BETWEEN $1 AND $2`
	_, err := tx.Exec(pgStr, rng[0], rng[1])
	return err
}

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

package btc_test

import (
	"math/big"

	"github.com/ethereum/go-ethereum/crypto"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/vulcanize/ipld-btc-indexer/pkg/btc"
	"github.com/vulcanize/ipld-btc-indexer/pkg/postgres"
	"github.com/vulcanize/ipld-btc-indexer/pkg/shared"
)

var (
	// Block 0
	// header variables
	blockHash1   = crypto.Keccak256Hash([]byte{00, 02})
	blocKNumber1 = big.NewInt(0)
	headerCid1   = shared.TestCID([]byte("mockHeader1CID"))
	headerMhKey1 = shared.MultihashKeyFromCID(headerCid1)
	parentHash   = crypto.Keccak256Hash([]byte{00, 01})
	headerModel1 = btc.HeaderModel{
		BlockHash:   blockHash1.String(),
		BlockNumber: blocKNumber1.String(),
		ParentHash:  parentHash.String(),
		CID:         headerCid1.String(),
		MhKey:       headerMhKey1,
	}

	// tx variables
	tx1CID    = shared.TestCID([]byte("mockTx1CID"))
	tx1MhKey  = shared.MultihashKeyFromCID(tx1CID)
	tx2CID    = shared.TestCID([]byte("mockTx2CID"))
	tx2MhKey  = shared.MultihashKeyFromCID(tx2CID)
	tx1Hash   = crypto.Keccak256Hash([]byte{01, 01})
	tx2Hash   = crypto.Keccak256Hash([]byte{01, 02})
	opHash    = crypto.Keccak256Hash([]byte{02, 01})
	txModels1 = []btc.TxModelWithInsAndOuts{
		{
			Index:  0,
			CID:    tx1CID.String(),
			MhKey:  tx1MhKey,
			TxHash: tx1Hash.String(),
			SegWit: true,
			TxInputs: []btc.TxInput{
				{
					Index:                 0,
					TxWitness:             []string{"mockWitness"},
					SignatureScript:       []byte{01},
					PreviousOutPointIndex: 0,
					PreviousOutPointHash:  opHash.String(),
				},
			},
			TxOutputs: []btc.TxOutput{
				{
					Index:        0,
					Value:        50000000,
					PkScript:     []byte{02},
					ScriptClass:  0,
					RequiredSigs: 1,
				},
			},
		},
		{
			Index:  1,
			CID:    tx2CID.String(),
			MhKey:  tx2MhKey,
			TxHash: tx2Hash.String(),
			SegWit: true,
		},
	}
	mockCIDPayload1 = &btc.CIDPayload{
		HeaderCID:       headerModel1,
		TransactionCIDs: txModels1,
	}

	// Block 1
	// header variables
	blockHash2   = crypto.Keccak256Hash([]byte{00, 03})
	blocKNumber2 = big.NewInt(1)
	headerCid2   = shared.TestCID([]byte("mockHeaderCID2"))
	headerMhKey2 = shared.MultihashKeyFromCID(headerCid2)
	headerModel2 = btc.HeaderModel{
		BlockNumber: blocKNumber2.String(),
		BlockHash:   blockHash2.String(),
		ParentHash:  blockHash1.String(),
		CID:         headerCid2.String(),
		MhKey:       headerMhKey2,
	}

	// tx variables
	tx3CID    = shared.TestCID([]byte("mockTx3CID"))
	tx3MhKey  = shared.MultihashKeyFromCID(tx3CID)
	tx3Hash   = crypto.Keccak256Hash([]byte{01, 03})
	txModels2 = []btc.TxModelWithInsAndOuts{
		{
			Index:  0,
			CID:    tx3CID.String(),
			MhKey:  tx3MhKey,
			TxHash: tx3Hash.String(),
			SegWit: true,
		},
	}
	mockCIDPayload2 = &btc.CIDPayload{
		HeaderCID:       headerModel2,
		TransactionCIDs: txModels2,
	}
	rngs   = [][2]uint64{{0, 1}}
	mhKeys = []string{
		headerMhKey1,
		headerMhKey2,
		tx1MhKey,
		tx2MhKey,
		tx3MhKey,
	}
	mockData = []byte{'\x01'}
)

var _ = Describe("Cleaner", func() {
	var (
		db      *postgres.DB
		repo    *btc.CIDIndexer
		cleaner *btc.DBCleaner
	)
	BeforeEach(func() {
		var err error
		db, err = shared.SetupDB()
		Expect(err).ToNot(HaveOccurred())
		repo = btc.NewCIDIndexer(db)
		cleaner = btc.NewDBCleaner(db)
	})

	Describe("Clean", func() {
		BeforeEach(func() {
			for _, key := range mhKeys {
				_, err := db.Exec(`INSERT INTO public.blocks (key, data) VALUES ($1, $2)`, key, mockData)
				Expect(err).ToNot(HaveOccurred())
			}
			err := repo.Index(mockCIDPayload1)
			Expect(err).ToNot(HaveOccurred())
			err = repo.Index(mockCIDPayload2)
			Expect(err).ToNot(HaveOccurred())

			tx, err := db.Beginx()
			Expect(err).ToNot(HaveOccurred())
			var startingIPFSBlocksCount int
			pgStr := `SELECT COUNT(*) FROM public.blocks`
			err = tx.Get(&startingIPFSBlocksCount, pgStr)
			Expect(err).ToNot(HaveOccurred())
			var startingTxCount int
			pgStr = `SELECT COUNT(*) FROM btc.transaction_cids`
			err = tx.Get(&startingTxCount, pgStr)
			Expect(err).ToNot(HaveOccurred())
			var startingHeaderCount int
			pgStr = `SELECT COUNT(*) FROM btc.header_cids`
			err = tx.Get(&startingHeaderCount, pgStr)
			Expect(err).ToNot(HaveOccurred())

			err = tx.Commit()
			Expect(err).ToNot(HaveOccurred())

			Expect(startingIPFSBlocksCount).To(Equal(5))
			Expect(startingTxCount).To(Equal(3))
			Expect(startingHeaderCount).To(Equal(2))
		})
		AfterEach(func() {
			btc.TearDownDB(db)
		})
		It("Cleans everything", func() {
			err := cleaner.Clean(rngs, shared.Full)
			Expect(err).ToNot(HaveOccurred())

			tx, err := db.Beginx()
			Expect(err).ToNot(HaveOccurred())
			var txCount int
			pgStr := `SELECT COUNT(*) FROM btc.transaction_cids`
			err = tx.Get(&txCount, pgStr)
			Expect(err).ToNot(HaveOccurred())
			var txInCount int
			pgStr = `SELECT COUNT(*) FROM btc.tx_inputs`
			err = tx.Get(&txInCount, pgStr)
			Expect(err).ToNot(HaveOccurred())
			var txOutCount int
			pgStr = `SELECT COUNT(*) FROM btc.tx_outputs`
			err = tx.Get(&txOutCount, pgStr)
			Expect(err).ToNot(HaveOccurred())
			var headerCount int
			pgStr = `SELECT COUNT(*) FROM btc.header_cids`
			err = tx.Get(&headerCount, pgStr)
			Expect(err).ToNot(HaveOccurred())
			var blocksCount int
			pgStr = `SELECT COUNT(*) FROM public.blocks`
			err = tx.Get(&blocksCount, pgStr)
			Expect(err).ToNot(HaveOccurred())

			err = tx.Commit()
			Expect(err).ToNot(HaveOccurred())

			Expect(blocksCount).To(Equal(0))
			Expect(txCount).To(Equal(0))
			Expect(txInCount).To(Equal(0))
			Expect(txOutCount).To(Equal(0))
			Expect(headerCount).To(Equal(0))
		})
		It("Cleans headers and all linked data", func() {
			err := cleaner.Clean(rngs, shared.Headers)
			Expect(err).ToNot(HaveOccurred())

			tx, err := db.Beginx()
			Expect(err).ToNot(HaveOccurred())
			var txCount int
			pgStr := `SELECT COUNT(*) FROM btc.transaction_cids`
			err = tx.Get(&txCount, pgStr)
			Expect(err).ToNot(HaveOccurred())
			var txInCount int
			pgStr = `SELECT COUNT(*) FROM btc.tx_inputs`
			err = tx.Get(&txInCount, pgStr)
			Expect(err).ToNot(HaveOccurred())
			var txOutCount int
			pgStr = `SELECT COUNT(*) FROM btc.tx_outputs`
			err = tx.Get(&txOutCount, pgStr)
			Expect(err).ToNot(HaveOccurred())
			var headerCount int
			pgStr = `SELECT COUNT(*) FROM btc.header_cids`
			err = tx.Get(&headerCount, pgStr)
			Expect(err).ToNot(HaveOccurred())
			var blocksCount int
			pgStr = `SELECT COUNT(*) FROM public.blocks`
			err = tx.Get(&blocksCount, pgStr)
			Expect(err).ToNot(HaveOccurred())

			err = tx.Commit()
			Expect(err).ToNot(HaveOccurred())

			Expect(blocksCount).To(Equal(0))
			Expect(txCount).To(Equal(0))
			Expect(txInCount).To(Equal(0))
			Expect(txOutCount).To(Equal(0))
			Expect(headerCount).To(Equal(0))
		})
		It("Cleans transactions", func() {
			err := cleaner.Clean(rngs, shared.Transactions)
			Expect(err).ToNot(HaveOccurred())

			tx, err := db.Beginx()
			Expect(err).ToNot(HaveOccurred())
			var txCount int
			pgStr := `SELECT COUNT(*) FROM btc.transaction_cids`
			err = tx.Get(&txCount, pgStr)
			Expect(err).ToNot(HaveOccurred())
			var txInCount int
			pgStr = `SELECT COUNT(*) FROM btc.tx_inputs`
			err = tx.Get(&txInCount, pgStr)
			Expect(err).ToNot(HaveOccurred())
			var txOutCount int
			pgStr = `SELECT COUNT(*) FROM btc.tx_outputs`
			err = tx.Get(&txOutCount, pgStr)
			Expect(err).ToNot(HaveOccurred())
			var headerCount int
			pgStr = `SELECT COUNT(*) FROM btc.header_cids`
			err = tx.Get(&headerCount, pgStr)
			Expect(err).ToNot(HaveOccurred())
			var blocksCount int
			pgStr = `SELECT COUNT(*) FROM public.blocks`
			err = tx.Get(&blocksCount, pgStr)
			Expect(err).ToNot(HaveOccurred())

			err = tx.Commit()
			Expect(err).ToNot(HaveOccurred())

			Expect(blocksCount).To(Equal(2))
			Expect(txCount).To(Equal(0))
			Expect(txInCount).To(Equal(0))
			Expect(txOutCount).To(Equal(0))
			Expect(headerCount).To(Equal(2))
		})
	})

	Describe("ResetValidation", func() {
		BeforeEach(func() {
			for _, key := range mhKeys {
				_, err := db.Exec(`INSERT INTO public.blocks (key, data) VALUES ($1, $2)`, key, mockData)
				Expect(err).ToNot(HaveOccurred())
			}

			err := repo.Index(mockCIDPayload1)
			Expect(err).ToNot(HaveOccurred())
			err = repo.Index(mockCIDPayload2)
			Expect(err).ToNot(HaveOccurred())

			var validationTimes []int
			pgStr := `SELECT times_validated FROM btc.header_cids`
			err = db.Select(&validationTimes, pgStr)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(validationTimes)).To(Equal(2))
			Expect(validationTimes[0]).To(Equal(1))
			Expect(validationTimes[1]).To(Equal(1))

			err = repo.Index(mockCIDPayload1)
			Expect(err).ToNot(HaveOccurred())

			validationTimes = []int{}
			pgStr = `SELECT times_validated FROM btc.header_cids ORDER BY block_number`
			err = db.Select(&validationTimes, pgStr)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(validationTimes)).To(Equal(2))
			Expect(validationTimes[0]).To(Equal(2))
			Expect(validationTimes[1]).To(Equal(1))
		})
		AfterEach(func() {
			btc.TearDownDB(db)
		})
		It("Resets the validation level", func() {
			err := cleaner.ResetValidation(rngs)
			Expect(err).ToNot(HaveOccurred())

			var validationTimes []int
			pgStr := `SELECT times_validated FROM btc.header_cids`
			err = db.Select(&validationTimes, pgStr)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(validationTimes)).To(Equal(2))
			Expect(validationTimes[0]).To(Equal(0))
			Expect(validationTimes[1]).To(Equal(0))

			err = repo.Index(mockCIDPayload2)
			Expect(err).ToNot(HaveOccurred())

			validationTimes = []int{}
			pgStr = `SELECT times_validated FROM btc.header_cids ORDER BY block_number`
			err = db.Select(&validationTimes, pgStr)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(validationTimes)).To(Equal(2))
			Expect(validationTimes[0]).To(Equal(0))
			Expect(validationTimes[1]).To(Equal(1))
		})
	})
})

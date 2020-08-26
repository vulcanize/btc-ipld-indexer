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
	"encoding/hex"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
)

// Converter interface for substituting mocks in tests
type Converter interface {
	Convert(payload BlockPayload) (*ConvertedPayload, error)
}

// PayloadConverter satisfies the PayloadConverter interface for bitcoin
type PayloadConverter struct {
	chainConfig *chaincfg.Params
}

// NewPayloadConverter creates a pointer to a new PayloadConverter which satisfies the PayloadConverter interface
func NewPayloadConverter(chainConfig *chaincfg.Params) *PayloadConverter {
	return &PayloadConverter{
		chainConfig: chainConfig,
	}
}

// Convert method is used to convert a bitcoin BlockPayload to an IPLDPayload
// Satisfies the shared.PayloadConverter interface
func (pc *PayloadConverter) Convert(payload BlockPayload) (*ConvertedPayload, error) {
	txMeta := make([]TxModelWithInsAndOuts, len(payload.Txs))
	for i, tx := range payload.Txs {
		txModel := TxModelWithInsAndOuts{
			TxHash:    tx.Hash().String(),
			Index:     int64(i),
			SegWit:    tx.HasWitness(),
			TxOutputs: make([]TxOutput, len(tx.MsgTx().TxOut)),
			TxInputs:  make([]TxInput, len(tx.MsgTx().TxIn)),
		}
		if tx.HasWitness() {
			txModel.WitnessHash = tx.WitnessHash().String()
		}
		for i, in := range tx.MsgTx().TxIn {
			txModel.TxInputs[i] = TxInput{
				Index:                 int64(i),
				SignatureScript:       in.SignatureScript,
				PreviousOutPointHash:  in.PreviousOutPoint.Hash.String(),
				PreviousOutPointIndex: in.PreviousOutPoint.Index,
				TxWitness:             convertBytesToHexArray(in.Witness),
			}
		}
		for i, out := range tx.MsgTx().TxOut {
			scriptClass, addresses, numberOfSigs, err := txscript.ExtractPkScriptAddrs(out.PkScript, pc.chainConfig)
			// if we receive an error but the txscript type isn't NonStandardTy then something went wrong
			if err != nil && scriptClass != txscript.NonStandardTy {
				return nil, err
			}
			stringAddrs := make([]string, len(addresses))
			for i, addr := range addresses {
				stringAddrs[i] = addr.EncodeAddress()
			}
			txModel.TxOutputs[i] = TxOutput{
				Index:        int64(i),
				Value:        out.Value,
				PkScript:     out.PkScript,
				RequiredSigs: int64(numberOfSigs),
				ScriptClass:  uint8(scriptClass),
				Addresses:    stringAddrs,
			}
		}
		txMeta[i] = txModel
	}
	return &ConvertedPayload{
		BlockPayload: payload,
		TxMetaData:   txMeta,
	}, nil
}

func convertBytesToHexArray(bytea [][]byte) []string {
	var strs []string
	for _, b := range bytea {
		strs = append(strs, hex.EncodeToString(b))
	}
	return strs
}

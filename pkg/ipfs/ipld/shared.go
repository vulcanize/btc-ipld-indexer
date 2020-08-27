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

package ipld

import (
	"github.com/ipfs/go-cid"
	mh "github.com/multiformats/go-multihash"
)

// IPLD Codecs for Ethereum
// See the authoritative document:
// https://github.com/multiformats/multicodec/blob/master/table.csv
const (
	MEthStateTrie   = 0x96
	MEthStorageTrie = 0x98
	MBitcoinHeader  = 0xb0
	MBitcoinTx      = 0xb1
)

// RawdataToCid takes the desired codec and a slice of bytes
// and returns the proper cid of the object.
func RawdataToCid(codec uint64, rawdata []byte, multiHash uint64) (cid.Cid, error) {
	c, err := cid.Prefix{
		Codec:    codec,
		Version:  1,
		MhType:   multiHash,
		MhLength: -1,
	}.Sum(rawdata)
	if err != nil {
		return cid.Cid{}, err
	}
	return c, nil
}

// sha256ToCid takes a sha246 hash and returns its cid based on the
// codec given
func sha256ToCid(codec uint64, h []byte) cid.Cid {
	hash, err := mh.Encode(h, mh.DBL_SHA2_256)
	if err != nil {
		panic(err)
	}

	return cid.NewCidV1(codec, hash)
}

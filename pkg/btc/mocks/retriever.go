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

package mocks

import (
	"github.com/vulcanize/ipld-btc-indexer/pkg/btc"
)

// CIDRetriever is a mock CID retriever for use in tests
type CIDRetriever struct {
	GapsToRetrieve              []btc.DBGap
	GapsToRetrieveErr           error
	CalledTimes                 int
	FirstBlockNumberToReturn    int64
	RetrieveFirstBlockNumberErr error
}

// RetrieveLastBlockNumber mock method
func (*CIDRetriever) RetrieveLastBlockNumber() (int64, error) {
	panic("implement me")
}

// RetrieveFirstBlockNumber mock method
func (mcr *CIDRetriever) RetrieveFirstBlockNumber() (int64, error) {
	return mcr.FirstBlockNumberToReturn, mcr.RetrieveFirstBlockNumberErr
}

// RetrieveGapsInData mock method
func (mcr *CIDRetriever) RetrieveGapsInData(int) ([]btc.DBGap, error) {
	mcr.CalledTimes++
	return mcr.GapsToRetrieve, mcr.GapsToRetrieveErr
}

// SetGapsToRetrieve mock method
func (mcr *CIDRetriever) SetGapsToRetrieve(gaps []btc.DBGap) {
	if mcr.GapsToRetrieve == nil {
		mcr.GapsToRetrieve = make([]btc.DBGap, 0)
	}
	mcr.GapsToRetrieve = append(mcr.GapsToRetrieve, gaps...)
}

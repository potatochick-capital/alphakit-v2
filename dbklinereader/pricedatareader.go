package dbklinereader

import (
	"time"

	"github.com/thecolngroup/alphakit/market"
)

type PriceDataReader struct {
	klines    []*market.Kline
	positions uint64
}

func (p *PriceDataReader) Next() (*market.Kline, error) {
	if p.positions >= uint64(len(p.klines)) {
		return nil, nil
	}

	kline := p.klines[p.positions]
	p.positions++

	return kline, nil
}

func NewDataBasePriceDataReader(assetsId uint64, startTime, endTime time.Time) *PriceDataReader {
	return nil
}

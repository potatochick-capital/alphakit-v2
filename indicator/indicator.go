package indicator

import (
	"github.com/potatochick-capital/alphakit-v2/market"
	"github.com/shopspring/decimal"
)

type SwingPoint struct {
	Index int
	Price decimal.Decimal
	Label string // e.g. "SHH", "SLL", "LHL", "LLH"
}

type Indicator interface {
	Update(i int, kline market.Kline)
	GetSwingPoints() ([]SwingPoint, []SwingPoint)
}

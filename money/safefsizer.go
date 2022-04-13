package money

import (
	"math"

	"github.com/colngroup/zero2algo/dec"
	"github.com/shopspring/decimal"
)

type SafeFSizer struct {
	InitialCapital decimal.Decimal
	F              float64
	ScaleF         float64
}

func NewSafeFSizer(initialCapital decimal.Decimal, f, scaleF float64) *SafeFSizer {
	return &SafeFSizer{
		InitialCapital: initialCapital,
		F:              f,
		ScaleF:         scaleF,
	}
}

func (s *SafeFSizer) Size(price, capital, risk decimal.Decimal) decimal.Decimal {

	sqrtGrowthFactor := 1.0
	profit := capital.Sub(s.InitialCapital)
	if profit.IsPositive() {
		capitalGrowthFactor := 1 + profit.Div(capital).InexactFloat64()
		sqrtGrowthFactor = math.Sqrt(capitalGrowthFactor)
	}
	safeF := s.F * s.ScaleF * sqrtGrowthFactor
	margin := capital.InexactFloat64() * safeF

	size := margin / risk.InexactFloat64()

	return dec.New(size)
}

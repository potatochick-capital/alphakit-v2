// Copyright 2022 The Coln Group Ltd
// SPDX-License-Identifier: MIT

package dbklinereader

import (
	"time"

	"github.com/thecolngroup/alphakit/market"
)

// KlineReader is an interface for reading candlesticks.
type KlineReader interface {
	ReadAll(startDate time.Time, endDate time.Time, assetsId uint64) (map[uint64][]market.Kline, error)
}

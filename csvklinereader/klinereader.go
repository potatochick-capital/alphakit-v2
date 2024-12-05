// Copyright 2022 The Coln Group Ltd
// SPDX-License-Identifier: MIT

package csvklinereader

import "github.com/potatochick-capital/alphakit-v2/market"

// KlineReader is an interface for reading candlesticks.
type KlineReader interface {
	Read() (market.Kline, error)
	ReadAll() ([]market.Kline, error)
}

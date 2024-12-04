// Copyright 2022 The Coln Group Ltd
// SPDX-License-Identifier: MIT

package dbklinereader

import "github.com/thecolngroup/alphakit/market"

// KlineReader is an interface for reading candlesticks.
type KlineReader interface {
	ReadAll() ([]market.Kline, error)
}

// Copyright 2022 The Coln Group Ltd
// SPDX-License-Identifier: MIT

// Package studyrun is internal and not part of exported API.
package studyrun

import (
	"errors"

	"github.com/potatochick-capital/alphakit-v2/broker"
	"github.com/potatochick-capital/alphakit-v2/broker/backtest"
)

// readDealerFromConfig creates a new simulated dealer from a config file params.
func readDealerFromConfig(config map[string]any) (broker.MakeSimulatedDealer, error) {
	var makeDealer broker.MakeSimulatedDealer

	if _, ok := config["dealer"]; !ok {
		return nil, errors.New("'dealer' key not found")
	}
	root := config["dealer"].(map[string]any)
	makeDealer = func() (broker.SimulatedDealer, error) {
		return backtest.MakeDealerFromConfig(root)
	}

	return makeDealer, nil
}

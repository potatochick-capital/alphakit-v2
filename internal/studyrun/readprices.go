// Copyright 2022 The Coln Group Ltd
// SPDX-License-Identifier: MIT

package studyrun

import (
	"errors"
	"fmt"

	"github.com/potatochick-capital/alphakit-v2/csvklinereader"
	"github.com/potatochick-capital/alphakit-v2/market"
	"github.com/potatochick-capital/alphakit-v2/optimize"
	"github.com/thecolngroup/gou/conv"
)

// readPricesFromConfig reads the price samples from a config file params.
func readPricesFromConfig(config map[string]any, typeRegistry map[string]any) (map[optimize.AssetId][]market.Kline, error) {

	if _, ok := config["samples"]; !ok {
		return nil, errors.New("'samples' key not found")
	}
	root := config["samples"].([]any)
	samples := make(map[optimize.AssetId][]market.Kline)

	for _, sub := range root {

		cfg := sub.(map[string]any)

		// Load decoder from type registry
		if _, ok := cfg["decoder"]; !ok {
			return nil, errors.New("'decoder' key not found")
		}
		decoder := conv.ToString(cfg["decoder"])
		if _, ok := typeRegistry[decoder]; !ok {
			return nil, fmt.Errorf("'%s' key not found in type registry", decoder)
		}
		maker := typeRegistry[decoder].(csvklinereader.MakeCSVKlineReader)

		// Load path to price files from config
		path := cfg["path"].(string)
		series, err := csvklinereader.ReadKlinesFromCSVWithDecoder(path, maker)
		if err != nil {
			return nil, err
		}

		// Load asset key from config
		assetId := optimize.AssetId(cfg["asset"].(string))
		samples[assetId] = series
	}

	return samples, nil
}

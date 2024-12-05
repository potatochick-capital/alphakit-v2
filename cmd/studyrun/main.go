// Copyright 2022 The Coln Group Ltd
// SPDX-License-Identifier: MIT

package main

import (
	"log"
	"os"

	"github.com/potatochick-capital/alphakit-v2/cmd/studyrun/app"
	"github.com/potatochick-capital/alphakit-v2/csvklinereader"
	"github.com/potatochick-capital/alphakit-v2/trader"
	"github.com/potatochick-capital/alphakit-v2/trader/hodl"
	"github.com/potatochick-capital/alphakit-v2/trader/trend"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		log.Fatal(err)
	}
}

func run(args []string) error {
	return app.Run(
		args,
		map[string]any{
			"hodl":        trader.MakeFromConfig(hodl.MakeBotFromConfig),
			"trend.cross": trader.MakeFromConfig(trend.MakeCrossBotFromConfig),
			"trend.apex":  trader.MakeFromConfig(trend.MakeApexBotFromConfig),
			"binance":     csvklinereader.MakeCSVKlineReader(csvklinereader.NewBinanceCSVKlineReader),
			"metatrader":  csvklinereader.MakeCSVKlineReader(csvklinereader.NewMetaTraderCSVKlineReader),
		},
		app.BuildVersion{
			GitTag:    buildGitTag,
			GitCommit: buildGitCommit,
			Time:      buildTime,
			User:      buildUser,
		},
	)
}

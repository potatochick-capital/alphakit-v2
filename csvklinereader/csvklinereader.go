// Copyright 2022 The Coln Group Ltd
// SPDX-License-Identifier: MIT

package csvklinereader

import (
	"encoding/csv"
	"io"

	"github.com/potatochick-capital/alphakit-v2/market"
)

var _ KlineReader = (*CSVKlineReader)(nil)

// CSVKlineReader is a KlineReader that reads from a CSV file.
type CSVKlineReader struct {
	csv     *csv.Reader
	decoder CSVKlineDecoder
}

// MakeCSVKlineReader is a factory method type that creates a new CSVKlineReader.
type MakeCSVKlineReader func(csv *csv.Reader) *CSVKlineReader

// NewCSVKlineReader creates a new CSVKlineReader with the default Binance decoder.
func NewCSVKlineReader(csv *csv.Reader) *CSVKlineReader {
	return &CSVKlineReader{
		csv:     csv,
		decoder: BinanceCSVKlineDecoder,
	}
}

// NewCSVKlineReaderWithDecoder creates a new CSVKlineReader with the given decoder.
func NewCSVKlineReaderWithDecoder(csv *csv.Reader, decoder CSVKlineDecoder) *CSVKlineReader {
	return &CSVKlineReader{
		csv:     csv,
		decoder: decoder,
	}
}

// Read reads the next Kline from the underlying CSV data.
func (r *CSVKlineReader) Read() (market.Kline, error) {
	var k market.Kline

	rec, err := r.csv.Read()
	if err != nil {
		return k, err
	}

	return r.decoder(rec)
}

// ReadAll reads all the Klines from the underlying CSV data.
func (r *CSVKlineReader) ReadAll() ([]market.Kline, error) {
	var ks []market.Kline
	for {
		k, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		ks = append(ks, k)
	}

	return ks, nil
}

package dbklinereader

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
	"github.com/thecolngroup/alphakit/market"
)

type TimescaleKlineReader struct {
	pool *pgxpool.Pool
}

func NewTimescaleKlineReader(connectionURI string) (*TimescaleKlineReader, error) {
	config, err := pgxpool.ParseConfig(connectionURI)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection URI: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	return &TimescaleKlineReader{pool: pool}, nil
}

func (r *TimescaleKlineReader) Close() {
	r.pool.Close()
}

func (r *TimescaleKlineReader) ReadAll(startDate, endDate time.Time, assetID uint64) (map[uint64][]market.Kline, error) {
	// Define the time frames and their corresponding table names
	timeFrames := map[uint64]string{
		1:     "bar_1s",
		60:    "bar_1m",
		900:   "bar_15m",
		1800:  "bar_30m",
		3600:  "bar_1h",
		14400: "bar_4h",
		86400: "bar_1d",
	}

	// Result map to store klines for each time frame
	result := make(map[uint64][]market.Kline)

	// Query each time frame
	for timeFrame, tableName := range timeFrames {
		query := fmt.Sprintf(`
			SELECT 
				bar_time,
				open,
				high,
				low,
				close,
				volume
			FROM price_data.%s
			WHERE 
				asset_id = $1 AND 
				bar_time BETWEEN $2 AND $3
			ORDER BY bar_time
		`, tableName)

		rows, err := r.pool.Query(context.Background(), query, assetID, startDate, endDate)
		if err != nil {
			return nil, fmt.Errorf("error querying %s: %w", tableName, err)
		}
		defer rows.Close()

		var klines []market.Kline
		for rows.Next() {
			var k market.Kline
			var openDecimal, highDecimal, lowDecimal, closeDecimal decimal.Decimal

			err := rows.Scan(
				&k.Start,
				&openDecimal,
				&highDecimal,
				&lowDecimal,
				&closeDecimal,
				&k.Volume,
			)
			if err != nil {
				return nil, fmt.Errorf("error scanning row from %s: %w", tableName, err)
			}

			// Convert decimal.Decimal to Kline struct
			k.O = openDecimal
			k.H = highDecimal
			k.L = lowDecimal
			k.C = closeDecimal

			klines = append(klines, k)
		}

		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("error after scanning rows from %s: %w", tableName, err)
		}

		// Only add to result if there are klines
		if len(klines) > 0 {
			result[timeFrame] = klines
		}
	}

	return result, nil
}

// Ensure TimescaleKlineReader implements KlineReader interface
var _ KlineReader = (*TimescaleKlineReader)(nil)

package dbklinereader

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/potatochick-capital/alphakit-v2/market"
	"github.com/shopspring/decimal"
)

type TimescaleKlineReader struct {
	pool *pgxpool.Pool
}

// NewTimescaleKlineReader initializes a new TimescaleKlineReader with the provided connection URI.
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

// Close gracefully closes the connection pool.
func (r *TimescaleKlineReader) Close() {
	r.pool.Close()
}

// ReadAll retrieves kline data for the specified asset ID and time range across different timeframes.
func (r *TimescaleKlineReader) ReadAll(startDate, endDate time.Time, assetId uint64) (map[uint64][]*market.Kline, error) {
	// Mapping of timeframe in seconds to their respective table name prefixes.
	timeFrames := map[uint64]string{
		1:     "bar_1s",
		60:    "bar_1m",
		900:   "bar_15m",
		1800:  "bar_30m",
		3600:  "bar_1h",
		14400: "bar_4h",
		86400: "bar_1d",
	}

	result := make(map[uint64][]*market.Kline)

	for timeFrame, tablePrefix := range timeFrames {
		// Dynamically construct the table name based on the assetId.
		tableName := fmt.Sprintf("price_data.%s_asset_%d", tablePrefix, assetId)

		// Construct the SQL query with the dynamically generated table name.
		query := fmt.Sprintf(`
            SELECT 
                bar_time,
                open,
                high,
                low,
                close,
                volume
            FROM %s
            WHERE 
                bar_time BETWEEN $1 AND $2
            ORDER BY bar_time
        `, tableName)

		// Execute the query with startDate and endDate as parameters.
		rows, err := r.pool.Query(context.Background(), query, startDate, endDate)
		if err != nil {
			return nil, fmt.Errorf("error querying %s: %w", tableName, err)
		}
		defer rows.Close()

		var klines []*market.Kline
		for rows.Next() {
			k := &market.Kline{}
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

			k.O = openDecimal
			k.H = highDecimal
			k.L = lowDecimal
			k.C = closeDecimal

			klines = append(klines, k)
		}

		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("error after scanning rows from %s: %w", tableName, err)
		}

		if len(klines) > 0 {
			result[timeFrame] = klines
		}
	}

	return result, nil
}

// Ensure TimescaleKlineReader implements KlineReader interface
var _ KlineReader = (*TimescaleKlineReader)(nil)

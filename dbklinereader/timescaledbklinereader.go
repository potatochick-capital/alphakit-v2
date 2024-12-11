package dbklinereader

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/potatochick-capital/alphakit-v2/market"
	"github.com/shopspring/decimal"
)

type TimescaleKlineReader struct {
	pool            *pgxpool.Pool
	pageSize        int
	largeTimeframes map[uint64]bool
}

// NewTimescaleKlineReader initializes a new TimescaleKlineReader with the provided connection URI.
// It also sets the page size and identifies large timeframes that require paging.
func NewTimescaleKlineReader(connectionURI string) (*TimescaleKlineReader, error) {
	config, err := pgxpool.ParseConfig(connectionURI)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection URI: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Define which timeframes are considered large and require paging
	largeTimeframes := map[uint64]bool{
		1:  true, // 1s
		60: true, // 1m
	}

	return &TimescaleKlineReader{
		pool:            pool,
		pageSize:        100000, // Define the page size as 100,000
		largeTimeframes: largeTimeframes,
	}, nil
}

// Close gracefully closes the connection pool.
func (r *TimescaleKlineReader) Close() {
	r.pool.Close()
}

// ReadAll retrieves kline data for the specified asset Id and time range across different timeframes.
// It implements paging for large timeframes (1s and 1m) to handle large datasets efficiently.
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

		if r.largeTimeframes[timeFrame] {
			// Implement paging for large timeframes
			klines, err := r.fetchKlinesWithPaging(tableName, startDate, endDate)
			if err != nil {
				return nil, fmt.Errorf("error fetching paged data from %s: %w", tableName, err)
			}
			if len(klines) > 0 {
				result[timeFrame] = klines
			}
		} else {
			// Regular query for smaller timeframes
			klines, err := r.fetchKlines(tableName, startDate, endDate)
			if err != nil {
				return nil, fmt.Errorf("error fetching data from %s: %w", tableName, err)
			}
			if len(klines) > 0 {
				result[timeFrame] = klines
			}
		}
	}

	return result, nil
}

// fetchKlines retrieves kline data without paging for smaller timeframes.
func (r *TimescaleKlineReader) fetchKlines(tableName string, startDate, endDate time.Time) ([]*market.Kline, error) {
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
        ORDER BY bar_time ASC
    `, tableName)

	rows, err := r.pool.Query(context.Background(), query, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("error executing query on %s: %w", tableName, err)
	}
	defer rows.Close()

	var klines []*market.Kline
	for rows.Next() {
		k, err := scanRow(rows, tableName)
		if err != nil {
			return nil, err
		}
		klines = append(klines, k)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows from %s: %w", tableName, err)
	}

	return klines, nil
}

// fetchKlinesWithPaging retrieves kline data with paging for large timeframes.
func (r *TimescaleKlineReader) fetchKlinesWithPaging(tableName string, startDate, endDate time.Time) ([]*market.Kline, error) {
	var klines []*market.Kline
	lastTime := startDate

	for {
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
                 bar_time > $1 AND 
                 bar_time <= $2
             ORDER BY bar_time ASC
             LIMIT %d
         `, tableName, r.pageSize)

		rows, err := r.pool.Query(context.Background(), query, lastTime, endDate)
		if err != nil {
			return nil, fmt.Errorf("error executing paged query on %s: %w", tableName, err)
		}

		pageKlines, err := r.scanRows(rows, tableName)
		rows.Close()
		if err != nil {
			return nil, err
		}

		if len(pageKlines) == 0 {
			// No more records to fetch
			break
		}

		klines = append(klines, pageKlines...)

		// Update lastTime to the bar_time of the last record fetched
		lastTime = pageKlines[len(pageKlines)-1].Start

		// If fewer records than pageSize were fetched, we've reached the end
		if len(pageKlines) < r.pageSize {
			break
		}
	}

	return klines, nil
}

// scanRows scans multiple rows and returns a slice of Kline pointers.
func (r *TimescaleKlineReader) scanRows(rows pgx.Rows, tableName string) ([]*market.Kline, error) {
	var klines []*market.Kline

	for rows.Next() {
		k, err := scanRow(rows, tableName)
		if err != nil {
			return nil, err
		}
		klines = append(klines, k)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows from %s: %w", tableName, err)
	}

	return klines, nil
}

// scanRow scans a single row into a Kline struct.
func scanRow(rows pgx.Rows, tableName string) (*market.Kline, error) {
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

	return k, nil
}

// Ensure TimescaleKlineReader implements KlineReader interface
var _ KlineReader = (*TimescaleKlineReader)(nil)

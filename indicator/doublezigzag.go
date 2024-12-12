// File: double_zigzag_indicator.go
package indicator

import (
	"fmt"

	"github.com/potatochick-capital/alphakit-v2/market"
	"github.com/shopspring/decimal"
)

// idxVal is a helper struct to manage indices and their corresponding values.
type idxVal struct {
	index int
	value decimal.Decimal
}

// DoubleZigZagIndicator represents the double ZigZag indicator with short and long periods.
type DoubleZigZagIndicator struct {
	periodShort int
	periodLong  int

	pivotsShort []SwingPoint
	pivotsLong  []SwingPoint

	dirShort int // 1=up, -1=down, 0=none
	dirLong  int // 1=up, -1=down, 0=none

	maxDequeShort []idxVal
	minDequeShort []idxVal
	maxDequeLong  []idxVal
	minDequeLong  []idxVal

	// Last committed swing points
	lastCommittedShort *SwingPoint
	lastCommittedLong  *SwingPoint
}

// NewDoubleZigZagIndicator creates a new instance of the Double ZigZag indicator.
func NewDoubleZigZagIndicator(shortPeriod, longPeriod int) *DoubleZigZagIndicator {
	return &DoubleZigZagIndicator{
		periodShort: shortPeriod,
		periodLong:  longPeriod,
		dirShort:    0,
		dirLong:     0,
	}
}

// Update processes a new Kline data point and updates the ZigZag indicators.
func (d *DoubleZigZagIndicator) Update(i int, kline market.Kline) {
	// Update short period deques
	updateMaxDeque(&d.maxDequeShort, i, kline.H)
	updateMinDeque(&d.minDequeShort, i, kline.L)

	if i >= d.periodShort {
		popFrontIfOutOfRange(&d.maxDequeShort, i, d.periodShort)
		popFrontIfOutOfRange(&d.minDequeShort, i, d.periodShort)
	}

	// Update long period deques
	updateMaxDeque(&d.maxDequeLong, i, kline.H)
	updateMinDeque(&d.minDequeLong, i, kline.L)

	if i >= d.periodLong {
		popFrontIfOutOfRange(&d.maxDequeLong, i, d.periodLong)
		popFrontIfOutOfRange(&d.minDequeLong, i, d.periodLong)
	}

	// Check if short pivot
	var shortPH, shortPL bool
	if i >= d.periodShort-1 {
		if len(d.maxDequeShort) > 0 && d.maxDequeShort[0].index == i {
			shortPH = true
		}
		if len(d.minDequeShort) > 0 && d.minDequeShort[0].index == i {
			shortPL = true
		}
	}

	// Check if long pivot
	var longPH, longPL bool
	if i >= d.periodLong-1 {
		if len(d.maxDequeLong) > 0 && d.maxDequeLong[0].index == i {
			longPH = true
		}
		if len(d.minDequeLong) > 0 && d.minDequeLong[0].index == i {
			longPL = true
		}
	}

	// Short period logic
	if shortPH && !shortPL {
		newDir := 1
		d.handleZigZag(&d.pivotsShort, &d.dirShort, &d.lastCommittedShort, kline.H, i, newDir, "S")
	} else if shortPL && !shortPH {
		newDir := -1
		d.handleZigZag(&d.pivotsShort, &d.dirShort, &d.lastCommittedShort, kline.L, i, newDir, "S")
	}

	// Long period logic
	if longPH && !longPL {
		newDir := 1
		d.handleZigZag(&d.pivotsLong, &d.dirLong, &d.lastCommittedLong, kline.H, i, newDir, "L")
	} else if longPL && !longPH {
		newDir := -1
		d.handleZigZag(&d.pivotsLong, &d.dirLong, &d.lastCommittedLong, kline.L, i, newDir, "L")
	}
}

// handleZigZag processes the ZigZag logic for either short or long period.
func (d *DoubleZigZagIndicator) handleZigZag(pivots *[]SwingPoint, dir *int, lastCommitted **SwingPoint, newPrice decimal.Decimal, idx int, newDir int, prefix string) {
	if *dir == 0 {
		// Initial direction setup
		*pivots = append(*pivots, SwingPoint{Index: idx, Price: newPrice, Label: ""})
		*dir = newDir
		return
	}

	if newDir != *dir {
		// Direction change detected, commit the previous swing point
		if len(*pivots) > 0 {
			lastSwing := (*pivots)[len(*pivots)-1]
			*lastCommitted = &lastSwing
		}
		// Add new swing point
		*pivots = append(*pivots, SwingPoint{Index: idx, Price: newPrice, Label: ""})
		*dir = newDir
	} else {
		// Same direction, possibly update the last swing point
		last := (*pivots)[len(*pivots)-1]
		updated := false
		if newDir == 1 && newPrice.GreaterThan(last.Price) {
			(*pivots)[len(*pivots)-1].Price = newPrice
			(*pivots)[len(*pivots)-1].Index = idx
			updated = true
		} else if newDir == -1 && newPrice.LessThan(last.Price) {
			(*pivots)[len(*pivots)-1].Price = newPrice
			(*pivots)[len(*pivots)-1].Index = idx
			updated = true
		}

		if updated {
			// Update label if needed
			*pivots = labelZigZagSwings(*pivots, *dir, prefix)
		}
	}
}

// GetSwingPoints returns all identified swing points for both long and short periods.
func (d *DoubleZigZagIndicator) GetSwingPoints() ([]SwingPoint, []SwingPoint) {
	return d.pivotsLong, d.pivotsShort
}

// GetLastCommittedSwingPoints returns the last committed swing points for both short and long periods.
// If no committed swing point exists for a period, the corresponding return value will be nil.
// The parameter 'n' is optional:
// - n = 0 (default): Last committed swing point
// - n = 1: One before the last committed swing point
// - n = 2: Two before the last committed swing point
func (d *DoubleZigZagIndicator) GetLastCommittedSwingPoints(n ...uint64) (*SwingPoint, *SwingPoint, error) {
	index := uint64(0) // Default value
	if len(n) > 0 {
		index = n[0]
	}

	var shortPoint *SwingPoint
	var longPoint *SwingPoint
	var err error

	// Retrieve the n-th last committed short swing point
	if len(d.pivotsShort) > int(index) {
		shortPoint = &d.pivotsShort[len(d.pivotsShort)-1-int(index)]
	} else {
		err = fmt.Errorf("not enough committed short swing points")
	}

	// Retrieve the n-th last committed long swing point
	if len(d.pivotsLong) > int(index) {
		longPoint = &d.pivotsLong[len(d.pivotsLong)-1-int(index)]
	} else {
		if err != nil {
			err = fmt.Errorf("%v; not enough committed long swing points", err)
		} else {
			err = fmt.Errorf("not enough committed long swing points")
		}
	}

	return shortPoint, longPoint, err
}

// --------------------- Helper Functions ---------------------

// labelZigZagSwings assigns labels (HH, LH, HL, LL) to the latest swing point based on direction.
func labelZigZagSwings(pivots []SwingPoint, dir int, prefix string) []SwingPoint {
	if len(pivots) < 2 {
		return pivots
	}
	lastIndex := len(pivots) - 1
	lastPivot := pivots[lastIndex]

	prevSameTypeIndex := lastIndex - 1
	if prevSameTypeIndex < 0 {
		// Not enough data
		return pivots
	}

	prevSameTypePivot := pivots[prevSameTypeIndex]

	var label string
	if dir == 1 {
		if lastPivot.Price.GreaterThan(prevSameTypePivot.Price) {
			label = prefix + "HH"
		} else {
			label = prefix + "LH"
		}
	} else if dir == -1 {
		if lastPivot.Price.GreaterThan(prevSameTypePivot.Price) {
			label = prefix + "HL"
		} else {
			label = prefix + "LL"
		}
	}

	pivots[lastIndex].Label = label
	return pivots
}

// updateMaxDeque maintains a deque for tracking maximum values within the specified period.
func updateMaxDeque(d *[]idxVal, i int, val decimal.Decimal) {
	for len(*d) > 0 && (*d)[len(*d)-1].value.LessThanOrEqual(val) {
		*d = (*d)[:len(*d)-1]
	}
	*d = append(*d, idxVal{index: i, value: val})
}

// updateMinDeque maintains a deque for tracking minimum values within the specified period.
func updateMinDeque(d *[]idxVal, i int, val decimal.Decimal) {
	for len(*d) > 0 && (*d)[len(*d)-1].value.GreaterThanOrEqual(val) {
		*d = (*d)[:len(*d)-1]
	}
	*d = append(*d, idxVal{index: i, value: val})
}

// popFrontIfOutOfRange removes elements from the front of the deque if they are out of the specified range.
func popFrontIfOutOfRange(d *[]idxVal, i, period int) {
	windowStart := i - period + 1
	if len(*d) > 0 && (*d)[0].index < windowStart {
		*d = (*d)[1:]
	}
}

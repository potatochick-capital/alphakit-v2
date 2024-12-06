package indicator

import (
	"github.com/potatochick-capital/alphakit-v2/market"
	"github.com/shopspring/decimal"
)

type idxVal struct {
	index int
	value decimal.Decimal
}

type DoubleZigZagIndicator struct {
	shortPeriod int
	longPeriod  int

	pivotsShort []SwingPoint
	pivotsLong  []SwingPoint

	dirShort int // 1=up, -1=down, 0=none
	dirLong  int // 1=up, -1=down, 0=none

	shortMaxDeque []idxVal
	shortMinDeque []idxVal
	longMaxDeque  []idxVal
	longMinDeque  []idxVal
}

// NewDoubleZigZagIndicator creates a new instance of the double zigzag indicator
func NewDoubleZigZagIndicator(shortPeriod, longPeriod int) *DoubleZigZagIndicator {
	return &DoubleZigZagIndicator{
		shortPeriod: shortPeriod,
		longPeriod:  longPeriod,
		dirShort:    0,
		dirLong:     0,
	}
}

func (d *DoubleZigZagIndicator) Update(i int, kline market.Kline) {
	// Update short period deques
	updateMaxDeque(&d.shortMaxDeque, i, kline.H)
	updateMinDeque(&d.shortMinDeque, i, kline.L)

	if i >= d.shortPeriod {
		popFrontIfOutOfRange(&d.shortMaxDeque, i, d.shortPeriod)
		popFrontIfOutOfRange(&d.shortMinDeque, i, d.shortPeriod)
	}

	// Update long period deques
	updateMaxDeque(&d.longMaxDeque, i, kline.H)
	updateMinDeque(&d.longMinDeque, i, kline.L)

	if i >= d.longPeriod {
		popFrontIfOutOfRange(&d.longMaxDeque, i, d.longPeriod)
		popFrontIfOutOfRange(&d.longMinDeque, i, d.longPeriod)
	}

	// Check if short pivot
	var shortPH, shortPL bool
	if i >= d.shortPeriod-1 {
		if len(d.shortMaxDeque) > 0 && d.shortMaxDeque[0].index == i {
			shortPH = true
		}
		if len(d.shortMinDeque) > 0 && d.shortMinDeque[0].index == i {
			shortPL = true
		}
	}

	// Check if long pivot
	var longPH, longPL bool
	if i >= d.longPeriod-1 {
		if len(d.longMaxDeque) > 0 && d.longMaxDeque[0].index == i {
			longPH = true
		}
		if len(d.longMinDeque) > 0 && d.longMinDeque[0].index == i {
			longPL = true
		}
	}

	// Short period logic
	if shortPH && !shortPL {
		newDir := 1
		updateZigZag(&d.pivotsShort, &d.dirShort, kline.H, i, newDir)
		d.pivotsShort = labelZigZagSwings(d.pivotsShort, d.dirShort, "S")
	} else if shortPL && !shortPH {
		newDir := -1
		updateZigZag(&d.pivotsShort, &d.dirShort, kline.L, i, newDir)
		d.pivotsShort = labelZigZagSwings(d.pivotsShort, d.dirShort, "S")
	}

	// Long period logic
	if longPH && !longPL {
		newDir := 1
		updateZigZag(&d.pivotsLong, &d.dirLong, kline.H, i, newDir)
		d.pivotsLong = labelZigZagSwings(d.pivotsLong, d.dirLong, "L")
	} else if longPL && !longPH {
		newDir := -1
		updateZigZag(&d.pivotsLong, &d.dirLong, kline.L, i, newDir)
		d.pivotsLong = labelZigZagSwings(d.pivotsLong, d.dirLong, "L")
	}
}

func (d *DoubleZigZagIndicator) GetSwingPoints() ([]SwingPoint, []SwingPoint) {
	return d.pivotsLong, d.pivotsShort
}

// --------------------- Helper Functions ---------------------

func labelZigZagSwings(pivots []SwingPoint, dir int, prefix string) []SwingPoint {
	if len(pivots) < 2 {
		return pivots
	}
	lastIndex := len(pivots) - 1
	lastPivot := pivots[lastIndex]

	prevSameTypeIndex := lastIndex - 2
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

func updateZigZag(pivots *[]SwingPoint, dir *int, newPrice decimal.Decimal, idx int, newDir int) {
	if *dir == 0 {
		*pivots = append(*pivots, SwingPoint{Index: idx, Price: newPrice})
		*dir = newDir
		return
	}

	if newDir != *dir {
		*pivots = append(*pivots, SwingPoint{Index: idx, Price: newPrice})
		*dir = newDir
	} else {
		last := (*pivots)[len(*pivots)-1]
		if newDir == 1 && newPrice.GreaterThan(last.Price) {
			(*pivots)[len(*pivots)-1].Price = newPrice
			(*pivots)[len(*pivots)-1].Index = idx
		} else if newDir == -1 && newPrice.LessThan(last.Price) {
			(*pivots)[len(*pivots)-1].Price = newPrice
			(*pivots)[len(*pivots)-1].Index = idx
		}
	}
}

func updateMaxDeque(d *[]idxVal, i int, val decimal.Decimal) {
	for len(*d) > 0 && (*d)[len(*d)-1].value.LessThanOrEqual(val) {
		*d = (*d)[:len(*d)-1]
	}
	*d = append(*d, idxVal{index: i, value: val})
}

func updateMinDeque(d *[]idxVal, i int, val decimal.Decimal) {
	for len(*d) > 0 && (*d)[len(*d)-1].value.GreaterThanOrEqual(val) {
		*d = (*d)[:len(*d)-1]
	}
	*d = append(*d, idxVal{index: i, value: val})
}

func popFrontIfOutOfRange(d *[]idxVal, i, period int) {
	windowStart := i - period + 1
	if len(*d) > 0 && (*d)[0].index < windowStart {
		*d = (*d)[1:]
	}
}

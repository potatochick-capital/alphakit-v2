package trader

import "github.com/thecolngroup/zerotoalgo/market"

// Predicter is used by a bot to indicate price direction.
// Child packages provide specific prediction implementations.
type Predicter interface {
	market.Receiver

	// Predict gives a score between -1 (short) and +1 (long) that a
	// bot uses to generate buy and sell signals.
	Predict() float64

	// Valid indicates readiness for prediction.
	Valid() bool
}

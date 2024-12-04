// Copyright 2022 The Coln Group Ltd
// SPDX-License-Identifier: MIT

package broker

import (
	"time"

	"github.com/thecolngroup/gou/id"
)

// DealId is a unique identifier for a dealer data entity.
type DealId string

// NewId returns a new Deal seeded with the current time.
func NewId() DealId {
	return DealId(id.New())
}

// NewIdWithTime returns a new Deal seeded with the given time.
func NewIdWithTime(t time.Time) DealId {
	return DealId(id.NewWithTime(t))
}

// Copyright 2022 The Coln Group Ltd
// SPDX-License-Identifier: MIT

package broker

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/slices"
)

func TestDealId(t *testing.T) {

	ids := []DealId{
		NewId(),
		NewIdWithTime(time.Now().Add(time.Hour * 2)),
		NewIdWithTime(time.Now().Add(time.Hour * 1)),
	}
	assert.False(t, slices.IsSorted(ids))

	slices.Sort(ids)
	assert.True(t, slices.IsSorted(ids))
}

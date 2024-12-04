// Copyright 2022 The Coln Group Ltd
// SPDX-License-Identifier: MIT

package optimize

import "github.com/thecolngroup/gou/id"

// ParamSet is a set of algo parameters to trial.
type ParamSet struct {
	Id     ParamSetId `csv:"id"`
	Params ParamMap   `csv:"params"`
}

// ParamSetId is a unique identifier for a ParamSet.
type ParamSetId string

// ParamMap is a map of algo parameters.
type ParamMap map[string]any

// NewParamSet returns a new param set with initialized
func NewParamSet() ParamSet {
	return ParamSet{
		Id:     ParamSetId(id.New()),
		Params: make(map[string]any),
	}
}

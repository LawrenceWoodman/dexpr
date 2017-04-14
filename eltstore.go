/*
 * Copyright (C) 2016 Lawrence Woodman <lwoodman@vlifesystems.com>
 *
 * Licensed under an MIT licence.  Please see LICENCE.md for details.
 */

package dexpr

// The eltStore is where the composite types store there elements
type eltStore struct {
	elts map[int64][]*ENode
	num  int64
}

func newEltStore() *eltStore {
	return &eltStore{elts: map[int64][]*ENode{}, num: 0}
}

// Get returns the elements for n from eltStore
func (e *eltStore) Get(n int64) []*ENode {
	return e.elts[n]
}

// Add adds a slice of ENode's to eltStore and returns the number
// that these are stored under for use by Get
func (e *eltStore) Add(ens []*ENode) int64 {
	rNum := e.num
	e.elts[e.num] = ens
	e.num++
	return rNum
}

package main

import "sync"
import "github.com/tmthrgd/go-popcount"

// Bitmask - simple bitmask data structire based on a map
type Bitmask struct {
	I     map[uint64]uint64
	mutex *sync.RWMutex
}

// Has - basic get/set methods
func (b *Bitmask) Has(val int64) bool {
	var v = uint64(val)
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	return (b.I[v/64] & (1 << (v % 64))) != 0
}

// Insert - basic get/set methods
func (b *Bitmask) Insert(val int64) {
	var v = uint64(val)
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.I[v/64] |= (1 << (v % 64))
}

// Len - total elements in mask (non performant!)
func (b *Bitmask) Len() uint64 {
	var l uint64
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	for _, v := range b.I {
		l += popcount.CountSlice64([]uint64{v})
	}
	return l
}

// NewBitMask - constructor
func NewBitMask() *Bitmask {
	return &Bitmask{
		I:     make(map[uint64]uint64),
		mutex: &sync.RWMutex{},
	}
}

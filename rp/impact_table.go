package rp

import (
	"net"
	"sync"
)

// impactTable is the implementation of a potential impact table. It records
// addresses and number of times they are impacted. This table is expected to
// grow into a very large size in the production instance.
//
// # Future Note:
// Since this is a demo implementation a map is used, but in the future a more
// specific and optimized implementation needs to be used; for example a
// patricia tree. Additionally, the table can be chunked and mutex can be used
// on those chunks instead of locking the whole table.
type impactTable struct {
	rwMutex sync.RWMutex
	impacts map[[16]byte]uint64
}

// newImpactTable creates a new potential impact table.
func newImpactTable() *impactTable {
	return &impactTable{
		impacts: make(map[[16]byte]uint64, 0),
	}
}

// Look is used to retrieve the impact value of an address.
func (pi *impactTable) Look(a net.IP) uint64 {
	pi.rwMutex.RLock()
	defer pi.rwMutex.RUnlock()

	if impact, ok := pi.impacts[[16]byte(a.To16())]; ok {
		return impact
	}
	return 0
}

// Increment is used to increment the impact value of an address.
func (pi *impactTable) Increment(a net.IP) {
	pi.rwMutex.Lock()
	defer pi.rwMutex.Unlock()

	if impact, ok := pi.impacts[[16]byte(a.To16())]; ok {
		pi.impacts[[16]byte(a.To16())] = impact + 1
	} else {
		pi.impacts[[16]byte(a.To16())] = 1
	}
}

// ResetAll resets all of the impact recordings in the table.
func (pi *impactTable) ResetAll() {
	pi.rwMutex.Lock()
	defer pi.rwMutex.Unlock()

	// Reset all impacts. I think there is a way to reset this very very fast.
	for e := range pi.impacts {
		pi.impacts[e] = 0.0
	}
}

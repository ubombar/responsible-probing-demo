package rp

import (
	"bytes"
	"fmt"
	"net"
	"sort"
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

// String converts the impact table into a user friendly string.
func (pi *impactTable) String() string {
	pi.rwMutex.RLock()
	defer pi.rwMutex.RUnlock()

	// Collect keys
	keys := make([][16]byte, 0, len(pi.impacts))
	for k := range pi.impacts {
		keys = append(keys, k)
	}

	// Sort lexicographically (IPv6 byte order)
	sort.Slice(keys, func(i, j int) bool {
		return bytes.Compare(keys[i][:], keys[j][:]) < 0
	})

	var b bytes.Buffer
	for _, k := range keys {
		ip := net.IP(k[:])
		fmt.Fprintf(&b, "%-16s %d\n", ip.String(), pi.impacts[k])
	}

	return b.String()
}

package rp

import (
	"fmt"
	"math"
	"net"
	"sync"
	"time"
)

// responsibleIssuer is the demo implementation of how the RP algorithm should
// work. Algorithm is used to generate PDs and throttle them using the FIE
// responses.
type responsibleIssuer struct {
	mutex        sync.Mutex
	entries      []*responsibleIssuerEntry
	currentIndex int
	impactTable  *impactTable

	// ImpactCoefficientAlpha is used to denote how aggressice the L1 loss will
	// be used in the probability computation.
	ImpactCoefficientAlpha float32

	// PotentialImpactTableUpdateRate is the update frequency of the potential
	// impact table.
	PotentialImpactTableUpdateRate float32

	// TargetProbingRate is the targeted probing rate.
	TargetProbingRate float32
}

// responsibleIssuerEntry is an entry in the dTable, it contains the directive
// impact related info.
type responsibleIssuerEntry struct {
	directiveIndex int
	directive      *PD
	prob           float32
	nearAddr       net.IP
	farAddr        net.IP
	lastIssuance   time.Time
}

func (rie responsibleIssuerEntry) String() string {
	return fmt.Sprintf("D%d : [p=%v] %v(%v) --> %v(%v) %v seconds ago",
		rie.directiveIndex,
		rie.prob,
		rie.nearAddr,
		rie.directive.NearTTL,
		rie.farAddr,
		rie.directive.NearTTL+1,
		int(time.Since(rie.lastIssuance).Seconds()),
	)
}

// NewResponsibleIssuer creates a new ResponsibleProber from the given set of
// PDs. If the given list is empty then it returns a non-nil error.
func NewResponsibleIssuer(pds []*PD) (*responsibleIssuer, error) {
	if len(pds) == 0 {
		return nil, fmt.Errorf("no directives are given")
	}

	entries := make([]*responsibleIssuerEntry, 0, len(pds))

	for i, pd := range pds {
		entries = append(entries, &responsibleIssuerEntry{
			directiveIndex: i,
			directive:      pd,
			prob:           1.0,
			lastIssuance:   time.Now(),

			// Near and far addresses starts as nil.
			nearAddr: nil,
			farAddr:  nil,
		})
	}

	impactTable := newImpactTable()

	return &responsibleIssuer{
		entries:      entries,
		currentIndex: 0,
		impactTable:  impactTable,
	}, nil
}

// Issue takes a randomVariable between 0 and 1.0 and checks if the probability
// of issuance is smaller than the that. If it is then the PD is issued, if not
// it is skipped and false is returned.
func (rp *responsibleIssuer) Issue(randomVariable float32) (*PD, bool) {
	rp.mutex.Lock()
	defer rp.mutex.Unlock()

	currentEntry := rp.entries[rp.currentIndex]
	rp.currentIndex = (rp.currentIndex + 1) % len(rp.entries)

	if randomVariable > currentEntry.prob {
		return nil, false
	}

	currentEntry.lastIssuance = time.Now()

	return currentEntry.directive, true
}

// RegisterImpact adds an impact record to a specific PD and then recomputes the
// probability. If there is an error it returns a non-nil error.
func (rp *responsibleIssuer) RegisterImpact(fie *FIE) error {
	rp.mutex.Lock()
	defer rp.mutex.Unlock()

	// Since ID is unsigned there is no need to check < 0.
	if fie.ID >= uint64(len(rp.entries)) {
		return fmt.Errorf("matching FIE with PD failed, there is no PD")
	}

	matchedEntry := rp.entries[fie.ID]

	// Update the near and far addresses of the entry.
	matchedEntry.nearAddr = fie.NearAddr
	matchedEntry.farAddr = fie.FarAddr

	// Then register the address impact.
	rp.impactTable.Increment(fie.NearAddr)
	rp.impactTable.Increment(fie.FarAddr)

	// Get the last issuance and update.
	lastIssuance := float32(time.Since(matchedEntry.lastIssuance).Seconds())

	// Compute the probability.
	largestImpact := max(rp.impactTable.Look(fie.NearAddr), rp.impactTable.Look(fie.FarAddr))
	normalizedPotentialImpactRate := float32(largestImpact) / (rp.PotentialImpactTableUpdateRate * lastIssuance)
	l1Loss := float32(math.Abs(float64(rp.TargetProbingRate - normalizedPotentialImpactRate)))

	matchedEntry.prob = 1.0 / (1.0 + rp.ImpactCoefficientAlpha*l1Loss)

	// fmt.Printf("matchedEntry.prob: %v\n", normalizedPotentialImpactRate)
	// fmt.Printf("rp.PotentialImpactTableUpdateRate: %v\n", rp.PotentialImpactTableUpdateRate)
	// fmt.Printf("lastIssuance: %v\n", lastIssuance)
	fmt.Printf("largestImpact: %v\n", largestImpact)

	return nil
}

func (rp *responsibleIssuer) ResetImpacts() {
	rp.mutex.Lock()
	defer rp.mutex.Unlock()

	rp.impactTable.ResetAll()
}

func (rp *responsibleIssuer) String() string {
	rp.mutex.Lock()
	defer rp.mutex.Unlock()

	header := fmt.Sprintf("Responsible Issuer (current=%d, total=%d)", rp.currentIndex, len(rp.entries))
	body := ""

	for _, v := range rp.entries {
		body = fmt.Sprintf("%v\n%v", body, v.String())
	}
	return fmt.Sprintf("%v%v\n", header, body)
}

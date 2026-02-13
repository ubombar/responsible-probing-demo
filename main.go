package main

import (
	"fmt"
	"math/rand"
	"net"
	"responsible-probing-demo/rp"
	"time"
)

const (
	// random seed.
	RandomSeed = 42

	// response rate of the prober.
	FIEResponseRate = 1.0

	// impact coeff a.
	ImpactCoefficientAlpha = 0.02

	// impact table update rate per seconds.
	PotentialImpactTableUpdateRate = 0.05 // every 2 seconds.

	// target probing rate.
	TargetProbingRate = 4.0
)

var responseAddresses = []net.IP{
	net.ParseIP("10.1.1.0"),
	net.ParseIP("10.1.1.1"),
	net.ParseIP("10.1.1.2"),
	net.ParseIP("10.1.1.3"),
	net.ParseIP("10.1.1.4"),
	net.ParseIP("10.1.1.5"),
	net.ParseIP("10.1.1.6"),
	net.ParseIP("10.1.1.7"),
	net.ParseIP("10.1.1.8"),
	net.ParseIP("10.1.1.9"),
}

// main contains the demo for responsible probing used in the retina.
func main() {
	// New random with seed.
	r := rand.New(rand.NewSource(RandomSeed))

	// Create PDs.
	pds := []*rp.PD{
		{
			ID:              0,
			DestinationAddr: net.ParseIP("1.1.1.0"),
			NearTTL:         6,
		},
		{
			ID:              1,
			DestinationAddr: net.ParseIP("1.1.1.1"),
			NearTTL:         7,
		},
		{
			ID:              2,
			DestinationAddr: net.ParseIP("1.1.1.2"),
			NearTTL:         8,
		},
		{
			ID:              3,
			DestinationAddr: net.ParseIP("1.1.1.3"),
			NearTTL:         9,
		},
	}

	// Create the issuer.
	issuer, err := rp.NewResponsibleIssuer(pds)
	if err != nil {
		panic(err)
	}

	// Set the parameters.
	issuer.ImpactCoefficientAlpha = ImpactCoefficientAlpha
	issuer.PotentialImpactTableUpdateRate = PotentialImpactTableUpdateRate
	issuer.TargetProbingRate = TargetProbingRate

	// Reset potential impact table in parallel.
	go func() {
		potentialImpactTableUpdatePeriod := 1 / PotentialImpactTableUpdateRate

		for {
			time.Sleep(time.Second * time.Duration(potentialImpactTableUpdatePeriod))
			issuer.ResetImpacts()
		}
	}()

	for {
		fmt.Printf("%v\n", issuer)

		// Sleep for a second.
		time.Sleep(time.Millisecond * 1000)

		// Issue or not issue
		pd, ok := issuer.Issue(r.Float32())
		if !ok {
			continue
		}

		// Simulate probing.
		fie := simulateProbing(r, pd)
		if fie == nil {
			continue
		}

		// Record impact.
		if err := issuer.RegisterImpact(fie); err != nil {
			panic(err)
		}
	}
}

// simulateProbing get a rand and a directive, it simulates probing, it can
// return nil.
// The near and far addresses are selected from available addresses.
func simulateProbing(r *rand.Rand, pd *rp.PD) *rp.FIE {
	time.Sleep(time.Millisecond * time.Duration(100+r.Intn(300)))

	// addressOffset := 123

	if r.Float32() < FIEResponseRate {
		return &rp.FIE{
			ID:      pd.ID,
			NearTTL: pd.NearTTL,
			// To make sure the impact is the same on addresses, a more
			// deterministic method is used so a PD would impact the same near
			// and far addresses.
			// NearAddr: responseAddresses[int(pd.ID)+addressOffset%len(responseAddresses)],
			// NearAddr: net.ParseIP("10.10.10.1"),
			// FarAddr:  responseAddresses[int(pd.ID+1)+addressOffset%len(responseAddresses)],

			NearAddr: responseAddresses[r.Intn(len(responseAddresses))],
			FarAddr:  responseAddresses[r.Intn(len(responseAddresses))],
		}
	}

	return nil
}

package rp

import "net"

// PD is a minimal representation of a probing directive.
type PD struct {
	ID              uint64
	DestinationAddr net.IP
	NearTTL         uint16
}

// FIE is a minimal representation of a forwarding info element.
type FIE struct {
	ID       uint64
	NearAddr net.IP
	FarAddr  net.IP
	NearTTL  uint16
}

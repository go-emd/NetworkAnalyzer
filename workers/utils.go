package workers

import (
	"code.google.com/p/gopacket"
)

type Metadata struct {
	LinkFLow gopacket.Flow
	TransportFlow gopacket.Flow
	NetworkFlow gopacket.Flow
}

type Metric map[string]interface{}


package workers

import (
	"code.google.com/p/gopacket"

	"time"
)

type Metadata struct {
	Timestamp time.Time
	SrcMac, DstMac gopacket.Endpoint
	SrcIp, DstIp gopacket.Endpoint
	SrcPort, DstPort gopacket.Endpoint
}

type Metric map[string]interface{}


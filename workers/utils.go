package workers

import (
	//"code.google.com/p/gopacket"

	"time"
)

type Netflow struct {
	Start time.Time
	Duration uint64
	Protocol string
	Ipversion uint8
	SrcIp string
	DstIp string
	SrcPort uint16
	DstPort uint16
	Packets uint32
	Bytes uint32
	Flows uint32
}

/*Timestamp time.Time
SrcMac, DstMac gopacket.Endpoint
SrcIp, DstIp gopacket.Endpoint
SrcPort, DstPort gopacket.Endpoint*/

type Metric map[string]interface{}

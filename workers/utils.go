package workers

import (
	"time"
)

type Netflow struct {
	Start time.Time
	Duration time.Time
	Protocol uint8
	IpVersion uint8
	SrcIp []uint8
	DstIp []uint8
	SrcPort uint16
	DstPort uint16
	Bytes int
	Optional []byte // Protocol Specific (TCP: flags, UDP: nil, ICMP: type/code
}

type Metric map[string]interface{}

func appendByteArray(bas ...[]byte) []byte {
	fba := make([]byte, 0)

	for i := range bas {
		fba = append(fba, bas[i]...)
	}	

	return fba
}
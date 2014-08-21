package workers

import (
	"time"
)

type Netflow struct {
	Start time.Time
	Duration uint64
	Protocol uint8
	IpVersion uint8
	SrcIp []uint8
	DstIp []uint8
	SrcPort uint16
	DstPort uint16
	Bytes int
}

type Metric map[string]interface{}

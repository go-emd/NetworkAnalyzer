package workers

import (
	"github.com/go-emd/emd/log"
	"github.com/go-emd/emd/worker"

	"encoding/binary"
)

var (
	icmpFlows *Flows
)

type IcmpFlow struct {
	worker.Work
}

func (w IcmpFlow) Init() {
	for _, p := range w.Ports() {
		p.Open()
	}

	endOfFlowSeq := []byte{0x0} // Echo Reply
	icmpFlows = NewFlows(endOfFlowSeq)

	log.INFO.Println("Worker " + w.Name_ + " inited.")
}

func (w IcmpFlow) Run() {
	log.INFO.Println("IcmpFlow is running.")

	// Catch any errors that could happen
	defer func() {
		if r := recover(); r != nil {
			log.ERROR.Println("Uncaught error occurred, exiting.")
			log.ERROR.Println(r)

			w.Stop()
		}
	}()

	for {
		select {
		case cmd := <-w.Ports()["MGMT_IcmpFlow"].Channel():
			if cmd == "STOP" {
				w.Stop()
				return
			} else if cmd == "STATUS" {
				w.Ports()["MGMT_IcmpFlow"].Channel() <- "Healthy"
			} else if cmd == "METRICS" {
				w.Ports()["MGMT_IcmpFlow"].Channel() <- Metric{

					"partialFlowSize": len(icmpFlows.PartialFlows),
					"finalFlowSize": len(icmpFlows.FinalFlows),
				}
			}
		case netflow := <-w.Ports()["Sniffer_and_IcmpFlow"].Channel():
			srcPort := make([]byte, 2)
			dstPort := make([]byte, 2)

			binary.BigEndian.PutUint16(srcPort, netflow.(Netflow).SrcPort)
			binary.BigEndian.PutUint16(dstPort, netflow.(Netflow).DstPort)

			icmpFlows.Update(
				appendByteArray(
					netflow.(Netflow).Optional[0:], // Required to be the first byte
					[]byte(netflow.(Netflow).SrcIp),
					[]byte(netflow.(Netflow).DstIp),
					srcPort,
					dstPort,
					[]byte{netflow.(Netflow).IpVersion},
				),
				netflow.(Netflow),
			)
		}
	}
}

func (w IcmpFlow) Stop() {
	w.Ports()["MGMT_IcmpFlow"].Close()
	w.Ports()["Sniffer_and_IcmpFlow"].Close()

	log.INFO.Println("Worker " + w.Name_ + " stopped.")
}

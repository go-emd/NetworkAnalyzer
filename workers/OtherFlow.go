package workers

import (
	"github.com/go-emd/emd/log"
	"github.com/go-emd/emd/worker"

	"encoding/binary"
)

var (
	otherFlows *Flows
)

type OtherFlow struct {
	worker.Work
}

func (w OtherFlow) Init() {
	for _, p := range w.Ports() {
		p.Open()
	}

	endOfFlowSeq := []byte{}
	otherFlows = NewFlows(endOfFlowSeq)

	log.INFO.Println("Worker " + w.Name_ + " inited.")
}

func (w OtherFlow) Run() {
	log.INFO.Println("OtherFlow is running.")

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
		case cmd := <-w.Ports()["MGMT_OtherFlow"].Channel():
			if cmd == "STOP" {
				w.Stop()
				return
			} else if cmd == "STATUS" {
				w.Ports()["MGMT_OtherFlow"].Channel() <- "Healthy"
			} else if cmd == "METRICS" {
				w.Ports()["MGMT_OtherFlow"].Channel() <- Metric{
					"partialFlowSize": len(otherFlows.PartialFlows),
					"finalFlowSize": len(otherFlows.FinalFlows),
				}
			}
		case netflow := <-w.Ports()["Sniffer_and_OtherFlow"].Channel():
			srcPort := make([]byte, 2)
			dstPort := make([]byte, 2)

			binary.BigEndian.PutUint16(srcPort, netflow.(Netflow).SrcPort)
			binary.BigEndian.PutUint16(dstPort, netflow.(Netflow).DstPort)

			otherFlows.Update(
				appendByteArray(
					netflow.(Netflow).Optional, // Required to be the first byte
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

func (w OtherFlow) Stop() {
	w.Ports()["MGMT_OtherFlow"].Close()
	w.Ports()["Sniffer_and_OtherFlow"].Close()

	log.INFO.Println("Worker " + w.Name_ + " stopped.")
}

package workers

import (
	"github.com/go-emd/emd/log"
	"github.com/go-emd/emd/worker"

	"encoding/binary"
)

var (
	udpFlows *Flows
)

type UdpFlow struct {
	worker.Work
}

func (w UdpFlow) Init() {
	for _, p := range w.Ports() {
		p.Open()
	}

	endOfFlowSeq := []byte{}
	udpFlows = NewFlows(endOfFlowSeq)

	log.INFO.Println("Worker " + w.Name_ + " inited.")
}

func (w UdpFlow) Run() {
	log.INFO.Println("UdpFlow is running.")

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
		case cmd := <-w.Ports()["MGMT_UdpFlow"].Channel():
			if cmd == "STOP" {
				w.Stop()
				return
			} else if cmd == "STATUS" {
				w.Ports()["MGMT_UdpFlow"].Channel() <- "Healthy"
			} else if cmd == "METRICS" {
				w.Ports()["MGMT_UdpFlow"].Channel() <- Metric{
					"partialFlowSize": len(udpFlows.PartialFlows),
					"finalFlowSize": len(udpFlows.FinalFlows),
				}
			}
		case netflow := <-w.Ports()["Sniffer_and_UdpFlow"].Channel():
			srcPort := make([]byte, 2)
			dstPort := make([]byte, 2)

			binary.BigEndian.PutUint16(srcPort, netflow.(Netflow).SrcPort)
			binary.BigEndian.PutUint16(dstPort, netflow.(Netflow).DstPort)

			udpFlows.Update(
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

func (w UdpFlow) Stop() {
	w.Ports()["MGMT_UdpFlow"].Close()
	w.Ports()["Sniffer_and_UdpFlow"].Close()

	log.INFO.Println("Worker " + w.Name_ + " stopped.")
}

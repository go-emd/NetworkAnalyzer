package workers

import (
	"github.com/go-emd/emd/log"
	"github.com/go-emd/emd/worker"

	"encoding/binary"
)

var (
	tcpFlows *Flows
)

type TcpFlow struct {
	worker.Work
}

func (w TcpFlow) Init() {
	for _, p := range w.Ports() {
		p.Open()
	}

	// FIN
	endOfFlowSeq := []byte{0x1}
	tcpFlows = NewFlows(endOfFlowSeq)

	log.INFO.Println("Worker " + w.Name_ + " inited.")
}

func (w TcpFlow) Run() {
	log.INFO.Println("TcpFlow is running.")

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
		case cmd := <-w.Ports()["MGMT_TcpFlow"].Channel():
			if cmd == "STOP" {
				w.Stop()
				return
			} else if cmd == "STATUS" {
				w.Ports()["MGMT_TcpFlow"].Channel() <- "Healthy"
			} else if cmd == "METRICS" {
				w.Ports()["MGMT_TcpFlow"].Channel() <- Metric{
					"partialFlowSize": len(tcpFlows.PartialFlows),
					"finalFlowSize": len(tcpFlows.FinalFlows),
				}

				f, p := tcpFlows.Flush(true)
				log.INFO.Println(f)
				for _, v := range p {
					log.INFO.Println(v.Netflow_)
				}
			}
		case netflow := <-w.Ports()["Sniffer_and_TcpFlow"].Channel():
			srcPort := make([]byte, 2)
			dstPort := make([]byte, 2)

			binary.BigEndian.PutUint16(srcPort, netflow.(Netflow).SrcPort)
			binary.BigEndian.PutUint16(dstPort, netflow.(Netflow).DstPort)

			tcpFlows.Update(
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

func (w TcpFlow) Stop() {
	w.Ports()["MGMT_TcpFlow"].Close()
	w.Ports()["Sniffer_and_TcpFlow"].Close()

	log.INFO.Println("Worker " + w.Name_ + " stopped.")
}

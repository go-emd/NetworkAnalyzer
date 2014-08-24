package workers

import (
	"github.com/go-emd/emd/log"
	"github.com/go-emd/emd/worker"

	//flows "./flows"
)

var (
	//tcpFlows flows.Flows
)

type TcpFlow struct {
	worker.Work
}

func (w TcpFlow) Init() {
	for _, p := range w.Ports() {
		p.Open()
	}

	//endOfFlowSeq := []byte{} // TODO
	//tcpFlows = flows.New(endOfFlowSeq)

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
				w.Ports()["MGMT_TcpFlow"].Channel() <- Metric{"metrics name": "TODO metrics."}
			}
		case netflow := <-w.Ports()["Sniffer_and_TcpFlow"].Channel():
			//tcpFlows.Update([]byte{}, netflow.(Netflow))
			log.INFO.Println(netflow.(Netflow))
		}
	}
}

func (w TcpFlow) Stop() {
	w.Ports()["MGMT_TcpFlow"].Close()
	w.Ports()["Sniffer_and_TcpFlow"].Close()

	log.INFO.Println("Worker " + w.Name_ + " stopped.")
}

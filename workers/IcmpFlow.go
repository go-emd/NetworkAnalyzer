package workers

import (
	"github.com/go-emd/emd/log"
	"github.com/go-emd/emd/worker"

	//flows "./flows"
)

var (
	//icmpFlows flows.Flows
)

type IcmpFlow struct {
	worker.Work
}

func (w IcmpFlow) Init() {
	for _, p := range w.Ports() {
		p.Open()
	}

	//endOfFlowSeq := []byte{} // TODO
	//icmpFlows = flows.New(endOfFlowSeq)

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
		case cmd := <-w.Ports()["MGMT_TcpFlow"].Channel():
			if cmd == "STOP" {
				w.Stop()
				return
			} else if cmd == "STATUS" {
				w.Ports()["MGMT_IcmpFlow"].Channel() <- "Healthy"
			} else if cmd == "METRICS" {
				w.Ports()["MGMT_IcmpFlow"].Channel() <- Metric{"metrics name": "TODO metrics."}
			}
		case data := <-w.Ports()["Sniffer_and_IcmpFlow"].Channel():
			//icmpFlows.Update([]byte{}, netflow.(Netflow))
			log.INFO.Println(data.(Netflow))
		}
	}
}

func (w IcmpFlow) Stop() {
	w.Ports()["MGMT_IcmpFlow"].Close()
	w.Ports()["Sniffer_and_IcmpFlow"].Close()

	log.INFO.Println("Worker " + w.Name_ + " stopped.")
}

package workers

import (
	"github.com/go-emd/emd/log"
	"github.com/go-emd/emd/worker"

	//flows "./flows"
)

var (
	//otherFlows flows.Flows
)

type OtherFlow struct {
	worker.Work
}

func (w OtherFlow) Init() {
	for _, p := range w.Ports() {
		p.Open()
	}

	//endOfFlowSeq := []byte{} // TODO
	//otherFlows = flows.New(endOfFlowSeq)

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
				w.Ports()["MGMT_OtherFlow"].Channel() <- Metric{"metrics name": "TODO metrics."}
			}
		case netflow := <-w.Ports()["Sniffer_and_OtherFlow"].Channel():
			//otherFlows.Update([]byte{}, netflow.(Netflow))
			log.INFO.Println(netflow.(Netflow))
		}
	}
}

func (w OtherFlow) Stop() {
	w.Ports()["MGMT_OtherFlow"].Close()
	w.Ports()["Sniffer_and_OtherFlow"].Close()

	log.INFO.Println("Worker " + w.Name_ + " stopped.")
}

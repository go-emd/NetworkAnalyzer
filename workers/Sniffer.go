package workers

import (
	"github.com/go-emd/emd/log"
	"github.com/go-emd/emd/worker"
)

type Sniffer struct {
	worker.Work
}

func (w Sniffer) Init() {
	for _, p := range w.Ports() {
		p.Open()
	}

	log.INFO.Println("Worker " + w.Name_ + " inited.")
}

func (w Sniffer) Run() {
	log.INFO.Println("Sniffer is running.")

	// Catch any errors that could happen
	defer func() {
		if r := recover(); r != nil {
			log.ERROR.Println("Uncaught error occurred, exiting.")

			w.Stop()
		}
	}()

	w.Ports()["Sniffer_and_Parser"].Channel() <- "Parser this"

	for {
		select {
		case cmd := <-w.Ports()["MGMT_Sniffer"].Channel():
			if cmd == "STOP" {
				w.Stop()
				return
			} else if cmd == "STATUS" {
				w.Ports()["MGMT_Sniffer"].Channel() <- "Healthy"
			} else if cmd == "METRICS" {
				w.Ports()["MGMT_Sniffer"].Channel() <- Metric{"health": "TODO metrics."}
			}
		case data := <-w.Ports()["Sniffer_and_Parser"].Channel():
			log.INFO.Println(data)
		}
	}
}

func (w Sniffer) Stop() {
	w.Ports()["MGMT_Sniffer"].Close()
	w.Ports()["Sniffer_and_Parser"].Close()

	log.INFO.Println("Worker " + w.Name_ + " stopped.")
}

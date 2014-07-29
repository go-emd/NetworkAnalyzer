package workers

import (
	"emd/log"
	"emd/worker"
	"strings"
)

type Parser struct {
	worker.Work
}

func (w Parser) Init() {
	for _, p := range w.Ports() {
		p.Open()
	}

	log.INFO.Println("Worker " + w.Name_ + " inited.")
}

func (w Parser) Run() {
	log.INFO.Println("Parser is running.")

	// Catch any errors that could happen
	defer func() {
		if r := recover(); r != nil {
			log.ERROR.Println("Uncaught error occurred, notifying leader and exiting.")

			w.Stop()
		}
	}()

	for {
		select {
		case cmd := <-w.Ports()["MGMT_Parser"].Channel():
			if cmd == "STOP" {
				w.Stop()
				return
			} else if cmd == "STATUS" {
				w.Ports()["MGMT_Parser"].Channel() <- "Healthy"
			} else if cmd == "METRICS" {
				w.Ports()["MGMT_Parser"].Channel() <- Metric{"health": "TODO metrics."}
			}
		case data := <-w.Ports()["Sniffer_and_Parser"].Channel():
			w.Ports()["Sink_and_Parser"].Channel() <- strings.ToUpper(data.(string))
		}
	}
}

func (w Parser) Stop() {
	w.Ports()["MGMT_Parser"].Close()
	w.Ports()["Sniffer_and_Parser"].Close()
	w.Ports()["Sink_and_Parser"].Close()

	log.INFO.Println("Worker " + w.Name_ + " stopped.")
}

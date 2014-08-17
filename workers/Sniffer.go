package workers

import (
	"github.com/go-emd/emd/log"
	"github.com/go-emd/emd/worker"

	"code.google.com/p/gopacket"
	"code.google.com/p/gopacket/pcap"
)

type Sniffer struct {
	worker.Work
}

var packetSource *gopacket.PacketSource

func (w Sniffer) Init() {
	for _, p := range w.Ports() {
		p.Open()
	}

	if handle, err := pcap.OpenLive("wlan0", 1600, true, 0); err != nil {
		log.ERROR.Println(err)
	} else {
		packetSource = gopacket.NewPacketSource(handle, handle.LinkType())
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
		default:
			packet, err := packetSource.NextPacket()
			if err != nil {
				log.ERROR.Println(err)
			} else {
				w.Ports()["Sniffer_and_Sink"].Channel() <- Metadata{
					packet.LinkLayer().LinkFlow(),
					packet.TransportLayer().TransportFlow(),
					packet.NetworkLayer().NetworkFlow(),
				}
			}
		}
	}
}

func (w Sniffer) Stop() {
	w.Ports()["MGMT_Sniffer"].Close()
	w.Ports()["Sniffer_and_Sink"].Close()

	log.INFO.Println("Worker " + w.Name_ + " stopped.")
}

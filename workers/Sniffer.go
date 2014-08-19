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
			log.ERROR.Println(r)

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
				if err := packet.ErrorLayer(); err != nil {
					log.ERROR.Println(err)
				} else {
					netflow := Netflow{}
					netflow.Start = packet.Metadata().CaptureInfo.Timestamp
					
					
					/*
						ETHERNET HDR: 14bytes
						IP HDR: v4: 20bytes, v6: 36bytes
							- version: 1st byte
					*/
					raw := packet.Data()
					netflow.Ipversion = raw[15] >> 2
				}




				/*if tmp := packet.TransportLayer(); tmp != nil {
					netflow.SrcIp = tmp.TransportFlow().Src()
					netflow.DstIp = tmp.TransportFlow().Dst()
				} else { continue }

				if tmp := packet.NetworkLayer(); tmp != nil {
					netflow.SrcPort = tmp.NetworkFlow().Src()
					netflow.DstPort = tmp.NetworkFlow().Dst()
				} else { continue }

				w.Ports()["Sniffer_and_Sink"].Channel() <- netflow*/
			}
		}
	}
}

func (w Sniffer) Stop() {
	w.Ports()["MGMT_Sniffer"].Close()
	w.Ports()["Sniffer_and_Sink"].Close()

	log.INFO.Println("Worker " + w.Name_ + " stopped.")
}

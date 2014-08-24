package workers

import (
	"github.com/go-emd/emd/log"
	"github.com/go-emd/emd/worker"

	"code.google.com/p/gopacket"
	"code.google.com/p/gopacket/pcap"
	
	"encoding/binary"
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
			// Enforcing sampled for full intake will happen here when configuration 
			// mechanism of each worker is complete.  Defaulting to full intake.
			packet, err := packetSource.NextPacket()
			if err != nil {
				log.ERROR.Println(err)
			} else {
				if err := packet.ErrorLayer(); err != nil {
					log.ERROR.Println(err)
				} else {
					netflow := Netflow{}
					netflow.Start = packet.Metadata().CaptureInfo.Timestamp

					raw := packet.Data()

					netflow.IpVersion = raw[14] >> 4
					netflow.Protocol = raw[14+9]
					netflow.SrcIp = raw[14+12:14+12+4]
					netflow.DstIp = raw[14+16:14+16+4]
					netflow.SrcPort = binary.BigEndian.Uint16(raw[14+20:14+20+2])
					netflow.DstPort = binary.BigEndian.Uint16(raw[14+20+2:14+20+4])
					netflow.Bytes = len(raw)

					switch netflow.Protocol {
					case 1: // ICMP v4
					case 58: // ICMP v6
						w.Ports()["Sniffer_and_IcmpFlow"].Channel() <- netflow
					case 6: // TCP
						w.Ports()["Sniffer_and_TcpFlow"].Channel() <- netflow
					case 17: // UDP
						w.Ports()["Sniffer_and_UdpFlow"].Channel() <- netflow
					default: // Other
						w.Ports()["Sniffer_and_OtherFlow"].Channel() <- netflow
					}
				}
			}
		}
	}
}

func (w Sniffer) Stop() {
	w.Ports()["MGMT_Sniffer"].Close()

	w.Ports()["Sniffer_and_TcpFlow"].Close()
	w.Ports()["Sniffer_and_UdpFlow"].Close()
	w.Ports()["Sniffer_and_IcmpFlow"].Close()
	w.Ports()["Sniffer_and_OtherFlow"].Close()

	log.INFO.Println("Worker " + w.Name_ + " stopped.")
}

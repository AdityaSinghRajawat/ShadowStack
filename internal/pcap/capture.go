//go:build darwin

package pcap

import (
	"fmt"
	"time"

	"shadowStack/internal/protocol"
	"shadowStack/internal/ui"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

// RunPcapCapture listens to macOS network traffic natively
func RunPcapCapture(
	iface string,
	eventsChan chan<- ui.QueryMsg,
	factory *protocol.ParserFactory,
) error {
	handle, err := pcap.OpenLive(iface, 1600, true, pcap.BlockForever)
	if err != nil {
		return fmt.Errorf("failed to open pcap: %v", err)
	}
	defer handle.Close()

	err = handle.SetBPFFilter("tcp port 5432 or tcp port 3306 or tcp port 27017 or tcp port 6379")
	if err != nil {
		return fmt.Errorf("failed to set filter: %v", err)
	}

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	inFlight := make(map[uint16]time.Time)

	for packet := range packetSource.Packets() {
		tcpLayer := packet.Layer(layers.LayerTypeTCP)
		if tcpLayer == nil {
			continue
		}
		tcp := tcpLayer.(*layers.TCP)
		appLayer := packet.ApplicationLayer()
		if appLayer == nil || len(appLayer.Payload()) == 0 {
			continue
		}

		srcPort, dstPort, payload := uint16(tcp.SrcPort), uint16(tcp.DstPort), appLayer.Payload()
		isRequest := factory.GetParser(dstPort) != nil
		isResponse := factory.GetParser(srcPort) != nil

		if isRequest {
			if _, exists := inFlight[srcPort]; exists {
				continue
			}
			parser := factory.GetParser(dstPort)
			if query, err := parser.Parse(payload); err == nil {
				inFlight[srcPort] = time.Now()
				eventsChan <- ui.QueryMsg{Comm: "macOS", DBType: parser.Name(), Query: query, Port: srcPort}
			}
		} else if isResponse {
			if startTime, exists := inFlight[dstPort]; exists {
				latency := time.Since(startTime)
				delete(inFlight, dstPort)
				eventsChan <- ui.QueryMsg{IsUpdate: true, Port: dstPort, Latency: latency}
			}
		}
	}
	return nil
}

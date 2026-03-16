package protocol

import (
	"fmt"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

// DecodedPacket holds the extracted application payload and port info
type DecodedPacket struct {
	SrcPort uint16
	DstPort uint16
	Payload []byte
}

// DecodeNetworkPacket safely strips Ethernet, IP, and TCP headers
func DecodeNetworkPacket(rawBytes []byte) (*DecodedPacket, error) {
	packet := gopacket.NewPacket(rawBytes, layers.LayerTypeEthernet, gopacket.Default)

	tcpLayer := packet.Layer(layers.LayerTypeTCP)
	if tcpLayer == nil {
		return nil, fmt.Errorf("not a TCP packet")
	}

	tcp := tcpLayer.(*layers.TCP)
	appLayer := packet.ApplicationLayer()

	if appLayer == nil || len(appLayer.Payload()) == 0 {
		return nil, fmt.Errorf("no application payload")
	}

	return &DecodedPacket{
		SrcPort: uint16(tcp.SrcPort),
		DstPort: uint16(tcp.DstPort),
		Payload: appLayer.Payload(),
	}, nil
}

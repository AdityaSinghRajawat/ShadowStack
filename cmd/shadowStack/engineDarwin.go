//go:build darwin

package main

import (
	"log"

	"shadowStack/internal/pcap"
	"shadowStack/internal/protocol"
	"shadowStack/internal/ui"
)

// startEngine handles native packet sniffing on macOS via PCAP
func startEngine(eventsChan chan<- ui.QueryMsg) {
	parserFactory := protocol.NewFactory()

	// Run PCAP Engine (Mac Local Development) on 'lo0' loopback
	if err := pcap.RunPcapCapture("lo0", eventsChan, parserFactory); err != nil {
		log.Fatalf("PCAP Capture error (Did you run with sudo?): %v", err)
	}
}

// startProxyEngine initializes the transparent proxy for external databases
func startProxyEngine(localPort, remoteAddr string, eventsChan chan<- ui.QueryMsg) {
	parserFactory := protocol.NewFactory()

	// Run the Proxy implementation
	if err := pcap.RunProxy(localPort, remoteAddr, eventsChan, parserFactory); err != nil {
		log.Fatalf("Proxy Engine error: %v", err)
	}
}

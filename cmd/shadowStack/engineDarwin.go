//go:build darwin

package main

import (
	"log"

	"shadowStack/internal/pcap"
	"shadowStack/internal/protocol"
	"shadowStack/internal/ui"
)

func startEngine(eventsChan chan<- ui.QueryMsg) {
	parserFactory := protocol.NewFactory()

	// Run PCAP Engine (Mac Local Development)
	// On Mac, 'lo0' is the localhost loopback interface
	if err := pcap.RunPcapCapture("lo0", eventsChan, parserFactory); err != nil {
		log.Fatalf("PCAP Capture error (Did you run with sudo?): %v", err)
	}
}

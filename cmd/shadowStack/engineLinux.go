//go:build linux

package main

import (
	"log"
	"net"

	"shadowStack/internal/ebpf"
	"shadowStack/internal/protocol"
	"shadowStack/internal/ui"
)

func startEngine(eventsChan chan<- ui.QueryMsg) {
	// Auto-detect interfaces
	activeIface := getInterface([]string{"lo", "eth0"})
	if activeIface == "" {
		log.Fatal("No active network interface found")
	}

	parserFactory := protocol.NewFactory()

	objs, sockFile, err := ebpf.LoadAndAttach(activeIface)
	if err != nil {
		log.Fatalf("Failed to attach eBPF on %s: %v\n", activeIface, err)
	}
	defer objs.Close()
	defer sockFile.Close()

	if err := ebpf.ReadRingBuf(objs, eventsChan, parserFactory); err != nil {
		log.Printf("RingBuf reader error: %v", err)
	}
}

func getInterface(preferences []string) string {
	interfaces, _ := net.Interfaces()
	for _, pref := range preferences {
		for _, iface := range interfaces {
			if iface.Name == pref && iface.Flags&net.FlagUp != 0 {
				return iface.Name
			}
		}
	}
	return interfaces[0].Name
}

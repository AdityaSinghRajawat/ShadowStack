//go:build linux

package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"shadowStack/internal/ebpf"
	"shadowStack/internal/protocol"
	"shadowStack/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// 1. Auto-detect active interfaces
	interfaces, err := net.Interfaces()
	if err != nil {
		log.Fatalf("Failed to list interfaces: %v", err)
	}

	var activeIface string
	for _, iface := range interfaces {
		// Look for an interface that is UP and has a valid Index
		if iface.Flags&net.FlagUp != 0 && iface.Index > 0 {
			// We prioritize 'lo' for local testing, but 'eth0' is a secondary choice
			if iface.Name == "lo" || iface.Name == "eth0" {
				activeIface = iface.Name
				break
			}
		}
	}

	if activeIface == "" {
		log.Fatal("No active network interface found (tried lo and eth0)")
	}

	// 2. Load and Attach to the detected interface
	objs, sockFile, err := ebpf.LoadAndAttach(activeIface)
	if err != nil {
		log.Fatalf("Failed to attach eBPF on %s: %v\n", activeIface, err)
	}
	defer objs.Close()
	defer sockFile.Close()

	// 3. Normal ShadowStack Pipeline
	eventsChan := make(chan ebpf.ParsedQuery, 100)
	parserFactory := protocol.NewFactory()

	go func() {
		if err := ebpf.ReadRingBuf(objs, eventsChan, parserFactory); err != nil {
			log.Printf("RingBuf reader error: %v", err)
		}
	}()

	p := tea.NewProgram(ui.New())

	go func() {
		for event := range eventsChan {
			p.Send(ui.QueryMsg{
				PID:      event.PID,
				Comm:     event.Comm,
				DBType:   event.DBType,
				Query:    event.Query,
				Port:     event.Port,
				IsUpdate: event.IsUpdate,
				Latency:  event.Latency,
			})
		}
	}()

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error starting UI: %v\n", err)
		os.Exit(1)
	}
}

//go:build linux

package main

import (
	"fmt"
	"log"
	"os"

	"shadowStack/internal/ebpf"
	"shadowStack/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	iface := "lo"
	objs, sockFile, err := ebpf.LoadAndAttach(iface)
	if err != nil {
		log.Fatalf("Failed to attach eBPF: %v\n", err)
	}
	defer objs.Close()
	defer sockFile.Close()

	// 1. Create a channel to pass data from Kernel to UI
	eventsChan := make(chan ebpf.ParsedQuery, 100)

	// 2. Start reading the Ring Buffer in the background
	go func() {
		if err := ebpf.ReadRingBuf(objs, eventsChan); err != nil {
			log.Printf("RingBuf reader error: %v", err)
		}
	}()

	// 3. Initialize Bubble Tea UI
	p := tea.NewProgram(ui.New())

	// 4. Start a background bridge to convert channel events into UI messages
	go func() {
		for event := range eventsChan {
			p.Send(ui.QueryMsg{
				PID:   event.PID,
				Query: event.Query,
			})
		}
	}()

	// 5. Run the UI (This blocks until the user presses 'q')
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error starting UI: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\nDetaching from kernel and exiting...")
}

//go:build linux

package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	// Replace this string with your actual module name from go.mod if it differs
	"shadowStack/internal/ebpf"
)

func main() {
	fmt.Println("ShadowStack 🕵️‍♂️ : Initializing eBPF Tap...")

	// 1. Attach to the loopback interface (localhost traffic)
	iface := "lo"
	objs, sockFile, err := ebpf.LoadAndAttach(iface)
	if err != nil {
		log.Fatalf("Failed to attach eBPF: %v\n(Are you running as root on a Linux kernel?)", err)
	}
	defer objs.Close()
	defer sockFile.Close()

	fmt.Printf("Successfully tapped into interface: %s\n", iface)

	// 2. Setup graceful shutdown (Ctrl+C)
	stopper := make(chan os.Signal, 1)
	signal.Notify(stopper, os.Interrupt, syscall.SIGTERM)

	// 3. Start reading the Ring Buffer in the background
	go func() {
		if err := ebpf.ReadRingBuf(objs); err != nil {
			log.Printf("RingBuf reader error: %v", err)
		}
	}()

	// Wait for the exit signal
	<-stopper
	fmt.Println("\nDetaching from kernel and exiting...")
}

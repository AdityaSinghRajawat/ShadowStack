//go:build linux

package ebpf

import (
	"fmt"
	"net"
	"os"
	"syscall"

	"github.com/cilium/ebpf/link"
)

// LoadAndAttach compiles the BPF program and attaches it to a network interface
func LoadAndAttach(interfaceName string) (*shadowstackObjects, *os.File, error) {
	// 1. Load the compiled eBPF objects into the kernel
	var objs shadowstackObjects
	if err := loadShadowstackObjects(&objs, nil); err != nil {
		return nil, nil, fmt.Errorf("failed to load eBPF objects: %v", err)
	}

	// 2. Find the network interface (e.g., "lo" or "eth0")
	iface, err := net.InterfaceByName(interfaceName)
	if err != nil {
		objs.Close()
		return nil, nil, fmt.Errorf("failed to find network interface %s: %v", interfaceName, err)
	}

	// 3. Open a raw socket on that interface to capture traffic
	sockFd, err := syscall.Socket(
		syscall.AF_PACKET,
		syscall.SOCK_RAW,
		int(htons(syscall.ETH_P_ALL)),
	)
	if err != nil {
		objs.Close()
		return nil, nil, fmt.Errorf("failed to create raw socket: %v", err)
	}

	sll := syscall.SockaddrLinklayer{
		Ifindex:  iface.Index,
		Protocol: htons(syscall.ETH_P_ALL),
	}
	if err := syscall.Bind(sockFd, &sll); err != nil {
		syscall.Close(sockFd)
		objs.Close()
		return nil, nil, fmt.Errorf("failed to bind socket: %v", err)
	}

	// Wrap the raw socket file descriptor into an *os.File
	// This satisfies the syscall.Conn interface required by cilium/ebpf
	sockFile := os.NewFile(uintptr(sockFd), "raw_sock")

	// 4. Attach our eBPF socket_handler to the raw socket
	// Notice: It only returns an error now. Detaching happens when we close sockFile.
	err = link.AttachSocketFilter(sockFile, objs.SocketHandler)
	if err != nil {
		sockFile.Close() // This safely closes sockFd as well
		objs.Close()
		return nil, nil, fmt.Errorf("failed to attach socket filter: %v", err)
	}

	return &objs, sockFile, nil
}

// htons converts a short integer from host byte order to network byte order.
func htons(i uint16) uint16 {
	return (i << 8) | (i >> 8)
}

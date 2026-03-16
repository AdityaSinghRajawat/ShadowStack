//go:build linux

package ebpf

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/cilium/ebpf/ringbuf"

	"shadowStack/internal/protocol"
)

type SQLEvent struct {
	PID        uint32
	PayloadLen uint32
	Payload    [256]byte
}

func ReadRingBuf(objs *shadowstackObjects) error {
	rd, err := ringbuf.NewReader(objs.Events)
	if err != nil {
		return fmt.Errorf("opening ringbuf reader: %v", err)
	}
	defer rd.Close()

	fmt.Println("Listening for network packets... (Press Ctrl+C to stop)")

	var event SQLEvent
	for {
		record, err := rd.Read()
		if err != nil {
			if errors.Is(err, ringbuf.ErrClosed) {
				return nil
			}
			return fmt.Errorf("reading from ringbuf: %v", err)
		}

		if err := binary.Read(bytes.NewBuffer(record.RawSample), binary.LittleEndian, &event); err != nil {
			continue
		}

		// Get the exact slice of data captured by the kernel
		payloadBytes := event.Payload[:event.PayloadLen]

		// Use the Protocol Parser to try and find a SQL query!
		query, err := protocol.ParsePostgresQuery(payloadBytes)
		if err == nil {
			// We found a clean query! Print it out beautifully.
			fmt.Printf("🎯 [PID: %d] Query: %s\n", event.PID, query)
		}
	}
}

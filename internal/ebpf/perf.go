//go:build linux

package ebpf

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"

	"shadowStack/internal/protocol"

	"github.com/cilium/ebpf/ringbuf"
)

type SQLEvent struct {
	PID        uint32
	PayloadLen uint32
	Payload    [256]byte
}

type ParsedQuery struct {
	PID   uint32
	Query string
}

// Accept a channel to send events out
func ReadRingBuf(objs *shadowstackObjects, eventsChan chan<- ParsedQuery) error {
	rd, err := ringbuf.NewReader(objs.Events)
	if err != nil {
		return fmt.Errorf("opening ringbuf reader: %v", err)
	}
	defer rd.Close()

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

		payloadBytes := event.Payload[:event.PayloadLen]

		query, err := protocol.ParsePostgresQuery(payloadBytes)
		if err == nil {
			// Send the clean query straight to the UI channel
			eventsChan <- ParsedQuery{
				PID:   event.PID,
				Query: query,
			}
		}
	}
}

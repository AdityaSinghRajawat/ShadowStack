//go:build linux

package ebpf

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"strings"

	"shadowStack/internal/protocol"

	"github.com/cilium/ebpf/ringbuf"
)

type SQLEvent struct {
	PID        uint32
	PayloadLen uint32
	Payload    [512]byte
}

type ParsedQuery struct {
	PID    uint32
	Comm   string
	DBType string
	Query  string
}

// getProcessName reads the Linux procfs to find the name of the executable
func getProcessName(pid uint32) string {
	path := fmt.Sprintf("/proc/%d/comm", pid)
	data, err := os.ReadFile(path)
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(data))
}

func ReadRingBuf(
	objs *shadowstackObjects,
	eventsChan chan<- ParsedQuery,
	factory *protocol.ParserFactory,
) error {
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

		// Safety check for payload length to avoid slice bounds out of range
		if event.PayloadLen > uint32(len(event.Payload)) {
			event.PayloadLen = uint32(len(event.Payload))
		}
		payloadBytes := event.Payload[:event.PayloadLen]

		// 1. Decode network layers
		decoded, err := protocol.DecodeNetworkPacket(payloadBytes)
		if err != nil || decoded == nil {
			continue
		}

		// 2. Route to the correct parser based on Port
		parser := factory.GetParser(decoded.DstPort)
		if parser == nil {
			parser = factory.GetParser(decoded.SrcPort)
		}

		// 3. Parse the specific wire protocol
		if parser != nil {
			query, err := parser.Parse(decoded.Payload)
			if err == nil {
				eventsChan <- ParsedQuery{
					PID:    event.PID,
					Comm:   getProcessName(event.PID),
					DBType: parser.Name(),
					Query:  query,
				}
			}
		}
	}
}

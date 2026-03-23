//go:build linux

package ebpf

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"shadowStack/internal/protocol"
	"shadowStack/internal/ui"

	"github.com/cilium/ebpf/ringbuf"
)

type SQLEvent struct {
	PID        uint32
	PayloadLen uint32
	Payload    [512]byte
}

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
	eventsChan chan<- ui.QueryMsg,
	factory *protocol.ParserFactory,
) error {
	rd, err := ringbuf.NewReader(objs.Events)
	if err != nil {
		return fmt.Errorf("opening ringbuf reader: %v", err)
	}
	defer rd.Close()

	inFlight := make(map[uint16]time.Time)

	var event SQLEvent
	for {
		record, err := rd.Read()
		if err != nil {
			if errors.Is(err, ringbuf.ErrClosed) {
				return nil
			}
			continue
		}

		if err := binary.Read(bytes.NewBuffer(record.RawSample), binary.LittleEndian, &event); err != nil {
			continue
		}

		if event.PayloadLen > uint32(len(event.Payload)) {
			event.PayloadLen = uint32(len(event.Payload))
		}
		payloadBytes := event.Payload[:event.PayloadLen]

		decoded, err := protocol.DecodeNetworkPacket(payloadBytes)
		if err != nil || decoded == nil {
			continue
		}

		isRequest := factory.GetParser(decoded.DstPort) != nil
		isResponse := factory.GetParser(decoded.SrcPort) != nil

		if isRequest {
			if _, exists := inFlight[decoded.SrcPort]; exists {
				continue
			}

			parser := factory.GetParser(decoded.DstPort)
			query, err := parser.Parse(decoded.Payload)
			if err == nil {
				inFlight[decoded.SrcPort] = time.Now()

				eventsChan <- ui.QueryMsg{
					PID:    event.PID,
					Comm:   getProcessName(event.PID),
					DBType: parser.Name(),
					Query:  query,
					Port:   decoded.SrcPort,
				}
			}
		} else if isResponse {
			clientPort := decoded.DstPort
			if startTime, exists := inFlight[clientPort]; exists {
				latency := time.Since(startTime)
				delete(inFlight, clientPort)

				eventsChan <- ui.QueryMsg{
					IsUpdate: true,
					Port:     clientPort,
					Latency:  latency,
				}
			}
		}
	}
}

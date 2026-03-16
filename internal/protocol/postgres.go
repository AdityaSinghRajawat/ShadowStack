package protocol

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

// ParsePostgresQuery attempts to extract a Postgres SQL query from a raw ethernet frame
func ParsePostgresQuery(rawBytes []byte) (string, error) {
	// 1. Decode the raw packet starting from the Ethernet layer
	packet := gopacket.NewPacket(rawBytes, layers.LayerTypeEthernet, gopacket.Default)

	// 2. Check if it has an Application Layer (Payload)
	appLayer := packet.ApplicationLayer()
	if appLayer == nil {
		return "", fmt.Errorf("no application payload")
	}

	payload := appLayer.Payload()
	if len(payload) == 0 {
		return "", fmt.Errorf("empty payload")
	}

	// 3. Postgres Wire Protocol Logic
	// A Simple Query starts with 'Q' (byte 81), followed by a 4-byte length, then the SQL string.
	if payload[0] == 'Q' {
		// Ensure the payload is long enough to contain the 4-byte length
		if len(payload) < 5 {
			return "", fmt.Errorf("payload too short to be a Postgres query")
		}

		// Extract everything after the 'Q' and the 4-byte length header
		queryBytes := payload[5:]

		// Postgres strings are null-terminated (end with a \x00 byte).
		// We need to trim any null bytes or trailing whitespace.
		query := string(bytes.TrimRight(queryBytes, "\x00"))
		query = strings.TrimSpace(query)

		if query != "" {
			return query, nil
		}
	}

	return "", fmt.Errorf("not a postgres simple query")
}

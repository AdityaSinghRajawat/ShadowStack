package protocol

import (
	"bytes"
	"fmt"
	"strings"
)

type RedisParser struct{}

func (p *RedisParser) Parse(payload []byte) (string, error) {
	if len(payload) < 4 {
		return "", fmt.Errorf("payload too short")
	}

	// Redis commands in RESP usually start with '*' (Arrays)
	// or inline commands for simple telnet-like sessions.
	if payload[0] == '*' || payload[0] == '+' || payload[0] == '$' {
		// Clean up the binary RESP junk to show the readable command
		// We'll extract printable characters to show the command and key
		lines := strings.Split(string(payload), "\r\n")
		var commandParts []string

		for _, line := range lines {
			// Skip RESP metadata (lengths starting with $ or *)
			if len(line) > 0 && line[0] != '$' && line[0] != '*' && line[0] != ':' {
				commandParts = append(commandParts, strings.TrimSpace(line))
			}
		}

		if len(commandParts) > 0 {
			return strings.Join(commandParts, " "), nil
		}
	}

	// Fallback for simple inline commands
	return strings.TrimSpace(string(bytes.TrimRight(payload, "\x00"))), nil
}

func (p *RedisParser) Name() string { return "Redis" }

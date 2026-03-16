package protocol

import (
	"fmt"
	"strings"
)

type MySQLParser struct{}

func (p *MySQLParser) Parse(payload []byte) (string, error) {
	if len(payload) < 5 {
		return "", fmt.Errorf("payload too short")
	}
	// MySQL packet structure: [3 bytes length][1 byte sequence][1 byte command]
	// 0x03 is the COM_QUERY command
	if payload[4] == 0x03 {
		query := string(payload[5:])
		query = strings.TrimSpace(query)
		if query != "" {
			return query, nil
		}
	}
	return "", fmt.Errorf("not a mysql COM_QUERY")
}

func (p *MySQLParser) Name() string { return "MySQL" }

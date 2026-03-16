package protocol

import (
	"bytes"
	"fmt"
	"strings"
)

type PostgresParser struct{}

func (p *PostgresParser) Parse(payload []byte) (string, error) {
	if len(payload) < 5 {
		return "", fmt.Errorf("payload too short")
	}
	// 'Q' indicates a Simple Query
	if payload[0] == 'Q' {
		queryBytes := payload[5:]
		query := string(bytes.TrimRight(queryBytes, "\x00"))
		query = strings.TrimSpace(query)
		if query != "" {
			return query, nil
		}
	}
	return "", fmt.Errorf("not a postgres simple query")
}

func (p *PostgresParser) Name() string { return "PostgreSQL" }

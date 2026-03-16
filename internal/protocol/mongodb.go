package protocol

import (
	"encoding/binary"
	"fmt"
	"regexp"
	"strings"
)

type MongoDBParser struct{}

func (p *MongoDBParser) Parse(payload []byte) (string, error) {
	if len(payload) < 16 {
		return "", fmt.Errorf("payload too short")
	}

	// MongoDB OpCodes: 2004 (OP_QUERY), 2013 (OP_MSG)
	opCode := binary.LittleEndian.Uint32(payload[12:16])
	if opCode != 2004 && opCode != 2013 {
		return "", fmt.Errorf("not a mongo command")
	}

	body := string(payload[16:])
	// Improved regex to catch command, collection, and operators
	re := regexp.MustCompile(`[a-zA-Z0-9_$]{2,}`)
	matches := re.FindAllString(body, -1)

	if len(matches) >= 2 {
		// Format: db.collection.command({args})
		return fmt.Sprintf(
			"db.%s.%s(%s)",
			matches[1],
			matches[0],
			strings.Join(matches[2:], " "),
		), nil
	}
	return "unknown mongo operation", nil
}

func (p *MongoDBParser) Name() string { return "MongoDB" }

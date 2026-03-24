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
	re := regexp.MustCompile(`[a-zA-Z0-9_$]{2,}`)
	matches := re.FindAllString(body, -1)

	if len(matches) >= 2 {
		ignoredCmds := map[string]bool{
			"ismaster": true, "isMaster": true, "hello": true, "helloOk": true,
			"ping": true, "buildInfo": true, "getLog": true,
			"saslStart": true, "saslContinue": true,
		}

		// Scan the first 5 words of the payload.
		// If ANY of them are in our ignore list, drop the query.
		for i := 0; i < len(matches) && i < 5; i++ {
			if ignoredCmds[matches[i]] {
				return "", fmt.Errorf("ignored background noise")
			}
		}

		command := matches[0]
		return fmt.Sprintf(
			"db.%s.%s(%s)",
			matches[1],
			command,
			strings.Join(matches[2:], " "),
		), nil
	}
	return "unknown mongo operation", nil
}

func (p *MongoDBParser) Name() string { return "MongoDB" }

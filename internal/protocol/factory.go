package protocol

// DatabaseParser defines the contract for all wire-protocol parsers
type DatabaseParser interface {
	Parse(payload []byte) (string, error)
	Name() string
}

// ParserFactory manages and routes to different DB parsers
type ParserFactory struct {
	parsers map[uint16]DatabaseParser
}

// NewFactory initializes the supported databases
func NewFactory() *ParserFactory {
	return &ParserFactory{
		parsers: map[uint16]DatabaseParser{
			5432:  &PostgresParser{},
			3306:  &MySQLParser{},
			27017: &MongoDBParser{},
			6379:  &RedisParser{},
		},
	}
}

func (f *ParserFactory) GetParser(port uint16) DatabaseParser {
	return f.parsers[port]
}

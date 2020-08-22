package hits

type (
	CommandType uint32
	EventType   uint32

	CommandUnmarshaller func(cmdType CommandType, data []byte) interface{}
	EventMarshaller     func(eventType EventType, event interface{}) []byte
	Processor           func(cmdType CommandType, cmd interface{}, timestamp uint64) (eventType EventType, event interface{})

	Command struct {
		Type    CommandType
		Data    []byte
		ReplyTo chan<- ReplyValue
	}

	ReplyValue struct {
		Type  EventType
		Event interface{}
	}

	inputElem struct {
		timestamp uint64
		sequence  uint64
		replyTo   chan<- ReplyValue

		cmdType CommandType
		data    []byte
		command interface{}
	}

	outputElem struct {
		timestamp uint64
		sequence  uint64
		replyTo   chan<- ReplyValue

		eventType EventType
		event     interface{}
		data      []byte
	}
)

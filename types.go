package hits

import "errors"

type (
	CommandType uint32
	EventType   uint32

	Command struct {
		Type    CommandType
		Data    []byte
		ReplyTo chan<- Event
	}

	MarshalledEvent struct {
		Type      EventType
		Sequence  uint64
		Timestamp uint64
		Data      []byte
	}

	Event struct {
		Type      EventType
		Sequence  uint64
		Timestamp uint64
		Value     interface{}
	}

	NullMarshalledEvent struct {
		Valid bool
		Event MarshalledEvent
	}
)

var ErrEventsNotFound = errors.New("events from this sequence not found")

type (
	CommandUnmarshaller func(cmdType CommandType, data []byte) interface{}
	EventMarshaller     func(eventType EventType, event interface{}) []byte
	Processor           interface {
		Init() uint64
		Process(cmdType CommandType, cmd interface{},
			timestamp uint64) (eventType EventType, event interface{})
	}
	Journaler interface {
		Store(events []MarshalledEvent)
		ReadFrom(fromSequence uint64) ([]MarshalledEvent, error)
	}
	DBWriter interface {
		Write(events []Event)
	}
)

type (
	inputElem struct {
		timestamp uint64
		sequence  uint64
		replyTo   chan<- Event

		cmdType CommandType
		data    []byte
		command interface{}
	}

	outputElem struct {
		timestamp uint64
		sequence  uint64
		replyTo   chan<- Event

		eventType EventType
		event     interface{}
		data      []byte
	}
)

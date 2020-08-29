package hits

import "errors"

type (
	// CommandType the number specify which command
	CommandType uint32
	// EventType the number specify which event
	EventType uint32

	// Command hold data for each command
	Command struct {
		Type    CommandType
		Value   interface{}
		ReplyTo chan<- Event
	}

	// MarshalledEvent an event that has been marshalled to bytes
	MarshalledEvent struct {
		Type      EventType
		Sequence  uint64
		Timestamp uint64
		Data      []byte
	}

	// Event used for write to DB
	Event struct {
		Type      EventType
		Sequence  uint64
		Timestamp uint64
		Value     interface{}
	}

	// NullMarshalledEvent represent a nullable value
	NullMarshalledEvent struct {
		Valid bool
		Event MarshalledEvent
	}
)

// ErrEventsNotFound Journaler returns when some of expecting events had been deleted
var ErrEventsNotFound = errors.New("events from this sequence not found")

type (
	// EventMarshaller for marshalling events
	EventMarshaller func(eventType EventType, event interface{}) []byte

	// Processor is the core of logic
	Processor interface {
		Init() uint64
		Process(cmdType CommandType, cmd interface{},
			timestamp uint64) (eventType EventType, event interface{})
	}

	// Journaler for event persistent
	Journaler interface {
		Store(events []MarshalledEvent)
		ReadFrom(fromSequence uint64) ([]MarshalledEvent, error)
	}

	// DBWriter for storing the whole data
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

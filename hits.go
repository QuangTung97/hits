package hits

import (
	"hits/sequence"
)

const (
	defaultRingBufferSize = 1 << 16
)

type (
	// event related types
	InputEventType  uint32
	OutputEventType uint32

	ringBufferEvent struct {
		// in milliseconds
		timestamp uint64
		sequence  uint64

		inputType  InputEventType
		inputData  []byte
		inputEvent interface{}

		outputType  OutputEventType
		outputEvent interface{}
		outputData  []byte
	}

	UnmarshalledInputEvent struct {
		Type InputEventType
		Data []byte
	}

	InputEvent struct {
		Type  InputEventType
		Event interface{}
	}

	OutputEvent struct {
		Type  OutputEventType
		Event interface{}
	}

	MarshalledOutputEvent struct {
		Type OutputEventType
		Data []byte
	}

	// interfaces

	InputReplicator interface {
		Replicate(events []UnmarshalledInputEvent) error
	}

	InputUnmarshaller interface {
		Unmarshal(inputType InputEventType, data []byte) (interface{}, error)
	}

	OutputMarshaller interface {
		Marshal(outputType OutputEventType, event interface{}) ([]byte, error)
	}

	// structs

	Config struct {
		RingBufferSize  uint32
		InputReplicator InputReplicator
	}

	Context struct {
		seqCtx     *sequence.Context
		cfg        Config
		ringBuffer []ringBufferEvent
	}
)

func NewContext(cfg Config) *Context {
	if cfg.RingBufferSize == 0 {
		cfg.RingBufferSize = defaultRingBufferSize
	}

	return &Context{
		seqCtx: sequence.NewContext(),
		cfg:    cfg,
	}
}

func (c *Context) Run() {
}

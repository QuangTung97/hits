package hits

import (
	"hits/sequence"
	"sync"
	"time"
)

type (
	Config struct {
		RingBufferShift     uint8
		CommandUnmarshaller CommandUnmarshaller
		EventMarshaller     EventMarshaller
		Processor           Processor
	}

	sequencers struct {
		producer     sequence.Sequencer
		unmarshaller sequence.Sequencer
		processor    sequence.Sequencer
		marshaller   sequence.Sequencer
		journaler    sequence.Sequencer
		dbWriter     sequence.Sequencer
		eventEmitter sequence.Sequencer
		replier      sequence.Sequencer
	}

	sequenceBarriers struct {
		producer     sequence.Barrier
		unmarshaller sequence.Barrier
		processor    sequence.Barrier
		marshaller   sequence.Barrier
		journaler    sequence.Barrier
		dbWriter     sequence.Barrier
		eventEmitter sequence.Barrier
		replier      sequence.Barrier
	}

	callbacks struct {
		commandUnmarshaller CommandUnmarshaller
		processor           Processor
	}

	WaitStrategies struct {
		Producer     sequence.WaitStrategy
		Unmarshaller sequence.WaitStrategy
		Processor    sequence.WaitStrategy
		Marshaller   sequence.WaitStrategy
		Journaler    sequence.WaitStrategy
		DBWriter     sequence.WaitStrategy
		EventEmitter sequence.WaitStrategy
		Replier      sequence.WaitStrategy
	}

	Context struct {
		seqCtx       *sequence.Context
		bufferMask   uint64
		bufferSize   uint64
		inputBuffer  []inputElem
		outputBuffer []outputElem
		seqs         sequencers
		barriers     sequenceBarriers
		strats       WaitStrategies
		callbacks    callbacks
	}
)

func shiftToMask(shift uint8) uint64 {
	return (1 << shift) - 1
}

func newSequencers(ctx *sequence.Context) sequencers {
	return sequencers{
		producer:     ctx.NewSequencer(),
		unmarshaller: ctx.NewSequencer(),
		processor:    ctx.NewSequencer(),
		marshaller:   ctx.NewSequencer(),
		journaler:    ctx.NewSequencer(),
		dbWriter:     ctx.NewSequencer(),
		eventEmitter: ctx.NewSequencer(),
		replier:      ctx.NewSequencer(),
	}
}

func newBarriers(seqs sequencers) sequenceBarriers {
	return sequenceBarriers{
		producer:     sequence.NewBarrier(seqs.eventEmitter, seqs.replier),
		unmarshaller: sequence.NewBarrier(seqs.producer),
		processor:    sequence.NewBarrier(seqs.unmarshaller),
		marshaller:   sequence.NewBarrier(seqs.processor),
		journaler:    sequence.NewBarrier(seqs.marshaller),
		dbWriter:     sequence.NewBarrier(seqs.journaler),
		eventEmitter: sequence.NewBarrier(seqs.dbWriter),
		replier:      sequence.NewBarrier(seqs.dbWriter),
	}
}

func DefaultWaitStrategies() WaitStrategies {
	sleepWait := sequence.SleepWaitStrategy{Duration: 10 * time.Microsecond}
	return WaitStrategies{
		Producer:     sleepWait,
		Unmarshaller: sleepWait,
		Processor:    sleepWait,
		Marshaller:   sleepWait,
		Journaler:    sleepWait,
		DBWriter:     sleepWait,
		EventEmitter: sleepWait,
		Replier:      sleepWait,
	}
}

func NewContext(cfg Config) *Context {
	mask := shiftToMask(cfg.RingBufferShift)
	size := uint64(1 << cfg.RingBufferShift)

	inputBuffer := make([]inputElem, size)
	outputBuffer := make([]outputElem, size)

	seqCtx := sequence.NewContext()

	seqs := newSequencers(seqCtx)
	barriers := newBarriers(seqs)

	callbacks := callbacks{}

	if cfg.CommandUnmarshaller == nil {
		panic("CommandUnmarshaller must not be nil")
	}
	callbacks.commandUnmarshaller = cfg.CommandUnmarshaller

	if cfg.Processor == nil {
		panic("Processor must not be nil")
	}
	callbacks.processor = cfg.Processor

	return &Context{
		seqCtx:       seqCtx,
		bufferMask:   mask,
		bufferSize:   size,
		inputBuffer:  inputBuffer,
		outputBuffer: outputBuffer,
		seqs:         seqs,
		barriers:     barriers,
		strats:       DefaultWaitStrategies(),
		callbacks:    callbacks,
	}
}

func (c *Context) getInput(sequence uint64) *inputElem {
	return &c.inputBuffer[sequence&c.bufferMask]
}

func (c *Context) getOutput(sequence uint64) *outputElem {
	return &c.outputBuffer[sequence&c.bufferMask]
}

func (c *Context) Run(cmdChan <-chan Command) {
	var wg sync.WaitGroup
	wg.Add(5)

	initSequence := uint64(0)

	go c.runProducer(&wg, cmdChan, initSequence)
	go c.runUnmarshaller(&wg, initSequence)
	go c.runProcessor(&wg, initSequence)

	wg.Wait()
}

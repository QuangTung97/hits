package hits

import (
	"github.com/QuangTung97/hits/sequence"
	"google.golang.org/grpc"
)

type (
	// Config for configure HITS
	Config struct {
		RingBufferShift uint8
		EventMarshaller EventMarshaller
		Processor       Processor
		Journaler       Journaler
		DBWriter        DBWriter
		ObserverAddress string
		GRPCOptions     []grpc.ServerOption
	}

	sequencers struct {
		producer     sequence.Sequencer
		processor    sequence.Sequencer
		marshaller   sequence.Sequencer
		journaler    sequence.Sequencer
		dbWriter     sequence.Sequencer
		eventEmitter sequence.Sequencer
		replier      sequence.Sequencer
	}

	sequenceBarriers struct {
		producer     sequence.Barrier
		processor    sequence.Barrier
		marshaller   sequence.Barrier
		journaler    sequence.Barrier
		dbWriter     sequence.Barrier
		eventEmitter sequence.Barrier
		replier      sequence.Barrier
	}

	callbacks struct {
		processor       Processor
		eventMarshaller EventMarshaller
		journaler       Journaler
		dbWriter        DBWriter
	}

	// WaitStrategies configure wait strategies
	WaitStrategies struct {
		Producer     sequence.WaitStrategy
		Processor    sequence.WaitStrategy
		Marshaller   sequence.WaitStrategy
		Journaler    sequence.WaitStrategy
		DBWriter     sequence.WaitStrategy
		EventEmitter sequence.WaitStrategy
		Replier      sequence.WaitStrategy
	}

	// Context is the context of HITS
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
		observer     *observerService
		observerAddr string
		gRPCOptions  []grpc.ServerOption
	}
)

func shiftToMask(shift uint8) uint64 {
	return (1 << shift) - 1
}

func newSequencers(ctx *sequence.Context) sequencers {
	return sequencers{
		producer:     ctx.NewSequencer(),
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
		processor:    sequence.NewBarrier(seqs.producer),
		marshaller:   sequence.NewBarrier(seqs.processor),
		journaler:    sequence.NewBarrier(seqs.marshaller),
		dbWriter:     sequence.NewBarrier(seqs.journaler),
		eventEmitter: sequence.NewBarrier(seqs.dbWriter),
		replier:      sequence.NewBarrier(seqs.dbWriter),
	}
}

// DefaultWaitStrategies returns the default strategies
func DefaultWaitStrategies() WaitStrategies {
	return WaitStrategies{
		Producer:     sequence.NewBlockingWaitStrategy(),
		Processor:    sequence.NewBlockingWaitStrategy(),
		Marshaller:   sequence.NewBlockingWaitStrategy(),
		Journaler:    sequence.NewBlockingWaitStrategy(),
		DBWriter:     sequence.NewBlockingWaitStrategy(),
		EventEmitter: sequence.NewBlockingWaitStrategy(),
		Replier:      sequence.NewBlockingWaitStrategy(),
	}
}

// NewContext returns a new context
func NewContext(cfg Config) *Context {
	mask := shiftToMask(cfg.RingBufferShift)
	size := uint64(1 << cfg.RingBufferShift)

	inputBuffer := make([]inputElem, size)
	outputBuffer := make([]outputElem, size)

	seqCtx := sequence.NewContext()

	seqs := newSequencers(seqCtx)
	barriers := newBarriers(seqs)

	callbacks := callbacks{}

	if cfg.Processor == nil {
		panic("Processor must not be nil")
	}
	callbacks.processor = cfg.Processor

	if cfg.EventMarshaller == nil {
		panic("EventMarshaller must not be nil")
	}
	callbacks.eventMarshaller = cfg.EventMarshaller

	if cfg.Journaler == nil {
		panic("Journaler must not be nil")
	}
	callbacks.journaler = cfg.Journaler

	if cfg.DBWriter == nil {
		panic("DBWriter must not be nil")
	}
	callbacks.dbWriter = cfg.DBWriter

	observerAddr := cfg.ObserverAddress
	if observerAddr == "" {
		panic("ObserverAddress must not be empty")
	}

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
		observer:     newObserverService(),
		observerAddr: cfg.ObserverAddress,
		gRPCOptions:  cfg.GRPCOptions,
	}
}

func (c *Context) getInput(sequence uint64) *inputElem {
	return &c.inputBuffer[sequence&c.bufferMask]
}

func (c *Context) getOutput(sequence uint64) *outputElem {
	return &c.outputBuffer[sequence&c.bufferMask]
}

func (c *Context) initSequencers(initSequence uint64) {
	c.seqCtx.Commit(c.seqs.producer, initSequence)
	c.seqCtx.Commit(c.seqs.processor, initSequence)
	c.seqCtx.Commit(c.seqs.marshaller, initSequence)
	c.seqCtx.Commit(c.seqs.journaler, initSequence)
	c.seqCtx.Commit(c.seqs.dbWriter, initSequence)
	c.seqCtx.Commit(c.seqs.eventEmitter, initSequence)
	c.seqCtx.Commit(c.seqs.replier, initSequence)
}

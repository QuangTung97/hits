package sequence

import (
	"sync/atomic"
	"time"
)

type (
	atomicSequence struct {
		value uint64
		// fit into 128 bytes (about 2 cache lines)
		// to prevent "false sharing"
		_ [120]byte
	}

	// Context for barriers
	Context struct {
		atomicSequences []atomicSequence
	}

	// Sequencer represent a point on the disruptor
	Sequencer struct {
		offset uint32
	}

	// Barrier is a set of Sequencer for waiting
	Barrier struct {
		offsets []uint32
	}

	// WaitStrategy on barrier
	WaitStrategy interface {
		Wait()
	}

	// SleepWaitStrategy using sleep
	SleepWaitStrategy struct {
		Duration time.Duration
	}

	// BusySpinWaitStrategy using busy loop
	BusySpinWaitStrategy struct {
		Loop uint64
	}
)

// NewContext create a context
func NewContext() *Context {
	atomicSequences := make([]atomicSequence, 0)

	return &Context{
		atomicSequences: atomicSequences,
	}
}

// NewSequencer create a Sequencer
func (c *Context) NewSequencer() Sequencer {
	offset := len(c.atomicSequences)
	c.atomicSequences = append(c.atomicSequences, atomicSequence{value: 0})
	return Sequencer{
		offset: uint32(offset),
	}
}

// NewBarrier create a Barrier
func NewBarrier(seqs []Sequencer) Barrier {
	if len(seqs) == 0 {
		panic("Barrier should contain at least one sequence")
	}

	offsets := make([]uint32, 0, len(seqs))
	for _, seq := range seqs {
		offsets = append(offsets, seq.offset)
	}
	return Barrier{
		offsets: offsets,
	}
}

// WaitFor wait for a barrier
func (c *Context) WaitFor(b Barrier, expectedSequence uint64, waitStrategy WaitStrategy) uint64 {
	for {
		minSequence := atomic.LoadUint64(&c.atomicSequences[0].value)
		for _, offset := range b.offsets {
			seq := atomic.LoadUint64(&c.atomicSequences[offset].value)
			if seq < minSequence {
				minSequence = seq
			}
		}
		if minSequence >= expectedSequence {
			return minSequence
		}
		waitStrategy.Wait()
	}
}

// Commit write the sequence to the Sequencer
func (c *Context) Commit(seq Sequencer, value uint64) {
	atomic.StoreUint64(&c.atomicSequences[seq.offset].value, value)
}

// WaitStrategy definitions

// Wait for sleep strategy
func (s SleepWaitStrategy) Wait() {
	time.Sleep(s.Duration)
}

var fakeCount uint = 0

// Wait for busy spin strategy
func (s BusySpinWaitStrategy) Wait() {
	for i := uint64(0); i < s.Loop; i++ {
		fakeCount++
	}
}

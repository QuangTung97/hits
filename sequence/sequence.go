package sequence

import (
	"runtime"
	"sync"
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
		NotifyAll()
	}

	// BlockingWaitStrategy using mutex and condition variable
	BlockingWaitStrategy struct {
		mut  sync.Mutex
		cond *sync.Cond
	}

	// SleepWaitStrategy using sleep
	SleepWaitStrategy struct {
		Duration time.Duration
	}

	// YieldingWaitStrategy using yielding
	YieldingWaitStrategy struct {
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
func NewBarrier(seqs ...Sequencer) Barrier {
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
		first := b.offsets[0]
		minSequence := atomic.LoadUint64(&c.atomicSequences[first].value)

		for _, offset := range b.offsets[1:] {
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

//-------------------------------
// WaitStrategy definitions
//-------------------------------

// NewBlockingWaitStrategy create a blocking wait strategy
func NewBlockingWaitStrategy() *BlockingWaitStrategy {
	s := &BlockingWaitStrategy{}
	s.cond = sync.NewCond(&s.mut)
	return s
}

// Wait for blocking strategy
func (s *BlockingWaitStrategy) Wait() {
	s.mut.Lock()
	s.cond.Wait()
	s.mut.Unlock()
}

// NotifyAll notify all waiting goroutines
func (s *BlockingWaitStrategy) NotifyAll() {
	s.cond.Broadcast()
}

// Wait for sleep strategy
func (s SleepWaitStrategy) Wait() {
	time.Sleep(s.Duration)
}

// NotifyAll notify all waiting goroutines
func (s SleepWaitStrategy) NotifyAll() {
}

// Wait for yielding strategy
func (s YieldingWaitStrategy) Wait() {
	runtime.Gosched()
}

// NotifyAll notify all waiting goroutines
func (s YieldingWaitStrategy) NotifyAll() {
}

package hits

import (
	"log"
	"sync"
	"time"
)

func getNowUnixMilli() uint64 {
	return uint64(time.Now().UnixNano()) / 1000000
}

func (c *Context) runProducer(wg *sync.WaitGroup, cmdChan <-chan Command, initSequence uint64) {
	defer wg.Done()

	sequence := initSequence
	for {
		cmd := <-cmdChan

		if sequence >= c.bufferSize {
			c.seqCtx.WaitFor(c.barriers.producer, sequence+1-c.bufferSize, c.strats.Producer)
		}
		sequence++
		input := c.getInput(sequence)

		input.cmdType = cmd.Type
		input.data = cmd.Data
		input.replyTo = cmd.ReplyTo
		input.sequence = sequence
		input.timestamp = getNowUnixMilli()

		c.seqCtx.Commit(c.seqs.producer, sequence)
	}
}

func (c *Context) runUnmarshaller(wg *sync.WaitGroup, initSequence uint64) {
	defer wg.Done()

	sequence := initSequence
	for {
		c.seqCtx.WaitFor(c.barriers.unmarshaller, sequence+1, c.strats.Unmarshaller)
		sequence++

		input := c.getInput(sequence)
		input.command = c.callbacks.commandUnmarshaller(input.cmdType, input.data)
		input.data = nil

		log.Printf("Unmarshaller: %d, %+v\n", sequence, input)

		c.seqCtx.Commit(c.seqs.unmarshaller, sequence)
	}
}

func (c *Context) runProcessor(wg *sync.WaitGroup, initSequence uint64) {
	defer wg.Done()

	sequence := initSequence
	for {
		c.seqCtx.WaitFor(c.barriers.processor, sequence+1, c.strats.Processor)
		sequence++

		input := c.getInput(sequence)
		eventType, event := c.callbacks.processor(input.cmdType, input.command, input.timestamp)

		output := c.getOutput(sequence)
		output.timestamp = input.timestamp
		output.sequence = sequence
		output.replyTo = input.replyTo
		output.eventType = eventType
		output.event = event

		input.replyTo = nil
		input.command = nil

		c.seqCtx.Commit(c.seqs.processor, sequence)
	}
}


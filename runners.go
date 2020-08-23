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
		eventType, event := c.callbacks.processor.Process(
			input.cmdType, input.command, input.timestamp)

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

func (c *Context) runMarshaller(wg *sync.WaitGroup, initSequence uint64) {
	defer wg.Done()

	sequence := initSequence
	for {
		c.seqCtx.WaitFor(c.barriers.marshaller, sequence+1, c.strats.Marshaller)
		sequence++

		output := c.getOutput(sequence)
		output.data = c.callbacks.eventMarshaller(output.eventType, output.event)

		c.seqCtx.Commit(c.seqs.marshaller, sequence)
	}
}

func (c *Context) runJournaler(wg *sync.WaitGroup, initSequence uint64) {
	defer wg.Done()

	sequence := initSequence
	events := make([]MarshalledEvent, 0, 128)
	for {
		newSeq := c.seqCtx.WaitFor(c.barriers.journaler, sequence+1, c.strats.Journaler)

		log.Println("NEW SEQ", newSeq)

		for i := sequence + 1; i <= newSeq; i++ {
			output := c.getOutput(i)
			events = append(events, MarshalledEvent{
				Type:      output.eventType,
				Sequence:  output.sequence,
				Timestamp: output.timestamp,
				Data:      output.data,
			})
		}
		c.callbacks.journaler.Store(events)

		// clean up events
		for i := range events {
			events[i].Data = nil
		}
		events = events[:0]

		sequence = newSeq
		c.seqCtx.Commit(c.seqs.journaler, sequence)
	}
}

func (c *Context) runDBWriter(wg *sync.WaitGroup, initSequence uint64) {
	defer wg.Done()

	sequence := initSequence
	events := make([]Event, 0, 128)
	for {
		newSeq := c.seqCtx.WaitFor(c.barriers.dbWriter, sequence+1, c.strats.DBWriter)
		for i := sequence + 1; i <= newSeq; i++ {
			output := c.getOutput(i)
			events = append(events, Event{
				Type:      output.eventType,
				Sequence:  output.sequence,
				Timestamp: output.timestamp,
				Value:     output.event,
			})
		}
		c.callbacks.dbWriter.Write(events)

		// clean up events
		for i := range events {
			events[i].Value = nil
		}
		events = events[:0]

		sequence = newSeq
		c.seqCtx.Commit(c.seqs.dbWriter, sequence)
	}
}

func (c *Context) runEventEmitter(wg *sync.WaitGroup, initSequence uint64) {
	defer wg.Done()

	sequence := initSequence
	events := make([]MarshalledEvent, 0, 128)
	for {
		newSeq := c.seqCtx.WaitFor(c.barriers.eventEmitter, sequence+1, c.strats.EventEmitter)
		for i := sequence + 1; i <= newSeq; i++ {
			output := c.getOutput(i)
			events = append(events, MarshalledEvent{
				Type:      output.eventType,
				Sequence:  output.sequence,
				Timestamp: output.timestamp,
				Data:      output.data,
			})
		}

		log.Println("EMITTER", events)

		// clean up events
		for i := range events {
			events[i].Data = nil
		}
		events = events[:0]

		sequence = newSeq
		c.seqCtx.Commit(c.seqs.eventEmitter, sequence)
	}
}

func (c *Context) runReplier(wg *sync.WaitGroup, initSequence uint64) {
	defer wg.Done()

	sequence := initSequence
	for {
		newSeq := c.seqCtx.WaitFor(c.barriers.replier, sequence+1, c.strats.Replier)
		for i := sequence + 1; i <= newSeq; i++ {
			output := c.getOutput(i)
			event := Event{
				Type:      output.eventType,
				Sequence:  output.sequence,
				Timestamp: output.timestamp,
				Value:     output.event,
			}
			log.Println("REPLY Begin", event)
			output.replyTo <- event
			log.Println("REPLY End", event)

			output.replyTo = nil
		}

		sequence = newSeq
		c.seqCtx.Commit(c.seqs.replier, sequence)
	}
}

func (c *Context) Run(cmdChan <-chan Command) {
	var wg sync.WaitGroup
	wg.Add(7)

	initSequence := c.callbacks.processor.Init()

	c.initSequencers(initSequence)

	go c.runProducer(&wg, cmdChan, initSequence)
	go c.runUnmarshaller(&wg, initSequence)
	go c.runProcessor(&wg, initSequence)
	go c.runMarshaller(&wg, initSequence)
	go c.runJournaler(&wg, initSequence)
	go c.runDBWriter(&wg, initSequence)
	go c.runEventEmitter(&wg, initSequence)
	go c.runReplier(&wg, initSequence)

	wg.Wait()
}

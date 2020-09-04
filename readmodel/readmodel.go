package readmodel

import (
	"context"
	"errors"
	"io"

	"github.com/QuangTung97/hits"
)

type ReadModel interface {
	InitFromOriginalDB() uint64
	Process(event hits.MarshalledEvent)
}

type listenEvent struct {
	event hits.MarshalledEvent
	err   error
}

func Listen(ctx context.Context, address string, initSequence uint64, readmodel ReadModel) {
	ch := make(chan listenEvent, 1<<16)

	go func() {
		err := hits.Listen(ctx, address, func(event hits.MarshalledEvent) {
			ch <- listenEvent{
				event: event,
				err:   nil,
			}
		})
		ch <- listenEvent{
			err: err,
		}
	}()

	var processedSequence uint64
	var unprocessedEvents []hits.MarshalledEvent

	if initSequence == 0 {
		unprocessedEvents, processedSequence = handleWhenInitialize(ctx, ch, readmodel)
	} else {
		unprocessedEvents, processedSequence = handleWhenHaveExistingEvents(
			ctx, address, initSequence, ch, readmodel)
	}

	for _, event := range unprocessedEvents {
		if event.Sequence <= processedSequence {
			continue
		}
		readmodel.Process(event)
		processedSequence = event.Sequence
	}
}

func handleWhenInitialize(
	ctx context.Context, ch <-chan listenEvent,
	readmodel ReadModel,
) ([]hits.MarshalledEvent, uint64) {
	unprocessedEvents := make([]hits.MarshalledEvent, 0, 1<<16)

	seqChan := make(chan uint64)
	go func() {
		seq := readmodel.InitFromOriginalDB()
		seqChan <- seq
	}()

	for {
		select {
		case e := <-ch:
			if e.err != nil {
				panic(e.err)
			}
			unprocessedEvents = append(unprocessedEvents, e.event)

		case seq := <-seqChan:
			return unprocessedEvents, seq
		}
	}
}

func handleWhenHaveExistingEvents(
	ctx context.Context, address string, initSequence uint64,
	ch <-chan listenEvent, readmodel ReadModel,
) ([]hits.MarshalledEvent, uint64) {
	unprocessedEvents := make([]hits.MarshalledEvent, 0, 1<<16)

	type recvEvent struct {
		event hits.MarshalledEvent
		err   error
	}

	initChan := make(chan recvEvent, 1<<16)
	go func() {
		err := hits.ReadEventsFrom(ctx, address, initSequence, func(event hits.MarshalledEvent) {
			initChan <- recvEvent{
				event: event,
				err:   nil,
			}
		})
		initChan <- recvEvent{err: err}
	}()

	lastSequence := initSequence
	for {
		select {
		case e := <-ch:
			if e.err != nil {
				panic(e.err)
			}
			unprocessedEvents = append(unprocessedEvents, e.event)

		case e := <-initChan:
			if e.err != nil {
				if errors.Is(e.err, io.EOF) {
					return unprocessedEvents, lastSequence
				}
				panic(e.err)
			}
			readmodel.Process(e.event)
			lastSequence = e.event.Sequence
		}
	}

}

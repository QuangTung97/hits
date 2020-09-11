package main

import (
	"log"
	"time"

	"github.com/QuangTung97/hits"
)

type processor struct {
}

type dbJournaler struct {
}

type dbWriter struct {
}

func (p *processor) Process(
	cmdType hits.CommandType, cmd interface{}, timestamp uint64,
) (eventType hits.EventType, event interface{}) {
	log.Println("PROC:", cmdType, cmd, timestamp)
	time.Sleep(5 * time.Second)
	return hits.EventType(cmdType + 10), cmd.(string) + " tung"
}

func (p *processor) Init() uint64 {
	return 0
}

func (db *dbJournaler) Store(events []hits.MarshalledEvent) {
	log.Println("JOURNAL", events)
}

func (db *dbJournaler) ReadFrom(fromSequence uint64) ([]hits.MarshalledEvent, error) {
	log.Println("ReadFrom")
	res := make([]hits.MarshalledEvent, 0)
	res = append(res, hits.MarshalledEvent{
		Type:      1,
		Sequence:  1,
		Timestamp: 12354,
		Data:      nil,
	})
	res = append(res, hits.MarshalledEvent{
		Type:      1,
		Sequence:  2,
		Timestamp: 43567,
		Data:      nil,
	})

	return res, nil
}

func (db *dbWriter) Write(events []hits.Event) {
	log.Println("DBWriter", events)
}

func sendCommands(ch chan<- hits.Command) {
	replyChan := make(chan hits.Event, 9)
	for {
		ch <- hits.Command{
			Type:    1,
			Value:   "abcd",
			ReplyTo: replyChan,
		}
	}
}

func main() {
	p := processor{}
	db := dbJournaler{}
	w := dbWriter{}

	cfg := hits.Config{
		RingBufferShift: 2,
		Processor:       &p,
		EventMarshaller: func(eventType hits.EventType, event interface{}) []byte {
			log.Println("MARSHAL", eventType, event)
			return []byte(event.(string) + " marshalled")
		},
		Journaler:       &db,
		DBWriter:        &w,
		ObserverAddress: ":5000",
	}
	hitsCtx := hits.NewContext(cfg)

	cmdChan := make(chan hits.Command, 100)
	go sendCommands(cmdChan)
	hitsCtx.Run(cmdChan)
}

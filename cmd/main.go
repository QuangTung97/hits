package main

import (
	"hits"
	"log"
	"time"
)

type Processor struct {
}

type DBJournaler struct {
}

type DBWriter struct {
}

func (p *Processor) Process(
	cmdType hits.CommandType, cmd interface{}, timestamp uint64,
) (eventType hits.EventType, event interface{}) {
	log.Println("PROC:", cmdType, cmd, timestamp)
	return hits.EventType(cmdType + 10), cmd.(string) + " tung"
}

func (p *Processor) Init() uint64 {
	return 1
}

func (db *DBJournaler) Store(events []hits.MarshalledEvent) {
	log.Println("JOURNAL", events)
}

func (db *DBJournaler) ReadFrom(fromSequence uint64) ([]hits.MarshalledEvent, error) {
	res := make([]hits.MarshalledEvent, 0)
	return res, hits.ErrEventsNotFound
}

func (db *DBWriter) Write(events []hits.Event) {
	log.Println("DBWriter", events)
}

func sendCommands(ch chan<- hits.Command) {
	replyChan := make(chan hits.Event, 9)
	for {
		ch <- hits.Command{
			Type:    1,
			Data:    []byte{'a', 'b', 'c'},
			ReplyTo: replyChan,
		}
		time.Sleep(10 * time.Second)
	}
}

func main() {
	p := Processor{}
	db := DBJournaler{}
	w := DBWriter{}

	cfg := hits.Config{
		RingBufferShift: 2,
		CommandUnmarshaller: func(cmdType hits.CommandType, data []byte) interface{} {
			return string(data)
		},
		Processor: &p,
		EventMarshaller: func(eventType hits.EventType, event interface{}) []byte {
			log.Println("MARSHAL", eventType, event)
			return []byte(event.(string) + " marshalled")
		},
		Journaler: &db,
		DBWriter:  &w,
	}
	hitsCtx := hits.NewContext(cfg)

	cmdChan := make(chan hits.Command, 100)
	go sendCommands(cmdChan)
	hitsCtx.Run(cmdChan)
}

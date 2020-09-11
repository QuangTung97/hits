package main

import (
	"bufio"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/QuangTung97/hits"
)

type processor struct {
	count uint64
}

type dbtest struct {
	f    *os.File
	file *bufio.Writer
}

func newDBtest() *dbtest {
	f, err := os.OpenFile("events", os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		panic(err)
	}

	file := bufio.NewWriter(f)
	return &dbtest{
		f:    f,
		file: file,
	}
}

func (p *processor) Process(
	cmdType hits.CommandType,
	command interface{}, timestamp uint64,
) (hits.EventType, interface{}) {
	p.count++
	return hits.EventType(cmdType), strconv.FormatInt(int64(p.count), 10)
}

func (p *processor) Init() uint64 {
	return 0
}

func (db *dbtest) Store(events []hits.MarshalledEvent) {
	for _, e := range events {
		_, err := db.file.Write(e.Data)
		if err != nil {
			panic(err)
		}
	}
	err := db.file.Flush()
	if err != nil {
		panic(err)
	}

	err = db.f.Sync()
	if err != nil {
		panic(err)
	}
}

func (db *dbtest) ReadFrom(sequence uint64) ([]hits.MarshalledEvent, error) {
	return nil, hits.ErrEventsNotFound
}

func (db *dbtest) Write(events []hits.Event) {
	log.Println("WRITE", len(events))
}

func eventMarshaller(eventType hits.EventType, event interface{}) []byte {
	s := event.(string)
	s += "\n"
	return []byte(s)
}

func main() {
	p := &processor{
		count: 10,
	}
	db := newDBtest()

	cfg := hits.Config{
		RingBufferShift: 17,
		EventMarshaller: eventMarshaller,
		Processor:       p,
		Journaler:       db,
		DBWriter:        db,
		ObserverAddress: ":5000",
	}
	ctx := hits.NewContext(cfg)

	cmdChan := make(chan hits.Command, 1<<cfg.RingBufferShift)
	go ctx.Run(cmdChan)

	const numLoop = 1000000
	replyChan := make(chan hits.Event, 1<<cfg.RingBufferShift)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < numLoop; i++ {
			<-replyChan
		}
	}()

	begin := time.Now()
	for i := 0; i < numLoop; i++ {
		cmdChan <- hits.Command{
			Type:    1,
			Value:   nil,
			ReplyTo: replyChan,
		}
	}

	wg.Wait()
	d := time.Now().Sub(begin)
	log.Println(d.Nanoseconds())
}

package main

import (
	"hits"
	"log"
)

func main() {
	cfg := hits.Config{
		RingBufferShift: 4,
		CommandUnmarshaller: func(cmdType hits.CommandType, data []byte) interface{} {
			return string(data)
		},
		Processor: func(cmdType hits.CommandType, cmd interface{}, timestamp uint64) (eventType hits.EventType, event interface{}) {
			log.Println("PROC", cmdType, cmd, timestamp)
			return 1, cmd.(string) + " tung"
		},
	}
	hitsCtx := hits.NewContext(cfg)

	cmdChan := make(chan hits.Command, 100)

	replyChan := make(chan hits.ReplyValue)
	cmdChan <- hits.Command{
		Type:    1,
		Data:    []byte{'a', 'b', 'c'},
		ReplyTo: replyChan,
	}
	hitsCtx.Run(cmdChan)
}

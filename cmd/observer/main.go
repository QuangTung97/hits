package main

import (
	"context"
	"fmt"
	"github.com/QuangTung97/hits"
)

func main() {
	ch := make(chan hits.MarshalledEvent, 1024)
	go func() {
		err := hits.Listen(context.Background(), ":5000", ch)
		if err != nil {
			panic(err)
		}
	}()

	for {
		e := <-ch
		fmt.Println(e.Sequence, e.Timestamp, e.Data)
	}
}

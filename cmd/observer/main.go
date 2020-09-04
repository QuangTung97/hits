package main

import (
	"context"
	"fmt"
	"github.com/QuangTung97/hits"
)

func main() {
	err := hits.Listen(context.Background(), ":5000", func(e hits.MarshalledEvent) {
		fmt.Println(e.Sequence, e.Timestamp, e.Data)
	})
	if err != nil {
		panic(err)
	}
}

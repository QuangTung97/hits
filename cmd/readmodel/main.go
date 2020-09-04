package main

import (
	"context"
	"fmt"

	"github.com/QuangTung97/hits"
	"github.com/QuangTung97/hits/readmodel"
)

type readModel struct {
}

func (rm *readModel) InitFromOriginalDB() uint64 {
	return 0
}

func (rm *readModel) Process(event hits.MarshalledEvent) {
	fmt.Println("PROCESS", event)
}

func main() {
	ctx := context.Background()

	read := &readModel{}
	readmodel.Listen(ctx, "localhost:5000", 5, read)
}

package main

import (
	"context"
	"fmt"

	"github.com/QuangTung97/hits"
	"github.com/QuangTung97/hits/readmodel"
)

type ReadModel struct {
}

func (rm *ReadModel) InitFromOriginalDB() uint64 {
	return 0
}

func (rm *ReadModel) Process(event hits.MarshalledEvent) {
	fmt.Println("PROCESS", event)
}

func main() {
	ctx := context.Background()

	read := &ReadModel{}
	readmodel.Listen(ctx, "localhost:5000", 5, read)
}

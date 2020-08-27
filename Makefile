.PHONY: all test run build

all:
	cd rpc && protoc --go_out=plugins=grpc,paths=source_relative:. hits.proto
	go build -o main cmd/main.go
	go build -o observer cmd/observer/main.go

run:
	go run cmd/main.go

test:
	go test -run=NONE -bench=.

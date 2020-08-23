.PHONY: all test run

all:
	cd rpc && protoc --go_out=plugins=grpc,paths=source_relative:. hits.proto

run:
	go run cmd/main.go

test:
	go test -run=NONE -bench=.

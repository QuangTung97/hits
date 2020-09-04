.PHONY: all test run build

all:
	cd rpc && protoc --go_out=plugins=grpc,paths=source_relative:. hits.proto
	go build -o server cmd/server/main.go
	go build -o observer cmd/observer/main.go
	go build -o read cmd/readmodel/main.go
	go build -o perf cmd/perf/main.go


run:
	go run cmd/perf/main.go

test:
	go test -run=NONE -bench=.

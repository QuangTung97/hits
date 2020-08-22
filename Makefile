.PHONY: all test

all:
	go run cmd/main.go

test:
	go test -run=NONE -bench=.

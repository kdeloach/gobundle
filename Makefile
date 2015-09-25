.PHONY: all build test

all: build test

build:
	go build -o bin/gobundle

test:
	go test gobundle/gobundle


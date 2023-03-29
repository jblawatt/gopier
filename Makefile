#!/bin/env make

export GOPATH=$(shell pwd)/go

echo:
	echo "GOPATH $(GOPATH)"

run: clean
	go run main.go create --src test-template/src --dest test-template/dest -v test-template/values.yaml

clean:
	rm -rf test-template/dest

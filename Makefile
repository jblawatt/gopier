#!/bin/env make

export GOPATH=$(shell pwd)/go

echo:
	echo "GOPATH $(GOPATH)"

run: clean
	go run main.go create --src test-template/src --dest test-template/dest -v test-values.yaml

run-git: clean
	go run main.go create --src "git+https://gitea.h.j3nko.de/j3nko/gopier-test" --dest test-template/dest -v test-template/src/values.yaml

clean:
	rm -rf test-template/dest

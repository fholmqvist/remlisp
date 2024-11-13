SHELL := /bin/bash

BUILD-TARGET = /usr/local/bin
COVER = cover.out
HTML  = cover.html

.PHONY: build run run-silent repl install test cover cover-html clean

run:
	go run . --debug

run-silent:
	go run .

repl:
	go run . --repl

install:
	go build -o rem .
	sudo mv rem $(BUILD-TARGET)

test:
	go test ./...

clean:
	rm -rf $(COVER) $(HTML) out

cover:
	go test ./... -coverprofile=$(COVER)
	go tool cover -func $(COVER) | grep total: | awk '{print $$3}'

cover-html: clean cover
	go tool cover -html=$(COVER) -o $(HTML)
	xdg-open $(HTML)

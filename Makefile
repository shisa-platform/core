# -*- mode: Makefile-gmake -*-

SHELL := bash

TOP_DIR := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))

DIAGRAM_DIR := doc/diagram
DIAGRAMS := $(DIAGRAM_DIR)/architecture.png

all: test

doc/diagram/%.png: doc/%.dot
	@mkdir -p $(DIAGRAM_DIR)
	dot -Tpng $< > $@

doc: $(DIAGRAMS)

clean:
	rm -rf pkg

fmt:
	go fmt ./...

vet:
	go vet ./...

test: fmt vet
	go test -v ./...

.PHONY: clean doc vet fmt

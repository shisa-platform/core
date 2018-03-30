# -*- mode: Makefile-gmake -*-

SHELL := bash

TOP_DIR := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))

DIAGRAM_DIR := doc/diagram
DIAGRAMS := $(DIAGRAM_DIR)/architecture.png

BUILD_DIR := build
COVERAGE_DIR := $(BUILD_DIR)/coverage

SHISA_PKGS := $(shell go list ./... | grep -Ev 'examples|test')
SHISA_TEST_PKGS := $(addprefix coverage/,$(SHISA_PKGS))

all: test

doc/diagram/%.png: doc/%.dot
	@mkdir -p $(DIAGRAM_DIR)
	dot -Tpng $< > $@

doc: $(DIAGRAMS)

$(BUILD_DIR):
	@mkdir -p $@

$(COVERAGE_DIR):
	@mkdir -p $@

clean:
	rm -rf $(BUILD_DIR)

fmt:
	go fmt ./...

vet:
	go vet ./...

gen:
	find . -name '*_charlatan.go' | xargs rm
	go generate ./...

test: ${COVERAGE_DIR} ${SHISA_TEST_PKGS}

examples:
	go build -o $(TOP_DIR)/$(BUILD_DIR)/rest github.com/percolate/shisa/examples/rest
	go build -o $(TOP_DIR)/$(BUILD_DIR)/rpc github.com/percolate/shisa/examples/rpc
	go build -o $(TOP_DIR)/$(BUILD_DIR)/idp github.com/percolate/shisa/examples/idp
	go build -o $(TOP_DIR)/$(BUILD_DIR)/gw github.com/percolate/shisa/examples/gw

docker:
	docker build --tag shisa/examples/base -f examples/Dockerfile.base .
	docker-compose -f examples/docker-compose.yml build

coverage/%:
	go test -v -coverprofile=$(TOP_DIR)/$(COVERAGE_DIR)/$(@F)_coverage.out -covermode=atomic github.com/percolate/shisa/$(@F)

.PHONY: clean doc vet fmt test examples docker

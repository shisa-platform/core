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
	go build -o $(TOP_DIR)/$(BUILD_DIR)/rest github.com/shisa-platform/core/examples/rest
	go build -o $(TOP_DIR)/$(BUILD_DIR)/rpc github.com/shisa-platform/core/examples/rpc
	go build -o $(TOP_DIR)/$(BUILD_DIR)/idp github.com/shisa-platform/core/examples/idp
	go build -o $(TOP_DIR)/$(BUILD_DIR)/gw github.com/shisa-platform/core/examples/gw

docker:
	docker build --tag shisa/examples/gw -f examples/gw/Dockerfile .
	docker build --tag shisa/examples/idp -f examples/idp/Dockerfile .
	docker build --tag shisa/examples/rest -f examples/rest/Dockerfile .
	docker build --tag shisa/examples/rpc -f examples/rpc/Dockerfile .

coverage/%:
	go test -v -coverprofile=$(TOP_DIR)/$(COVERAGE_DIR)/$(@F)_coverage.out -covermode=atomic github.com/shisa-platform/core/$(@F)

t:
	echo $(SHISA_PKGS)

.PHONY: clean doc vet fmt test examples docker

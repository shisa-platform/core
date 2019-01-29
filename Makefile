# -*- mode: Makefile-gmake -*-

SHELL := bash

TOP_DIR := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))

BUILD_DIR := build
COVERAGE_DIR := $(BUILD_DIR)/coverage

DIAGRAMS := doc/diagram/auxiliary.png

SHISA_PKGS := $(shell go list ./... | grep -Ev 'test')
SHISA_TEST_PKGS := $(addprefix coverage/,$(SHISA_PKGS))

all: test

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

coverage/%:
	go test -v -coverprofile=$(TOP_DIR)/$(COVERAGE_DIR)/$(@F)_coverage.out -covermode=atomic github.com/shisa-platform/core/$(@F)

t:
	echo $(SHISA_PKGS)

.PHONY: clean doc vet fmt test

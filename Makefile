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

test: ${BUILD_DIR} ${SHISA_TEST_PKGS}

coverage/%: $(BUILD_DIR) $(COVERAGE_DIR)
	cd $(TOP_DIR)/$(@F) && go test -v -coverprofile=$(TOP_DIR)/$(COVERAGE_DIR)/$(@F)_coverage.out -covermode=atomic

.PHONY: clean doc vet fmt test

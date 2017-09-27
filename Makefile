# -*- mode: Makefile-gmake -*-

SHELL := bash

TOP_DIR := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))

DIAGRAM_DIR := doc/diagram
DIAGRAMS := $(DIAGRAM_DIR)/architecture.png

DIST_DIR := dist
GOBIN_DIR := $(TOP_DIR)/$(DIST_DIR)

VENDOR_PATH := $(TOP_DIR)/vendor

TARGET := api-gw
TARGET_PATH := $(GOBIN_DIR)/$(API_TARGET)

VERSION := $(shell cat VERSION.txt)
MAIN_PORT := 8001
DEBUG_PORT := 8002
PKG_NAME := api-gw_$(API_VERSION)
PKG_DIR := $(DIST_DIR)/$(PKG_NAME)
BUILD_DIR := build/package
CNTRL_FILE := DEBIAN/control
SUPERVISOR_FILE_DIR := etc/supervisor/conf.d
SUPERVISOR_FILE_NAME := api-gw.conf

BUILD_HOST := $(shell hostname -f)
BUILD_DATE := $(shell date +'%s')
BUILD_OS   := $(shell uname -s)
BUILD_VERS := $(shell uname -r)
BUILD_ARCH := $(shell uname -m)
GO_VERSION := $(shell go version | awk '{print $$3}')

BUILD_FLAGS := -ldflags="-X main.version=$(VERSION) -X main.compilerVersion=$(GO_VERSION) -X main.buildTimestamp=$(BUILD_DATE) -X main.buildHostname=$(BUILD_HOST) -X main.buildHostOS=$(BUILD_OS) -X main.buildHostOSVersion=$(BUILD_VERS) -X main.buildHostArch=$(BUILD_ARCH)"

all: test

install:
	build/bin/deploy

doc/diagram/%.png: doc/%.dot
	@mkdir -p $(DIAGRAM_DIR)
	dot -Tpng $< > $@

doc: $(DIAGRAMS)

$(DIST_DIR):
	@mkdir -p $@

clean:
	rm -rf $(DIST_DIR)
	rm -rf pkg
	rm -rf vendor/pkg

fmt:
	gofmt -l -w main.go src

vet:
	go vet ./src/...

test: fmt vet
	GOPATH=$(VENDOR_PATH):$(TOP_DIR) go test -v ./src/...

gw: $(DIST_DIR)
	GOBIN=$(GOBIN_DIR) GOPATH=$(VENDOR_PATH):$(TOP_DIR) go install $(BUILD_FLAGS)

package: gw
	@mkdir -p $(PKG_DIR)/percolate/api-gw/bin
	@mkdir -p $(PKG_DIR)/percolate/etc/api-gw
	@mkdir -p $(PKG_DIR)/$(SUPERVISOR_FILE_DIR)
	@cp $(DIST_DIR)/$(TARGET) $(PKG_DIR)/percolate/api-gw/bin
	@cp perc-service.yml $(PKG_DIR)/percolate/etc/api-gw
	@mkdir -p $(PKG_DIR)/DEBIAN
	@VERSION=$(VERSION) envsubst < $(BUILD_DIR)/$(CNTRL_FILE) > $(PKG_DIR)/$(CNTRL_FILE)
	@PORT=$(MAIN_PORT) DEBUGPORT=$(DEBUG_PORT) AUTOSTART=true envsubst < $(SUPERVISOR_FILE_DIR)/$(SUPERVISOR_FILE_NAME) > $(PKG_DIR)/$(SUPERVISOR_FILE_DIR)/$(SUPERVISOR_FILE_NAME)
	@cd $(DIST_DIR) && dpkg-deb --build $(PKG_NAME)

.PHONY: gw clean doc install package vet fmt

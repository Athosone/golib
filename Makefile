GOCOV = $(shell pwd)/bin/gocov
GOCOVXML = $(shell pwd)/bin/gocov-xml
MOCKGEN = $(shell pwd)/bin/mockgen
GINKGO = $(shell pwd)/bin/ginkgo

TEST_OUTPUT="./test_output"

default: build test

test: gotesting
	mkdir -p ${TEST_OUTPUT} || exit 1
	scripts/test $(GINKGO) ${TEST_OUTPUT} unit || exit 1
	$(GOCOV) convert ${TEST_OUTPUT}/*coverage.out | $(GOCOVXML) > ${TEST_OUTPUT}/libraries_coverage.xml

build:
	scripts/build

gotesting:
	scripts/go-get-tool $(GINKGO) github.com/onsi/ginkgo/v2/ginkgo
	scripts/go-get-tool $(GOCOV) github.com/axw/gocov/gocov
	scripts/go-get-tool $(GOCOVXML) github.com/AlekSi/gocov-xml
	scripts/go-get-tool $(MOCKGEN) github.com/golang/mock/mockgen

TARGETS := $(shell ls scripts)

.PHONY: $(TARGETS)

help:
	@echo "Available targets:"
	@echo "	$(TARGETS)"

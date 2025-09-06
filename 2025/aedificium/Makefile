# Purge built-in suffix-based inference-rules to speed up builds.
.SUFFIXES:

GO_DIR := /opt/go
DELETER := rm -f

LIBRARY_OF_LAMBDAS_SRCS := $(wildcard lol/*.go)

APP := explorer
export GOBIN := $(realpath $(dir $(APP)))

.PHONY: format clean

$(APP): $(LIBRARY_OF_LAMBDAS_SRCS) main.go
	$(GO_DIR)/bin/go build -o $(APP) -i .

format:
	$(GO_DIR)/bin/gofmt -w .

clean:
	$(DELETER) $(APP)

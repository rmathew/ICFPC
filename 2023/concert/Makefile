GO_DIR := /opt/go
DELETER := rm -f

PLAN_SRCS := $(wildcard plan/*.go)

APP := placer
export GOBIN := $(realpath $(dir $(APP)))

.PHONY: format clean

$(APP): $(PLAN_SRCS) main.go
	$(GO_DIR)/bin/go build -o $(APP) -i .

format:
	$(GO_DIR)/bin/gofmt -w .

clean:
	$(DELETER) $(APP)

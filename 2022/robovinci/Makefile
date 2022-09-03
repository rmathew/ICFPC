GO_DIR := /opt/go
DELETER := rm -f

PAINTER_SRCS := $(wildcard painter/*.go)

APP := lenny
export GOBIN := $(realpath $(dir $(APP)))

.PHONY: format clean

$(APP): $(PAINTER_SRCS) main.go
	$(GO_DIR)/bin/go build -o $(APP) -i .

format:
	$(GO_DIR)/bin/gofmt -w .

clean:
	$(DELETER) $(APP)

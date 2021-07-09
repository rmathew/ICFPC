GO_DIR = /usr/local/go
DEL = rm -f

SQUEEZE_SRCS = $(wildcard squeeze/*.go)

APP = poses
export GOBIN = $(realpath $(dir $(APP)))

.PHONY: fmt run clean

$(APP): $(SQUEEZE_SRCS) main.go
	$(GO_DIR)/bin/go build -o $(APP) -i .

run: $(APP)
	$(APP) /extra1/ICFPC21_Probs/1.problem

fmt:
	$(GO_DIR)/bin/gofmt -w .

clean: fmt
	$(DEL) $(APP)

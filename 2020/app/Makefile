# A path to a file containing the interaction-protocol to be run. The default
# is to run "galaxy.txt" obtained from one of:
#   https://github.com/icfpcontest2020/galaxy/tree/master/galaxy
#   https://message-from-space.readthedocs.io/en/latest/message42.html
IP_FILE = /extra1/ICFPC20_Probs/galaxy.txt

GO_DIR = /usr/local/go
DEL = rm -f

GALAXY_SRCS = $(wildcard galaxy/*.go)
RUNNER_SRCS = $(wildcard runner/*.go)

GALAXY_PAD = galaxy_pad
export GOBIN = $(realpath $(dir $(GALAXY_PAD)))

.PHONY: fmt run test clean

$(GALAXY_PAD): $(GALAXY_SRCS) $(RUNNER_SRCS)
	$(GO_DIR)/bin/go build -o $(GALAXY_PAD) -i ./runner

fmt:
	$(GO_DIR)/bin/gofmt -w .

run: $(GALAXY_PAD)
	$(GALAXY_PAD) $(IP_FILE)

test: $(GALAXY_PAD)
	$(GO_DIR)/bin/go test ./galaxy

clean: fmt
	$(DEL) $(GALAXY_PAD)

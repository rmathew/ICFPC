# A path to a file containing the interaction-protocol to be run. The default
# is to run "galaxy.txt" obtained from:
#     https://message-from-space.readthedocs.io/en/latest/message42.html
IP_FILE = /extra1/ICFPC20_Probs/galaxy.txt

# A file containing the API-key assigned by the organizers to the team.
API_KEY_FILE = $(HOME)/stuff/icfpc20.key

GO_DIR = /usr/local/go
DEL = rm -f

GALAXY_SRCS = $(wildcard galaxy/*.go)
RUNNER_SRCS = $(wildcard runner/*.go)

GALAXY_PAD = galaxy_pad
export GOBIN = $(realpath $(dir $(GALAXY_PAD)))

.PHONY: fmt run clean

$(GALAXY_PAD): $(GALAXY_SRCS) $(RUNNER_SRCS)
	$(GO_DIR)/bin/go build -o $(GALAXY_PAD) -i ./runner

fmt:
	$(GO_DIR)/bin/gofmt -w .

run: $(GALAXY_PAD)
	$(GALAXY_PAD) --api_key_file $(API_KEY_FILE) $(IP_FILE)

clean: fmt
	$(DEL) $(GALAXY_PAD)

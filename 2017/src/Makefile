# Simple Makefile to help run Punters and local servers for ICFPC 2017.
PUNTER_ID = 1
REMOTE_SERVER = punter.inf.ed.ac.uk
REMOTE_PORT = 9998
LOCAL_SERVER = localhost
LOCAL_PORT = 12345
WORLD_MAP = ./maps/sample.json
LAMDUCT = /extra1/lambda-duct/lamduct
LAMDUCT_LOG_LEVEL = 2
DEL = rm -f

PYC_FILES = $(patsubst %.py,%.pyc,$(wildcard *.py))
LOG_FILES = $(patsubst punter%,messages%.log,$(wildcard punter?))

.PHONY: remote_punter local_punter local_server clean

remote_punter:
	$(DEL) ./messages$(PUNTER_ID).log
	$(LAMDUCT) --client-instance-logfile=./messages$(PUNTER_ID).log \
	    --game-hostname=$(REMOTE_SERVER) --game-port=$(REMOTE_PORT) \
	    --log-level=$(LAMDUCT_LOG_LEVEL) ./punter$(PUNTER_ID)

local_punter:
	$(DEL) ./messages$(PUNTER_ID).log
	$(LAMDUCT) --client-instance-logfile=./messages$(PUNTER_ID).log \
	    --game-hostname=$(LOCAL_SERVER) --game-port=$(LOCAL_PORT) \
	    --log-level=$(LAMDUCT_LOG_LEVEL) ./punter$(PUNTER_ID)

local_server:
	python2 ./local_server.py $(WORLD_MAP)

clean:
	$(DEL) $(PYC_FILES)
	$(DEL) $(LOG_FILES)

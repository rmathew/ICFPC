DEL = rm -f
PUNTER1_LOG_FILE = punter1.log
PUNTER2_LOG_FILE = punter2.log
LAMBDA_DUCT = /extra1/lambda-duct/lamduct

.PHONY: clean punter1 punter2 server

PYC_FILES = $(patsubst %.py,%.pyc,$(wildcard *.py))

punter1:
	$(LAMBDA_DUCT) --client-instance-logfile=$(PUNTER1_LOG_FILE) \
	    --game-hostname=localhost --game-port=12345 --log-level=3 \
	    ./punter.py

server:
	./local_server.py sample.json

clean:
	$(DEL) $(PUNTER1_LOG_FILE)
	$(DEL) $(PUNTER2_LOG_FILE)
	$(DEL) $(PYC_FILES)

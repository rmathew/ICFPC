#!/usr/bin/env python2
import fcntl
import os
import sys
import utils
import lambda_punter

def _make_std_streams_block():
    for stream in [sys.stdin, sys.stdout, sys.stderr]:
        fd = stream.fileno()
        fl = fcntl.fcntl(fd, fcntl.F_GETFL)
        if fl == -1:
            utils.eprint("ERROR getting file-status of FD %d." % fd)
            continue
        if (fl & os.O_NONBLOCK) == os.O_NONBLOCK:
            utils.eprint("FD %d is non-blocking - making it blocking." % fd)
            fcntl.fcntl(fd, fcntl.F_SETFL, (fl & ~os.O_NONBLOCK) & ~os.O_ASYNC)

def run():
    _make_std_streams_block()

    punter_strategy = "lurk"
    if len(sys.argv) > 1:
        punter_strategy = sys.argv[1]

    if punter_strategy == "lurk":
        punter = lambda_punter.LambdaPunter("codermal_lurker")
    elif punter_strategy == "naive":
        punter = lambda_punter.NaivePunter("codermal_naive")
    else:
        utils.eprint("ERROR: Unknown Punter-strategy '%s'." % punter_strategy)
        return

    punter.punt()

if __name__ == "__main__":
    run()

#!/usr/bin/env python2
from __future__ import print_function

import fcntl
import json
import os
import sys
import utils

def make_streams_block():
    for stream in [sys.stdin, sys.stdout, sys.stderr]:
        fd = stream.fileno()
        fl = fcntl.fcntl(fd, fcntl.F_GETFL)
        if fl == -1:
            utils.eprint("ERROR getting file-status of FD %d." % fd)
            continue
        if (fl & os.O_NONBLOCK) == os.O_NONBLOCK:
            utils.eprint("FD %d is non-blocking - making it blocking." % fd)
            fcntl.fcntl(fd, fcntl.F_SETFL, (fl & ~os.O_NONBLOCK) & ~os.O_ASYNC)

def send_msg(obj):
    utils.encode_obj(sys.stdout, obj)

def recv_msg():
    return utils.decode_obj(sys.stdin)

def run():
    make_streams_block()
    send_msg({"me": "codermal"})
    utils.eprint(recv_msg())

if __name__ == "__main__":
    run()

#!/usr/bin/env python3
from datetime import datetime
from urllib import request

import code
import readline
import sys

class Repl(code.InteractiveConsole):
    SERVER = "https://boundvariable.space/communicate"
    IDX_DELTA = 33
    DEC_KEY = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!\"#$%&'()*+,-./:;<=>?@[\]^_`|~ \n"

    def __init__(self, key, log_file):
        code.InteractiveConsole.__init__(self)
        self.auth_key = key
        self.log_file = log_file
        self.enc_key = {}
        for i in range(94):
            self.enc_key[self.DEC_KEY[i]] = chr(i + self.IDX_DELTA)

    def encode(self, msg):
        out = "S"
        for c in msg:
            out += self.enc_key[c]
        return out

    def decode(self, msg):
        if len(msg) == 0 or msg[0] != 'S':
            sys.exit("Invalid encoded response: " + msg)

        out = ""
        for c in msg[1:]:
            out += self.DEC_KEY[ord(c) - self.IDX_DELTA]
        return out

    def runsource(self, source, filename="<input>", symbol="single"):
        self.log_file.write(">>>\n" + source + "\n")
        enc_msg = self.encode(source)
        hdrs = {"Authorization": self.auth_key}
        req = request.Request(self.SERVER, data=enc_msg.encode(), headers=hdrs,
                              method="POST")
        res = request.urlopen(req)
        dec_res = self.decode(res.read().decode('utf-8'))
        print(dec_res)
        self.log_file.write("<<<\n" + dec_res + "\n")

def main():
    if len(sys.argv) != 2:
        sys.exit("No auth-key")

    curr_dt = datetime.now().strftime("%Y-%m-%d-%H-%M-%S")
    with open("session-" + curr_dt + ".log", "a") as log_file:
        repl = Repl(sys.argv[1], log_file)
        repl.interact(banner="ICFPC 2024 REPL.", exitmsg="Bye!")

if __name__ == "__main__":
    main()

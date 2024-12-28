#!/usr/bin/env python3
from datetime import datetime
from urllib import request

import argparse
import code
import icfp_lang
import logging
import readline

SERVER = "https://boundvariable.space/communicate"


class MacrowareInsightRepl(code.InteractiveConsole):

    def __init__(self, key):
        code.InteractiveConsole.__init__(self)
        self.auth_key = key
        self.codec = icfp_lang.IcfpCodec()

    def runsource(self, source, filename="<input>", symbol="single"):
        logging.info(f"PLAIN >>> {source}\n")
        enc_req = self.codec.encode(source)
        logging.debug(f"CODED >>> {enc_req}\n")
        hdrs = {"Authorization": self.auth_key}
        svr_req = request.Request(SERVER, data=enc_req.encode(), headers=hdrs,
                                  method="POST")
        svr_res = request.urlopen(svr_req)
        enc_res = svr_res.read().decode('utf-8')
        logging.debug(f"CODED <<<\n{enc_res}\n")
        try:
            val = self.codec.decode(enc_res).eval()
            dec_res = "".join(val) if isinstance(val, list) else str(val)
            print(dec_res)
            logging.info(f"PLAIN <<<\n{dec_res}\n")
        except SyntaxError as syntax_err:
            err_msg = str(syntax_err)
            print(f"ERROR: {err_msg}\n{enc_res}")
            logging.error(f"{err_msg}\n{enc_res}\n")


def main():
    arg_parser = argparse.ArgumentParser(
        prog="MacrowareInsight",
        description="A simple REPL for chatting with the ICFP 2024 server.")
    arg_parser.add_argument("-l", "--log",
                            choices=["debug", "info", "warning", "error"],
                            default="info")
    arg_parser.add_argument("key", help="Authorization key for the team")
    args = arg_parser.parse_args()

    log_level = getattr(logging, args.log.upper(), None)
    logging.basicConfig(filename="icfpc24_sessions.log",
                        format="%(asctime)s %(levelname)s: %(message)s",
                        datefmt="%Y-%m-%d %H:%M:%S", level=log_level)

    repl = MacrowareInsightRepl(args.key)
    repl.interact(banner="Macroware Insight (ICFPC 2024).", exitmsg="Bye!")


if __name__ == "__main__":
    main()

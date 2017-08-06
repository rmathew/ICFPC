from __future__ import print_function

import json
import sys

def eprint(*args, **kwargs):
    print(*args, file=sys.stderr, **kwargs)
    sys.stderr.flush()

def encode_obj(stream, obj):
    msg = json.dumps(obj, ensure_ascii=True, separators=(',', ':'))
    stream.write("%d:%s" % (len(msg), msg))
    stream.flush()

def decode_obj(stream):
    num_str = ""
    char = stream.read(1)
    while char != ":":
        num_str += char
        char = stream.read(1)
    num = int(num_str)
    msg = stream.read(num)
    obj = json.loads(msg)
    return obj

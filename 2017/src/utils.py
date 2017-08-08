from __future__ import print_function

import json
import sys

def eprint(*args, **kwargs):
    print(*args, file=sys.stderr, **kwargs)
    sys.stderr.flush()

def encode_obj(stream, obj):
    message = json.dumps(obj, ensure_ascii=True, separators=(',', ':'))
    stream.write("%d:" % len(message))
    stream.write(message)
    stream.flush()

def decode_obj(stream):
    num_expected_str = ""
    char = stream.read(1)
    while char != ":":
        num_expected_str += char
        char = stream.read(1)
    num_expected = int(num_expected_str)
    message = ""
    while num_expected > 0:
        fragment = stream.read(num_expected)
        message += fragment
        num_expected -= len(fragment)
    return json.loads(message)

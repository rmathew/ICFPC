#!/usr/bin/env python2
from __future__ import print_function

import asyncore
import json
import socket
import sys
import utils
import lambda_world

class PunterHandler(asyncore.dispatcher_with_send):

    def __init__(self, world_state, sock):
        asyncore.dispatcher_with_send.__init__(self, sock=sock)
        self.world_state = world_state

    def read(self, buffer_size):
        return self.recv(buffer_size)

    def write(self, data):
        self.send(data)

    def flush(self):
        pass

    def handle_read(self):
        msg_obj = utils.decode_obj(self)
        if "me" in msg_obj:
            punter_name = msg_obj["me"]
            utils.eprint("INFO: Punter calls itself '%s'." % punter_name)
            punter_id = self.world_state.add_punter()
            utils.encode_obj(self, {"you": punter_name})
            utils.encode_obj(self, {"punter": punter_id,
                "punters": self.world_state.num_punters,
                "map": self.world_state.world_map.to_dict()})
        elif "ready" in msg_obj:
            pass
        else:
            utils.eprint("WARNING: Ignored message.")

class WorldServer(asyncore.dispatcher):

    def __init__(self, world_state):
        asyncore.dispatcher.__init__(self)
        self.world_state = world_state
        self.create_socket(socket.AF_INET, socket.SOCK_STREAM)
        self.set_reuse_addr()
        self.bind(("localhost", 12345))
        self.listen(3)

    def handle_accept(self):
        pair = self.accept()
        if pair is not None:
            sock, addr = pair
            utils.eprint("INFO: Punter connected.")
            handler = PunterHandler(self.world_state, sock)

def run():
    if len(sys.argv) != 2:
        utils.eprint("ERROR: Incorrect number of arguments.")
        return

    world_map_file = sys.argv[1]
    with open(world_map_file, "r") as f:
        world_map_str = f.read()
    world_state_dict = {}
    world_state_dict["map"] = json.loads(world_map_str)
    utils.eprint("INFO: World map loaded from '%s'." % world_map_file)
    world_state_dict["punters"] = 0
    world_state_dict["claims"] = []
    world_state = lambda_world.WorldState(world_state_dict)
    world_map = world_state.world_map
    utils.eprint("INFO: World map has %d sites, %d rivers and %d mines." %
        (world_map.get_num_sites(), world_map.get_num_rivers(),
            world_map.get_num_mines()))

    server = WorldServer(world_state)
    asyncore.loop()

if __name__ == "__main__":
    run()

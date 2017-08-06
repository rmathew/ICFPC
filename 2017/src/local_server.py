#!/usr/bin/env python2
from __future__ import print_function

import json
import socket
import sys
import utils
import lambda_world

class PunterComm():
    def __init__(self, sock, world_state):
        sock.setblocking(1)
        self.sock = sock
        self.world_state = world_state
        self.punter_id = lambda_world.UNKNOWN_PUNTER_ID
        self.claimed_river = lambda_world.INVALID_RIVER

    def shake_hands(self):
        msg_dict = utils.decode_obj(self)
        punter_name = msg_dict["me"]
        self.punter_id = self.world_state.add_punter()
        utils.eprint("INFO: Punter %d calls itself '%s'." %
            (self.punter_id, punter_name))
        utils.encode_obj(self, {"you": punter_name})

    def set_up(self):
        utils.encode_obj(self, {"punter": self.punter_id,
            "punters": self.world_state.num_punters,
            "map": self.world_state.world_map.to_dict()})
        msg_dict = utils.decode_obj(self)
        ack_punter_id = msg_dict.get("ready", lambda_world.UNKNOWN_PUNTER_ID)
        if ack_punter_id == self.punter_id:
            utils.eprint("INFO: Punter %d is ready." % self.punter_id)

    def make_a_move(self, turn_num, move_cmd_dict):
        self.claimed_river = lambda_world.INVALID_RIVER
        utils.encode_obj(self, move_cmd_dict)
        new_move_dict = utils.decode_obj(self)
        if not "claim" in new_move_dict:
            utils.eprint("INFO: [Turn %d] Punter %d passed on its turn." %
                (turn_num, self.punter_id))
            return
        river_dict = new_move_dict["claim"]
        claimed_river = lambda_world.River(int(river_dict["source"]),
            int(river_dict["target"]))
        is_valid_claim = self.world_state.add_punter_claim(self.punter_id,
            claimed_river)
        if is_valid_claim:
            self.claimed_river = claimed_river
            utils.eprint("INFO: [Turn %d] Punter %d claimed river %s." %
                (turn_num, self.punter_id, str(claimed_river)))
        else:
            utils.eprint(
                "WARNING: [Turn %d] Punter %d could not claim river %s." %
                (turn_num, self.punter_id, str(claimed_river)))

    def wrap_it_up(self, stop_cmd_dict):
        utils.encode_obj(self, stop_cmd_dict)
        utils.eprint("INFO: Disconnecting Punter %d." % self.punter_id)
        self.sock.close()

    def get_score(self):
        score = self.world_state.calculate_score(self.punter_id)
        utils.eprint("INFO: Punter %d scored %d." % (self.punter_id, score))
        return score

    def read(self, buffer_size):
        return self.sock.recv(buffer_size)

    def write(self, data):
        self.sock.send(data)

    def flush(self):
        pass

def wait_for_punter(server_socket, world_state):
    client_socket, _ = server_socket.accept()
    punter_comm = PunterComm(client_socket, world_state)
    punter_comm.shake_hands()
    return punter_comm

def get_prev_moves_list(punter_comms_list):
    prev_moves_list = []
    for a_punter_comm in punter_comms_list:
        move_dict = {"punter": a_punter_comm.punter_id}
        claimed_river = a_punter_comm.claimed_river
        if claimed_river == lambda_world.INVALID_RIVER:
            prev_moves_list.append({"pass": {
                "punter": a_punter_comm.punter_id}})
        else:
            prev_moves_list.append({"claim": {
                "punter": a_punter_comm.punter_id,
                "source": claimed_river.source,
                "target": claimed_river.target}})
    return prev_moves_list

def get_scores_list(punter_comms_list):
    scores_list = []
    for a_punter_comm in punter_comms_list:
        scores_list.append({"punter": a_punter_comm.punter_id,
            "score": a_punter_comm.get_score()})
    return scores_list

def play_game(punter_comms_list, world_state):
    for a_punter_comm in punter_comms_list:
        a_punter_comm.set_up()

    for turn_num in range(world_state.world_map.get_num_rivers()):
        move_cmd_dict = {"move": {"moves":
            get_prev_moves_list(punter_comms_list)}}
        for a_punter_comm in punter_comms_list:
            a_punter_comm.make_a_move(turn_num, move_cmd_dict)

    stop_cmd_dict = {"stop": {
        "moves": get_prev_moves_list(punter_comms_list),
        "scores": get_scores_list(punter_comms_list)}}
    for a_punter_comm in punter_comms_list:
        a_punter_comm.wrap_it_up(stop_cmd_dict)

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

    server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    server_socket.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
    server_socket.bind(("localhost", 12345))
    server_socket.listen(2)

    utils.eprint("INFO: Waiting for two Punters.")
    punter1_comm = wait_for_punter(server_socket, world_state)
    utils.eprint("INFO: Punter1 connected. Waiting for Punter2.")
    punter2_comm = wait_for_punter(server_socket, world_state)
    utils.eprint("INFO: Punter2 connected. Starting the game.")
    server_socket.close()

    play_game([punter1_comm, punter2_comm], world_state)

if __name__ == "__main__":
    run()

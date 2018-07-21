#!/usr/bin/env python2
from __future__ import print_function

import atexit
import json
import lambda_world
import math
import sdl2
import sdl2.ext
import sdl2.sdlgfx
import socket
import sys
import time
import utils

INTER_TURN_SLEEP_SECS = 0.3

def _close_socket(sock):
    try:
        sock.shutdown(socket.SHUT_RDWR)
    except:
        pass
    finally:
        sock.close()

class PunterComm():
    def __init__(self, sock, world_state):
        sock.setblocking(1)
        atexit.register(_close_socket, sock)
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

class Visualizer():

    WINDOW_WIDTH = 800
    WINDOW_HEIGHT = 600
    GUTTER_SIZE = 50
    IMAGES = sdl2.ext.Resources(__file__, "images")

    def __init__(self, world_state):
        self.world_state = world_state
        sdl2.ext.init()
        atexit.register(sdl2.ext.quit)
        self.window = sdl2.ext.Window("Lambda Punter",
            size=(Visualizer.WINDOW_WIDTH, Visualizer.WINDOW_HEIGHT))
        self.window.show()
        self.renderer=sdl2.ext.Renderer(self.window)
        # Solarized Dark colors.
        self.background_color = sdl2.ext.Color(0x00, 0x2B, 0x36, 0xFF)  # base03
        self.unclaimed_river_color = (0x83, 0x94, 0x96, 0xFF)  # base0
        self.claimed_river_colors = [
            (0x26, 0x8B, 0xD2, 0xFF),  # blue
            (0xB5, 0x89, 0x00, 0xFF)]  # yellow
        self._set_river_width()
        self._create_sprites()
        self.update_world_map()

    def check_keep_going(self):
        keep_going = True
        events = sdl2.ext.get_events()
        for event in events:
            if event.type == sdl2.SDL_QUIT:
                keep_going = False
                break
            elif event.type == sdl2.SDL_KEYDOWN:
                if event.key.keysym.sym == sdl2.SDLK_ESCAPE:
                    keep_going = False
                    break
        if not keep_going:
            sys.exit(0)
        self.update_world_map()

    def update_world_map(self):
        self.renderer.clear(self.background_color)
        claims_dict = self.world_state.claims_dict
        sites_dict = self.world_state.world_map.sites_dict
        uc_r, uc_g, uc_b, uc_a = self.unclaimed_river_color
        for a_site_id, a_site in sites_dict.items():
            site_x, site_y = self._locate_site_on_screen(a_site)
            for a_neighbor_id in a_site.neighbors_list:
                if a_neighbor_id < a_site_id:
                    continue
                river = lambda_world.River(a_site_id, a_neighbor_id)
                if river in claims_dict:
                    continue
                neighbor_x, neighbor_y = self._locate_site_on_screen(
                    sites_dict[a_neighbor_id])
                sdl2.sdlgfx.thickLineRGBA(self.renderer.sdlrenderer, site_x,
                    site_y, neighbor_x, neighbor_y, self.river_width, uc_r,
                    uc_g, uc_b, uc_a)
        for a_river, a_punter_id in claims_dict.items():
            src_site = sites_dict[a_river.source]
            src_x, src_y = self._locate_site_on_screen(src_site)
            tgt_site = sites_dict[a_river.target]
            tgt_x, tgt_y = self._locate_site_on_screen(tgt_site)
            color_idx = a_punter_id % len(self.claimed_river_colors)
            c_r, c_g, c_b, c_a = self.claimed_river_colors[color_idx]
            sdl2.sdlgfx.thickLineRGBA(self.renderer.sdlrenderer, src_x, src_y,
                tgt_x, tgt_y, self.river_width, c_r, c_g, c_b, c_a)

        self.sprite_renderer.render(self.site_sprites)
        self.renderer.present()

    def _set_river_width(self):
        num_rivers = self.world_state.world_map.get_num_rivers()
        self.river_width = 10 - int(math.log(num_rivers, 2.))
        if self.river_width <= 0:
            self.river_width = 1

    def _create_sprites(self):
        self._calculate_transforms()
        sites_dict = self.world_state.world_map.sites_dict
        mines_set = self.world_state.world_map.mines_set
        self.site_sprites = []
        sprite_factory = sdl2.ext.SpriteFactory(sdl2.ext.TEXTURE,
            renderer=self.renderer)
        have_too_many_sites = len(sites_dict) > 64
        for a_site_id, a_site in sites_dict.items():
            is_a_mine = a_site_id in mines_set
            if not is_a_mine and have_too_many_sites:
                continue
            if is_a_mine:
                sprite_image = "mine16x16.png"
                sprite_depth = 2
            else:
                sprite_image = "site16x16.png"
                sprite_depth = 1
            sprite = sprite_factory.from_image(Visualizer.IMAGES.get_path(
                sprite_image))
            sprite.depth = sprite_depth
            s_w, s_h = sprite.size
            site_x, site_y = self._locate_site_on_screen(a_site)
            sprite.position = site_x - s_w / 2, site_y - s_h / 2
            self.site_sprites.append(sprite)
        self.sprite_renderer = sprite_factory.create_sprite_render_system(
            self.window)

    def _calculate_transforms(self):
        min_world_x = sys.float_info.max
        min_world_y = sys.float_info.max
        max_world_x = -sys.float_info.min
        max_world_y = -sys.float_info.min
        for a_site in self.world_state.world_map.sites_dict.values():
            min_world_x = min(a_site.x, min_world_x)
            min_world_y = min(a_site.y, min_world_y)
            max_world_x = max(a_site.x, max_world_x)
            max_world_y = max(a_site.y, max_world_y)
        world_width = max_world_x - min_world_x
        world_height = max_world_y - min_world_y
        screen_width = Visualizer.WINDOW_WIDTH - 2 * Visualizer.GUTTER_SIZE
        screen_height = Visualizer.WINDOW_HEIGHT - 2 * Visualizer.GUTTER_SIZE
        world_to_screen_scale_x = float(screen_width) / world_width
        world_to_screen_scale_y = float(screen_height) / world_height
        self.world_centroid_x = (min_world_x + max_world_x) / 2.
        self.world_centroid_y = (min_world_y + max_world_y) / 2.
        self.world_to_screen_scale = min(world_to_screen_scale_x,
            world_to_screen_scale_y)

    def _locate_site_on_screen(self, site):
        new_x = (site.x - self.world_centroid_x) * self.world_to_screen_scale
        new_y = (site.y - self.world_centroid_y) * self.world_to_screen_scale
        new_x += Visualizer.WINDOW_WIDTH / 2
        new_y += Visualizer.WINDOW_HEIGHT / 2
        return (int(new_x), int(new_y))

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

def play_game(punter_comms_list, world_state, viz):
    for a_punter_comm in punter_comms_list:
        a_punter_comm.set_up()

    viz.check_keep_going()
    for turn_num in range(world_state.world_map.get_num_rivers()):
        move_cmd_dict = {"move": {"moves":
            get_prev_moves_list(punter_comms_list)}}
        punter_with_turn = turn_num % len(punter_comms_list)
        punter_comms_list[punter_with_turn].make_a_move(turn_num, move_cmd_dict)
        viz.check_keep_going()
        time.sleep(INTER_TURN_SLEEP_SECS)

    stop_cmd_dict = {"stop": {
        "moves": get_prev_moves_list(punter_comms_list),
        "scores": get_scores_list(punter_comms_list)}}
    for a_punter_comm in punter_comms_list:
        a_punter_comm.wrap_it_up(stop_cmd_dict)
    viz.check_keep_going()

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

    viz = Visualizer(world_state)
    server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    server_socket.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
    atexit.register(_close_socket, server_socket)
    server_socket.bind(("localhost", 12345))
    server_socket.listen(2)

    utils.eprint("INFO: Waiting for two Punters.")
    punter1_comm = wait_for_punter(server_socket, world_state)
    utils.eprint("INFO: Punter1 connected. Waiting for Punter2.")
    viz.check_keep_going()
    punter2_comm = wait_for_punter(server_socket, world_state)
    utils.eprint("INFO: Punter2 connected. Starting the game.")
    viz.check_keep_going()

    play_game([punter1_comm, punter2_comm], world_state, viz)

    while True:
        time.sleep(INTER_TURN_SLEEP_SECS)
        viz.check_keep_going()

if __name__ == "__main__":
    run()

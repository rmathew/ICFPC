import cPickle
import lambda_world
import random
import sys
import utils

class LambdaPunter():

    def __init__(self, name):
        self.punter_name = name
        self.punter_id = lambda_world.UNKNOWN_PUNTER_ID
        self.world_state = None

    def get_river_to_claim(self):
        return lambda_world.INVALID_RIVER

    def shake_hands(self):
        self.send_msg({"me": self.punter_name})
        response_dict = self.recv_msg()
        if response_dict.get("you", "<<UNKNOWN>>") != self.punter_name:
            utils.eprint("WARNING: Botched hand-shake.")

    def set_up(self, cmd_dict):
        utils.eprint("INFO: Hello, my name is '%s'." % self.punter_name)
        self.punter_id = int(cmd_dict["punter"])
        num_punters = int(cmd_dict["punters"])
        utils.eprint("INFO: I am the Punter with id %d among %d Punters." %
            (self.punter_id, num_punters))
        world_state_dict = {}
        world_state_dict["map"] = cmd_dict["map"]
        world_state_dict["punters"] = num_punters
        world_state_dict["claims"] = []
        self.world_state = lambda_world.WorldState(world_state_dict)
        world_map = self.world_state.world_map
        utils.eprint("INFO: World map has %d sites, %d rivers and %d mines." %
            (world_map.get_num_sites(), world_map.get_num_rivers(),
                world_map.get_num_mines()))
        self.send_msg({"ready": self.punter_id, "state": self.get_state_dict()})

    def make_a_move(self, cmd_dict):
        self.update_state(cmd_dict["state"], cmd_dict["move"]["moves"])

        claimed_river = self.get_river_to_claim()
        if claimed_river != lambda_world.INVALID_RIVER:
            is_valid_claim = self.world_state.add_punter_claim(
                self.punter_id, claimed_river)
            if not is_valid_claim:
                claimed_river = lambda_world.INVALID_RIVER
                utils.eprint("ERROR: Made invalid move %s." % move_str)
        utils.eprint("INFO: My move is %s." %
            self.claim_to_str(self.punter_id, claimed_river))

        response_dict = {"state": self.get_state_dict()}
        if claimed_river == lambda_world.INVALID_RIVER:
            response_dict["pass"] = {"punter": self.punter_id}
        else:
            response_dict["claim"] = {"punter": self.punter_id,
                "source": claimed_river.source, "target": claimed_river.target}
        self.send_msg(response_dict)

    def wrap_it_up(self, cmd_dict):
        self.update_state(cmd_dict["state"], cmd_dict["stop"]["moves"])

        scores_list = cmd_dict["stop"]["scores"]
        max_score = 0
        for a_punters_score_dict in scores_list:
            a_score = int(a_punters_score_dict["score"])
            max_score = max(max_score, a_score)
        my_score = -1
        final_scores_str = ""
        for a_punters_score_dict in scores_list:
            a_punter_id = int(a_punters_score_dict["punter"])
            a_score = int(a_punters_score_dict["score"])
            if a_punter_id == self.punter_id:
                my_score = a_score
                final_scores_str += "*"
            final_scores_str += "%d=%d" % (a_punter_id, a_score)
            if a_score == max_score:
                final_scores_str += "^ "
            else:
                final_scores_str += " "
        utils.eprint("INFO: Final scores:\n  %s" % final_scores_str)
        utils.eprint("INFO: The game has ended and my score is %d." % my_score)

    def punt(self):
        self.shake_hands()
        cmd_dict = self.recv_msg()
        if "punter" in cmd_dict:
            self.set_up(cmd_dict)
        elif "move" in cmd_dict:
            self.make_a_move(cmd_dict)
        elif "stop" in cmd_dict:
            self.wrap_it_up(cmd_dict)
        else:
            utils.eprint("ERROR: Unknown server-command.")
            utils.eprint(cmd_dict)

    def send_msg(self, obj):
        # utils.eprint("DEBUG: send_msg\n%s" % str(obj))
        utils.encode_obj(sys.stdout, obj)

    def recv_msg(self):
        recv_obj = utils.decode_obj(sys.stdin)
        # utils.eprint("DEBUG: recv_msg\n%s" % str(recv_obj))
        return recv_obj

    def claim_to_str(self, punter_id, river):
        if river == lambda_world.INVALID_RIVER:
            return "%d=P" % punter_id
        return "%d=C%s" % (punter_id, str(river))

    def update_state(self, state_dict, prev_moves_list):
        self.punter_id = int(state_dict["punter"])
        self.world_state = cPickle.loads(
            state_dict["world_state"].encode("ascii"))
        prev_moves_str = ""
        for a_move_dict in prev_moves_list:
            if "pass" in a_move_dict:
                punter_id = int(a_move_dict["pass"]["punter"])
                claimed_river = lambda_world.INVALID_RIVER
            else:
                claim_dict = a_move_dict["claim"]
                punter_id = int(claim_dict["punter"])
                source = int(claim_dict["source"])
                target = int(claim_dict["target"])
                claimed_river = lambda_world.River(source, target)
                if punter_id != self.punter_id:
                    is_valid_claim = self.world_state.add_punter_claim(
                        punter_id, claimed_river)
                    if not is_valid_claim:
                        claimed_river = lambda_world.INVALID_RIVER
                        utils.eprint("WARNING: Got invalid move %s." % move_str)
            if punter_id == self.punter_id:
                prev_moves_str += "*"
            prev_moves_str += self.claim_to_str(punter_id, claimed_river)
            prev_moves_str += " "
        utils.eprint("INFO: Previous moves:\n  %s" % prev_moves_str)

    def get_state_dict(self):
        return {"punter": int(self.punter_id),
            "world_state": cPickle.dumps(self.world_state, 0)}

class NaivePunter(LambdaPunter):

    def __init__(self, name):
        LambdaPunter.__init__(self, name)

    def get_river_to_claim(self):
        world_map = self.world_state.world_map
        mines_list = random.sample(world_map.mines_set,
            len(world_map.mines_set))
        for a_mine_id in mines_list:
            river = self.pick_random_free_river(a_mine_id,
                world_map.sites_dict[a_mine_id])
            if river != lambda_world.INVALID_RIVER:
                return river
        for a_site_id, a_site in world_map.sites_dict.items():
            if a_site_id in world_map.mines_set:
                continue
            river = self.pick_random_free_river(a_site_id, a_site)
            if river != lambda_world.INVALID_RIVER:
                return river
        return lambda_world.INVALID_RIVER

    def pick_random_free_river(self, site_id, site):
        neighbors_list = random.sample(site.neighbors_list,
            len(site.neighbors_list))
        for a_neighbor_id in neighbors_list:
            river = lambda_world.River(site_id, a_neighbor_id)
            claimant = self.world_state.get_claiming_punter(river)
            if claimant == lambda_world.UNKNOWN_PUNTER_ID:
                return river
        return lambda_world.INVALID_RIVER
